package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv          string
	LogLevel        string
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	NATSURL         string
	PublicBaseURL   string
	APIAddr         string
	RedirectAddr    string
	CacheTTL        time.Duration
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		AppEnv:          get("APP_ENV", "local"),
		LogLevel:        get("LOG_LEVEL", "info"),
		DatabaseURL:     get("DATABASE_URL", "postgres://shortener:shortener@localhost:5432/shortener?sslmode=disable"),
		RedisAddr:       get("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   get("REDIS_PASSWORD", ""),
		RedisDB:         getInt("REDIS_DB", 0),
		NATSURL:         get("NATS_URL", ""),
		PublicBaseURL:   get("PUBLIC_BASE_URL", "http://localhost:8081"),
		APIAddr:         get("API_ADDR", ":8080"),
		RedirectAddr:    get("REDIRECT_ADDR", ":8081"),
		CacheTTL:        getDuration("CACHE_TTL", 24*time.Hour),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
