package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func New(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("database connected", "url", maskURL(databaseURL))

	return &DB{db}, nil
}

func (db *DB) Close() error {
	slog.Info("database disconnected")
	return db.DB.Close()
}

func (db *DB) Transactional(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func maskURL(url string) string {
	if idx := len(url); idx > 0 {
		return url[:idx]
	}
	return "***"
}

// NewTestDB creates a test database connection for integration tests
func NewTestDB(t testing.TB) *DB {
	t.Helper()

	// Use test database URL
	testURL := "postgres://fatoumata@localhost:5432/minidb?sslmode=disable"

	db, err := New(testURL)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}

	return db
}
