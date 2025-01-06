package config

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Server ServerConfig
	DB     *gorm.DB
	Redis  *redis.Client
	JWT    JWTConfig
	Email  EmailConfig
	Geo    GeoConfig
}

type ServerConfig struct {
	Port string
}

type JWTConfig struct {
	Secret string
}

type EmailConfig struct {
	APIKey      string
	SenderEmail string
	SenderName  string
}

type GeoConfig struct {
	APIKey string
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
		return nil, err
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set max number of open connections
	sqlDB.SetMaxOpenConns(25)
	// Set max number of idle connections
	sqlDB.SetMaxIdleConns(10)
	// Set max lifetime of connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Setup Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("REDIS_HOST"), viper.GetString("REDIS_PORT")),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       0,
	})

	return &Config{
		Server: ServerConfig{
			Port: viper.GetString("PORT"),
		},
		DB:    db,
		Redis: rdb,
		JWT: JWTConfig{
			Secret: viper.GetString("JWT_SECRET"),
		},
		Email: EmailConfig{
			APIKey:      viper.GetString("EMAIL_API_KEY"),
			SenderEmail: viper.GetString("EMAIL_SENDER_EMAIL"),
			SenderName:  viper.GetString("EMAIL_SENDER_NAME"),
		},
		Geo: GeoConfig{
			APIKey: viper.GetString("GEO_API_KEY"),
		},
	}, nil
}
