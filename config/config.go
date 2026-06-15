package config

import (
	"fmt"
	"os"
)

// Config holds all environment-driven settings for the app.
type Config struct {
	DatabaseURL string
	Port        string
	AppEnv      string
}

// Load reads configuration from environment variables and
// fails fast if anything required is missing.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		AppEnv:      os.Getenv("APP_ENV"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}
	return cfg, nil
}
