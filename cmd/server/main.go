package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pension-manager/internal/api"
	"pension-manager/internal/config"
	"pension-manager/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from project root
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
	} else {
		slog.Info(".env file loaded successfully")
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))
	slog.SetDefault(logger)

	// Debug: log if NewsAPI key is loaded
	if cfg.NewsAPI.APIKey != "" {
		slog.Info("NewsAPI key loaded",
			"length", len(cfg.NewsAPI.APIKey),
			"key_prefix", string([]byte(cfg.NewsAPI.APIKey)[:8])+"...",
			"key_suffix", "..."+string([]byte(cfg.NewsAPI.APIKey)[len(cfg.NewsAPI.APIKey)-8:]))
	} else {
		slog.Warn("NewsAPI key is empty - will use mock data")
		slog.Info("Checking environment directly:",
			"NEWS_API_KEY from os.Getenv:", os.Getenv("NEWS_API_KEY"))
	}

	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	server := api.New(database, cfg)

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	slog.Info("starting server", "port", cfg.HTTPPort, "env", cfg.Env)

	go func() {
		if err := server.Start(addr); err != nil {
			slog.Error("http server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
