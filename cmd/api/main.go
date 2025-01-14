package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fowergram/config"
	"fowergram/internal/chat/broker/redpanda"
	"fowergram/internal/chat/handler"
	"fowergram/internal/chat/repository/mongodb"
	"fowergram/internal/chat/service"
	"fowergram/internal/core/services"
	"fowergram/internal/handlers"
	"fowergram/internal/repositories/postgres"
	"fowergram/internal/repositories/redis"
	"fowergram/internal/routes"
	"fowergram/pkg/email"
	"fowergram/pkg/geolocation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupMongoDB(cfg config.MongoDBConfig) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client.Database(cfg.Database), nil
}

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup MongoDB
	mongoDB, err := setupMongoDB(cfg.MongoDB)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Setup repositories
	userRepo := postgres.NewUserRepository(cfg.DB)
	authRepo := postgres.NewAuthRepository(cfg.DB)
	cacheRepo := redis.NewCacheRepository(cfg.Redis)

	// Setup services
	emailService := email.NewEmailService(cfg.Email.APIKey, cfg.Email.SenderEmail, cfg.Email.SenderName)
	geoService := geolocation.NewGeoService(cfg.Geo.APIKey)
	userService := services.NewUserService(userRepo, cacheRepo)
	authService := services.NewAuthService(authRepo, emailService, geoService, cacheRepo, cfg.JWT.Secret)

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
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup routes
	routes.SetupHealthRoutes(app)
	api := app.Group("/api/v1")
	routes.SetupAuthRoutes(api, authHandler)
	routes.SetupUserRoutes(api, userHandler)

	// Setup chat dependencies and routes
	redpandaBroker, err := redpanda.NewRedpandaBroker([]string{cfg.Redpanda.Broker})
	if err != nil {
		log.Fatalf("Failed to connect to Redpanda: %v", err)
	}
	defer redpandaBroker.Close()

	chatRepo := mongodb.NewChatRepository(mongoDB)
	wsManager := service.NewWebSocketManager(cfg.Redis)
	chatService := service.NewChatService(chatRepo, wsManager, redpandaBroker)
	chatHandler := handler.NewChatHandler(chatService, wsManager)
	routes.SetupChatRoutes(api, chatHandler)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server is running on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
