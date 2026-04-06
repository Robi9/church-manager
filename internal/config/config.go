package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", ""),
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
