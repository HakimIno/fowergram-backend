package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server  ServerConfig
	MongoDB MongoDBConfig
	Redis   RedisConfig
	JWT     JWTConfig
}

type ServerConfig struct {
	Port string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type JWTConfig struct {
	Secret string
}

func Load() (*Config, error) {
	redisPort, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		redisPort = 6379 // default port
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: os.Getenv("PORT"),
		},
		MongoDB: MongoDBConfig{
			URI:      os.Getenv("MONGODB_URI"),
			Database: os.Getenv("MONGODB_DATABASE"),
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     redisPort,
			Password: os.Getenv("REDIS_PASSWORD"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	if cfg.MongoDB.URI == "" {
		return fmt.Errorf("MONGODB_URI is required")
	}
	if cfg.MongoDB.Database == "" {
		return fmt.Errorf("MONGODB_DATABASE is required")
	}
	if cfg.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

func GetJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}
