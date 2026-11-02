package config

import (
	"strings"
	"testing"
)

func TestProductionRejectsUnsafeConfiguration(t *testing.T) {
	valid := func() Config {
		return Config{
			Environment: "production", DatabaseURL: "postgresql://user:password@db.example.org/db?sslmode=require",
			JWTSecret: strings.Repeat("a", 48), AllowedOrigins: []string{"https://church.example.org"},
		}
	}
	good := valid()
	if err := validateProduction(&good); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Config){
		func(c *Config) { c.DatabaseURL = "postgresql://db.example.org/db?sslmode=disable" },
		func(c *Config) { c.AllowedOrigins = []string{"http://church.example.org"} },
		func(c *Config) { c.AllowedOrigins = []string{"https://*.example.org"} },
		func(c *Config) { c.AllowRegistration = true },
		func(c *Config) { c.JWTSecret = "replace-with-at-least-32-random-characters" },
	} {
		cfg := valid()
		mutate(&cfg)
		if validateProduction(&cfg) == nil {
			t.Fatal("unsafe production configuration accepted")
		}
	}
}
