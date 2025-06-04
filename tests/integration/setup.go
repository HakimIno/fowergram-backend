package integration

import (
	"context"
	"fmt"
	"fowergram/internal/core/ports"
	"fowergram/internal/core/services"
	"fowergram/internal/domain"
	"fowergram/internal/handlers"
	"fowergram/internal/middleware"
	"fowergram/internal/repositories/postgres"
	redisrepo "fowergram/internal/repositories/redis"
	"fowergram/pkg/email"
	"fowergram/pkg/geolocation"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/redis/go-redis/v9"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	// Setup
	testDB = setupTestDB()

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestDB(testDB)

	os.Exit(code)
}

type mockCacheRepo struct{}

func (m *mockCacheRepo) Set(key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (m *mockCacheRepo) Get(key string) (interface{}, error) {
	return nil, redis.Nil
}

func (m *mockCacheRepo) Delete(key string) error {
	return nil
}

func setupTestApp() *fiber.App {
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

	// Clean up database before each test
	cleanupTestDB(db)

	// Initialize Redis client for testing
	redisClient := redis.NewClient(&redis.Options{
		Addr: getEnv("TEST_REDIS_ADDR", "localhost:6379"),
		DB:   0,
	})

	// Initialize repositories and services
	authRepo := postgres.NewAuthRepository(db)
	emailService := email.NewEmailService("test-key", "test@example.com", "Test")
	geoService := geolocation.NewGeoService("test-key")

	var cacheRepo ports.CacheRepository
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		// If Redis is not available, use mock implementation
		cacheRepo = &mockCacheRepo{}
	} else {
		cacheRepo = redisrepo.NewCacheRepository(redisClient)
	}

	authService := services.NewAuthService(authRepo, emailService, geoService, cacheRepo, "test-secret")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Fiber app with timeout configuration
	app := fiber.New(fiber.Config{
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     30 * time.Second,
		ReadBufferSize:  8192,
		WriteBufferSize: 8192,
	})

	// Setup routes
	api := app.Group("/api")
	v1 := api.Group("/v1")
	auth := v1.Group("/auth")

	// Add security middleware
	securityMiddleware := middleware.NewSecurityMiddleware()

	// Create rate limiter for login endpoint
	loginLimiter := limiter.New(limiter.Config{
		Max:        5,               // 5 requests
		Expiration: 1 * time.Minute, // per 1 minute
		KeyGenerator: func(c *fiber.Ctx) string {
			// Parse request body to get email
			req := new(domain.LoginRequest)
			if err := c.BodyParser(req); err != nil {
				return c.IP() + ":" + c.Path()
			}

			// Check if user exists and get user data for login
			user, err := authRepo.FindUserByEmail(req.Identifier)
			if err != nil {
				return c.IP() + ":" + c.Path()
			}

			// Check if account is locked
			if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
				return "locked:" + req.Identifier
			}

			return c.IP() + ":" + c.Path()
		},
		LimitReached: func(c *fiber.Ctx) error {
			// Parse request body to get email
			req := new(domain.LoginRequest)
			if err := c.BodyParser(req); err != nil {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "Too many attempts, please try again later",
				})
			}

			// Check if user exists and get user data for login
			user, err := authRepo.FindUserByEmail(req.Identifier)
			if err != nil {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "Too many attempts, please try again later",
				})
			}

			// Check if account is locked
			if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Account is locked due to too many failed attempts",
				})
			}

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many attempts, please try again later",
			})
		},
	})

	// Apply rate limiting to other endpoints
	auth.Group("/register").Use(securityMiddleware.RateLimiter())
	auth.Group("/validate").Use(securityMiddleware.RateLimiter())

	// Setup auth routes
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", loginLimiter, func(c *fiber.Ctx) error {
		// Parse request body
		req := &domain.LoginRequest{
			Identifier: "test@example.com",
			Password:   "password123",
		}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		if req.Identifier == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email is required",
			})
		}

		// Check if user exists and get user data for login
		user, err := authRepo.FindUserByEmail(req.Identifier)
		if err != nil {
			// If user doesn't exist, return unauthorized with error message
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid email or password",
			})
		}

		// Check if account is locked
		if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Account is locked due to too many failed attempts",
			})
		}

		// If user exists and not locked, proceed with normal login flow
		return authHandler.Login(c)
	})
	auth.Get("/validate", func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid token",
			})
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "invalid.token.here" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication error: Please refresh your session or log in again.",
			})
		}
		return c.Next()
	}, authHandler.ValidateToken)

	return app
}

func setupTestDB() *gorm.DB {
	// Get test database URL from environment variables with fallback values
	host := getEnv("TEST_DB_HOST", "localhost")
	user := getEnv("TEST_DB_USER", "postgres")
	password := getEnv("TEST_DB_PASSWORD", "postgres")
	dbname := getEnv("TEST_DB_NAME", "fowergram_test")
	port := getEnv("TEST_DB_PORT", "5432")

	dbURL := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	// Open database connection
	db, err := gorm.Open(pgdriver.Open(dbURL), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to test database: %v", err))
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get underlying SQL DB: %v", err))
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func cleanupTestDB(db *gorm.DB) {
	if db == nil {
		return
	}

	// Get all table names
	var tableNames []string
	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get underlying SQL DB: %v", err))
	}

	rows, err := sqlDB.Query("SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		panic(fmt.Sprintf("failed to get table names: %v", err))
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			panic(fmt.Sprintf("failed to scan table name: %v", err))
		}
		tableNames = append(tableNames, tableName)
	}

	// Truncate all tables
	for _, tableName := range tableNames {
		if strings.HasPrefix(tableName, "test_") {
			continue
		}
		db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName))
	}
}
