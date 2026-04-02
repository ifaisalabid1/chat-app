package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP     HTTPConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	LogLevel string
}

type HTTPConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DBConfig struct {
	DSN         string
	MaxConns    int32
	MinConns    int32
	MaxIdleTime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	cfg := &Config{
		HTTP: HTTPConfig{
			Port:         getEnv("HTTP_PORT", "8080"),
			ReadTimeout:  mustDuration("HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: mustDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  mustDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		DB: DBConfig{
			DSN:         requireEnv("DATABASE_URL"),
			MaxConns:    int32(mustInt("DB_MAX_CONNS", 25)),
			MinConns:    int32(mustInt("DB_MIN_CONNS", 2)),
			MaxIdleTime: mustDuration("DB_MAX_IDLE_TIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       mustInt("REDIS_DB", 0),
		},
		// JWT: JWTConfig{
		// 	Secret: requireEnv("JWT_SECRET"),
		// 	Expiry: mustDuration("JWT_EXPIRY", 15*time.Minute),
		// },
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func mustDuration(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		duration, err := time.ParseDuration(v)
		if err != nil {
			panic(fmt.Sprintf("env %s must be a duration: %v", key, err))
		}
		return duration
	}
	return defaultValue
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env %s is not set", key))
	}
	return v
}

func mustInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(fmt.Sprintf("env %s must be an integer: %v", key, err))
		}
		return i
	}
	return defaultValue
}
