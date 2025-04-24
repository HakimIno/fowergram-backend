package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fowergram/config"
	"fowergram/internal/chat/broker/redpanda"
	"fowergram/internal/chat/handler"
	"fowergram/internal/chat/repository/scylladb"
	"fowergram/internal/chat/service"
	"fowergram/internal/core/services"
	"fowergram/internal/handlers"
	"fowergram/internal/middleware"
	"fowergram/internal/repositories/postgres"
	"fowergram/internal/repositories/redis"
	"fowergram/internal/routes"
	"fowergram/pkg/email"
	"fowergram/pkg/geolocation"
	"fowergram/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// สร้าง logger instance
	log := logger.NewLogger(logger.InfoLevel)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", err,
			logger.NewField("error", err.Error()))
	}

	// ตัวอย่างการใช้งาน logger
	log.Info("Server starting",
		logger.NewField("port", cfg.Server.Port),
		logger.NewField("env", os.Getenv("GO_ENV")),
	)

	// Setup repositories
	userRepo := postgres.NewUserRepository(cfg.DB)
	authRepo := postgres.NewAuthRepository(cfg.DB)
	cacheRepo := redis.NewCacheRepository(cfg.Redis)

	// Setup services
	emailService := email.NewEmailService(cfg.Email.APIKey, cfg.Email.SenderEmail, cfg.Email.SenderName)
	geoService := geolocation.NewGeoService(cfg.Geo.APIKey)
	userService := services.NewUserService(userRepo, cacheRepo)
	authService := services.NewAuthServiceWithLogger(authRepo, emailService, geoService, cacheRepo, cfg.JWT.Secret, log)

	// Setup handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           120 * time.Second,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(middleware.RequestMonitoring(log))
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format:     "[${time}] ${status} | ${latency} | ${ip} | ${method} | ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))
	app.Use(cors.New())

	// Setup routes
	routes.SetupHealthRoutes(app)
	api := app.Group("/api/v1")
	routes.SetupAuthRoutes(api, authHandler)
	routes.SetupUserRoutes(api, userHandler, cfg.JWT.Secret)

	// Setup chat dependencies and routes
	redpandaBroker, err := redpanda.NewRedpandaBroker([]string{cfg.Redpanda.Broker})
	if err != nil {
		log.Error("Failed to connect to Redpanda", err,
			logger.NewField("broker", cfg.Redpanda.Broker))
		os.Exit(1)
	}
	defer redpandaBroker.Close()

	// Initialize ScyllaDB schema
	if err := scylladb.InitializeSchema(cfg.ScyllaDB.Hosts, cfg.ScyllaDB.Keyspace); err != nil {
		log.Error("Failed to initialize ScyllaDB schema", err,
			logger.NewField("hosts", cfg.ScyllaDB.Hosts),
			logger.NewField("keyspace", cfg.ScyllaDB.Keyspace))
		os.Exit(1)
	}

	chatRepo, err := scylladb.NewChatRepository(cfg.ScyllaDB.Hosts, cfg.ScyllaDB.Keyspace)
	if err != nil {
		log.Error("Failed to connect to ScyllaDB", err,
			logger.NewField("hosts", cfg.ScyllaDB.Hosts))
		os.Exit(1)
	}
	defer chatRepo.Close()

	// Initialize WebSocket manager
	wsManager := service.NewWebSocketManager()
	chatService := service.NewChatService(chatRepo, wsManager, redpandaBroker)
	chatHandler := handler.NewChatHandler(chatService, wsManager)
	routes.SetupChatRoutes(api, chatHandler, cfg.JWT.Secret)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Info("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Info("Server is running",
		logger.NewField("address", addr),
		logger.NewField("environment", os.Getenv("GO_ENV")),
	)

	if err := app.Listen(addr); err != nil {
		log.Error("Failed to start server", err,
			logger.NewField("address", addr))
		os.Exit(1)
	}
}
