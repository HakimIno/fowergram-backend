package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Server   ServerConfig
	DB       *gorm.DB
	MongoDB  MongoDBConfig
	Redis    *redis.Client
	Redpanda RedpandaConfig
	Email    EmailConfig
	Geo      GeoConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	User     string
	Password string
	Name     string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RedpandaConfig struct {
	Broker string
}

type EmailConfig struct {
	APIKey      string
	SenderEmail string
	SenderName  string
}

type GeoConfig struct {
	APIKey string
}

type JWTConfig struct {
	Secret string
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	// Setup Database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		viper.GetString("DB_HOST"),
		viper.GetString("DB_USER"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Setup Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("REDIS_HOST"), viper.GetString("REDIS_PORT")),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       0,
	})

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: viper.GetString("PORT"),
		},
		DB: db,
		MongoDB: MongoDBConfig{
			URI:      viper.GetString("MONGODB_URI"),
			Database: viper.GetString("MONGODB_DATABASE"),
		},
		Redis: redisClient,
		Redpanda: RedpandaConfig{
			Broker: viper.GetString("REDPANDA_BROKER"),
		},
		Email: EmailConfig{
			APIKey:      viper.GetString("EMAIL_API_KEY"),
			SenderEmail: viper.GetString("EMAIL_SENDER"),
			SenderName:  viper.GetString("EMAIL_SENDER_NAME"),
		},
		Geo: GeoConfig{
			APIKey: viper.GetString("GEO_API_KEY"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("JWT_SECRET"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
