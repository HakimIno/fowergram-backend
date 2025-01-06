package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fowergram/config"
	"fowergram/internal/core/services"
	"fowergram/internal/handlers"
	"fowergram/internal/repositories/postgres"
	"fowergram/internal/repositories/redis"
	"fowergram/pkg/email"
	"fowergram/pkg/geolocation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

	// Setup Fiber app with custom config
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

	// Health routes must be registered first
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
			"services": fiber.Map{
				"api":   "up",
				"db":    "up",
				"redis": "up",
			},
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)

	// User routes
	users := api.Group("/users")
	users.Get("/:id", userHandler.GetUser)
	users.Get("/", userHandler.GetUsers)

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
