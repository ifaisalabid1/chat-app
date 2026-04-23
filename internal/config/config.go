package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port            string        `envconfig:"PORT" default:"8080"`
	GracefulTimeout time.Duration `envconfig:"GRACEFUL_TIMEOUT" default:"5s"`
	DatabaseURL     string        `envconfig:"DATABASE_URL" required:"true"`
	RedisURL        string        `envconfig:"REDIS_URL" default:"localhost:6379"`
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("No .env file found or error reading it; relying on system environment variables")
	}

	var cfg Config
	err = envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process env variables: %w", err)
	}

	return &cfg, nil
}
