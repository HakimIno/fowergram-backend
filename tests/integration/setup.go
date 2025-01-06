package integration

import (
	"fmt"
	"fowergram/internal/core/domain"
	"fowergram/internal/core/services"
	"fowergram/internal/handlers"
	"fowergram/internal/repositories/postgres"
	redisrepo "fowergram/internal/repositories/redis"
	"fowergram/pkg/email"
	"fowergram/pkg/geolocation"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
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

	// Initialize Redis client for testing
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// Initialize repositories and services
	authRepo := postgres.NewAuthRepository(db)
	emailService := email.NewEmailService("test-key", "test@example.com", "Test")
	geoService := geolocation.NewGeoService("test-key")
	cacheRepo := redisrepo.NewCacheRepository(redisClient)
	authService := services.NewAuthService(authRepo, emailService, geoService, cacheRepo, "test-secret")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Fiber app
	app := fiber.New()

	// Setup routes
	api := app.Group("/api")
	v1 := api.Group("/v1")
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

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
		db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName))
	}
}
