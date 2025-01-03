package config

import (
	"fmt"

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
}

type ServerConfig struct {
	Port string
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
		return nil, err
	}

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
	}, nil
}
