package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	HTTPPort    int
	JWTSecret   string
	LogLevel    string
	Env         string
	Mpesa       MpesaConfig
}

type MpesaConfig struct {
	ConsumerKey    string
	ConsumerSecret string
	ShortCode      string
	Passkey        string
	CallbackURL    string
	APIVersion     string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/minidb?sslmode=disable"),
		HTTPPort:    getEnvInt("HTTP_PORT", 8080),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
		LogLevel:    strings.ToLower(getEnv("LOG_LEVEL", "info")),
		Env:         getEnv("APP_ENV", "development"),
		Mpesa: MpesaConfig{
			ConsumerKey:    getEnv("MPESA_CONSUMER_KEY", ""),
			ConsumerSecret: getEnv("MPESA_CONSUMER_SECRET", ""),
			ShortCode:      getEnv("MPESA_SHORTCODE", "174379"),
			Passkey:        getEnv("MPESA_PASSKEY", ""),
			CallbackURL:    getEnv("MPESA_CALLBACK_URL", "http://localhost:8080/mpesa/callback"),
			APIVersion:     getEnv("MPESA_API_VERSION", "v3"),
		},
	}

	if cfg.JWTSecret == "change-me-in-production" && cfg.Env == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fallback
		}
		return n
	}
	return fallback
}
