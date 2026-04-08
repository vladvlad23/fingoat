package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	Port             string
	RefreshTokenTTL  time.Duration
	DisableRegister  bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	cfg.DisableRegister = os.Getenv("DISABLE_REGISTER") == "true"
	if raw := os.Getenv("REFRESH_TOKEN_TTL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("REFRESH_TOKEN_TTL: %w", err)
		}
		cfg.RefreshTokenTTL = d
	} else {
		cfg.RefreshTokenTTL = 30 * 24 * time.Hour
	}
	return cfg, nil
}
