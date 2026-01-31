package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

// SetupTestDB creates a test database connection pool (fails test if DB unavailable).
func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	pool, cleanup, err := setupTestDBWithError()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	return pool, cleanup
}

// SetupTestDBOrSkip creates a test database connection pool, or skips the test if DB is unavailable.
func SetupTestDBOrSkip(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	pool, cleanup, err := setupTestDBWithError()
	if err != nil {
		t.Skipf("Skipping: test DB not available: %v", err)
	}
	return pool, cleanup
}

func setupTestDBWithError() (*pgxpool.Pool, func(), error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "testuser")
	dbPassword := getEnv("DB_PASSWORD", "testpass")
	dbName := getEnv("DB_NAME", "neuronip_test")

	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, nil, fmt.Errorf("parse config: %w", err)
	}

	config.MaxConns = 5
	config.MinConns = 1

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("ping: %w", err)
	}

	cleanup := func() { pool.Close() }
	return pool, cleanup, nil
}

// SetupTestQueries creates a test queries instance
func SetupTestQueries(t *testing.T) (*db.Queries, func()) {
	t.Helper()

	pool, cleanup := SetupTestDB(t)
	queries := db.NewQueries(pool)

	return queries, cleanup
}

// TruncateTables truncates all tables in the test database
func TruncateTables(t *testing.T, pool *pgxpool.Pool, tables []string) {
	t.Helper()

	ctx := context.Background()
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// SeedTestData seeds test data into the database
func SeedTestData(t *testing.T, pool *pgxpool.Pool, seedSQL string) {
	t.Helper()

	ctx := context.Background()
	_, err := pool.Exec(ctx, seedSQL)
	if err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
