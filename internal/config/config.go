package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL          string
	Port                 string
	JWTSecret            string
	AllowedOrigins       []string
	AllowRegistration    bool
	InitialAdminEmail    string
	InitialAdminPassword string
	Environment          string
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		Port:                 getEnv("PORT", "8080"),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		AllowedOrigins:       splitCSV(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
		AllowRegistration:    getEnvBool("ALLOW_REGISTRATION", false),
		InitialAdminEmail:    strings.TrimSpace(os.Getenv("INITIAL_ADMIN_EMAIL")),
		InitialAdminPassword: os.Getenv("INITIAL_ADMIN_PASSWORD"),
		Environment:          strings.ToLower(getEnv("APP_ENV", "development")),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if cfg.Environment == "production" && len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must have at least 32 characters in production")
	}
	if len(cfg.AllowedOrigins) == 0 {
		log.Fatal("ALLOWED_ORIGINS must contain at least one origin")
	}
	if (cfg.InitialAdminEmail == "") != (cfg.InitialAdminPassword == "") {
		log.Fatal("INITIAL_ADMIN_EMAIL and INITIAL_ADMIN_PASSWORD must be set together")
	}
	if cfg.InitialAdminPassword != "" && len(cfg.InitialAdminPassword) < 12 {
		log.Fatal("INITIAL_ADMIN_PASSWORD must have at least 12 characters")
	}

	if err := validateProduction(cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}

func splitCSV(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatal(fmt.Sprintf("%s must be true or false", key))
	}
	return parsed
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func validateProduction(cfg *Config) error {
	if cfg.Environment != "production" {
		return nil
	}
	database, err := url.Parse(cfg.DatabaseURL)
	if err != nil || (database.Scheme != "postgres" && database.Scheme != "postgresql") || database.Hostname() == "" {
		return fmt.Errorf("production DATABASE_URL must be a PostgreSQL URL")
	}
	mode := database.Query().Get("sslmode")
	if mode != "require" && mode != "verify-ca" && mode != "verify-full" {
		return fmt.Errorf("production DATABASE_URL must require TLS")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return fmt.Errorf("production ALLOWED_ORIGINS must be configured")
	}
	if len(cfg.JWTSecret) < 32 || strings.Contains(strings.ToLower(cfg.JWTSecret), "replace-with") {
		return fmt.Errorf("production JWT_SECRET must be a random secret of at least 32 characters")
	}
	if cfg.AllowRegistration {
		return fmt.Errorf("public registration must be disabled in production")
	}
	for _, origin := range cfg.AllowedOrigins {
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || strings.Contains(origin, "*") || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return fmt.Errorf("production ALLOWED_ORIGINS must contain exact HTTPS origins")
		}
	}
	return nil
}
