package integration

import (
	"fowergram/config"
	"fowergram/internal/core/domain"
	"fowergram/internal/core/services"
	"fowergram/internal/handlers"
	"fowergram/internal/middleware"
	"fowergram/internal/repositories/postgres"
	"fowergram/pkg/email"
	"fowergram/pkg/geolocation"

	"time"

	"strings"

	"fmt"

	"github.com/gofiber/fiber/v2"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestApp() *fiber.App {
	// Load test config with correct database connection
	cfg := &config.Config{
		DB: setupTestDB(),
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
		Email: config.EmailConfig{
			APIKey:      "test-key",
			SenderEmail: "test@example.com",
			SenderName:  "Test",
		},
		Geo: config.GeoConfig{
			APIKey: "test-key",
		},
	}

	// Initialize test database
	db := setupTestDB()
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.DeviceSession{},
		&domain.LoginHistory{},
		&domain.AuthCode{},
		&domain.AccountRecovery{},
	); err != nil {
		panic(err)
	}

	// Initialize repositories
	authRepo := postgres.NewAuthRepository(db)
	emailService := email.NewEmailService(cfg.Email.APIKey, cfg.Email.SenderEmail, cfg.Email.SenderName)
	geoService := geolocation.NewGeoService(cfg.Geo.APIKey)

	// Initialize services
	authService := services.NewAuthService(authRepo, emailService, geoService, cfg.JWT.Secret)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize middleware
	securityMiddleware := middleware.NewSecurityMiddleware()

	// Initialize app with increased timeout
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// Apply middleware
	app.Use(securityMiddleware.RateLimiter())

	// Setup routes
	api := app.Group("/api")
	v1 := api.Group("/v1")
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected routes
	users := v1.Group("/users")
	users.Use(func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		user, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		c.Locals("user", user)
		return c.Next()
	})
	users.Get("/me", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*domain.User)
		return c.JSON(user)
	})

	return app
}

func setupTestDB() *gorm.DB {
	// Connection parameters
	const (
		host     = "postgres"
		user     = "postgres"
		password = "postgres"
		sslmode  = "disable"
	)

	// เชื่อมต่อกับ postgres โดยไม่ระบุ database
	dsn := fmt.Sprintf("host=%s user=%s password=%s sslmode=%s",
		host, user, password, sslmode)
	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Kill all connections to the test database
	db.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'fowergram_test'")

	// สร้าง database ถ้ายังไม่มี
	db.Exec("DROP DATABASE IF EXISTS fowergram_test")
	db.Exec("CREATE DATABASE fowergram_test")

	// เชื่อมต่อกับ database ที่สร้างใหม่
	dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=fowergram_test sslmode=%s",
		host, user, password, sslmode)
	testDB, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return testDB
}
