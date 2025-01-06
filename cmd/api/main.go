package main

import (
	"log"
	"time"

	"fowergram/config"
	"fowergram/internal/core/services"
	"fowergram/internal/handlers"
	"fowergram/internal/jobs"
	"fowergram/internal/middleware"
	"fowergram/internal/repositories/postgres"
	"fowergram/internal/repositories/redis"
	"fowergram/pkg/email"
	"fowergram/pkg/geolocation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Initialize repositories
	userRepo := postgres.NewUserRepository(cfg.DB)
	postRepo := postgres.NewPostRepository(cfg.DB)
	cacheRepo := redis.NewCacheRepository(cfg.Redis)

	// Initialize services
	emailService := email.NewEmailService(
		cfg.Email.APIKey,
		cfg.Email.SenderEmail,
		cfg.Email.SenderName,
	)
	geoService := geolocation.NewGeoService(cfg.Geo.APIKey)
	authRepo := postgres.NewAuthRepository(cfg.DB)

	// Initialize services
	userService := services.NewUserService(userRepo, cacheRepo)
	postService := services.NewPostService(postRepo, cacheRepo)
	authService := services.NewAuthService(
		authRepo,
		emailService,
		geoService,
		cfg.JWT.Secret,
	)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	postHandler := handlers.NewPostHandler(postService)
	authHandler := handlers.NewAuthHandler(authService)

	// Routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Auth routes
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// User routes
	users := v1.Group("/users", middleware.AuthRequired(cfg.JWT.Secret))
	users.Get("/", userHandler.GetUsers)
	users.Get("/:id", userHandler.GetUser)

	// Post routes
	posts := v1.Group("/posts", middleware.AuthRequired(cfg.JWT.Secret))
	posts.Get("/", postHandler.GetPosts)
	posts.Post("/", postHandler.CreatePost)

	// Health check route
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// Metrics route (moved to specific endpoint)
	app.Get("/metrics", middleware.Prometheus())

	// Start background jobs
	go jobs.StartSessionCleanup(cfg.DB)

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := app.Listen(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
