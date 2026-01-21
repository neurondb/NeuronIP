package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/config"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

var (
	command    = flag.String("command", "up", "Migration command: up, down, status, create")
	migrationName = flag.String("name", "", "Migration name (for create command)")
	steps      = flag.Int("steps", 0, "Number of steps (for down command, 0 = all)")
)

func main() {
	flag.Parse()

	cfg := config.Load()
	pool, err := db.NewPool(nil, cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrator := NewMigrator(pool.Pool)

	switch *command {
	case "up":
		if err := migrator.MigrateUp(); err != nil {
			fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations applied successfully")
	case "down":
		if err := migrator.MigrateDown(*steps); err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations rolled back successfully")
	case "status":
		if err := migrator.Status(); err != nil {
			fmt.Fprintf(os.Stderr, "Status check failed: %v\n", err)
			os.Exit(1)
		}
	case "create":
		if *migrationName == "" {
			fmt.Fprintf(os.Stderr, "Migration name is required\n")
			os.Exit(1)
		}
		if err := migrator.CreateMigration(*migrationName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create migration: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migration created successfully")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *command)
		flag.Usage()
		os.Exit(1)
	}
}

/* Migrator provides migration management */
type Migrator struct {
	pool *pgxpool.Pool
}

/* NewMigrator creates a new migrator */
func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{pool: pool}
}

/* MigrateUp applies all pending migrations */
func (m *Migrator) MigrateUp() error {
	// Ensure migrations table exists
	if err := m.ensureMigrationsTable(); err != nil {
		return err
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Get all migration files
	migrationsDir := "migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}

	// Sort files
	sort.Strings(files)

	// Apply pending migrations
	for _, file := range files {
		migrationName := filepath.Base(file)
		if contains(applied, migrationName) {
			continue
		}

		fmt.Printf("Applying migration: %s\n", migrationName)
		if err := m.applyMigration(file, migrationName); err != nil {
			return fmt.Errorf("failed to apply %s: %w", migrationName, err)
		}
	}

	return nil
}

/* MigrateDown rolls back migrations */
func (m *Migrator) MigrateDown(steps int) error {
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		fmt.Println("No migrations to roll back")
		return nil
	}

	// Sort applied migrations (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(applied)))

	rollbackCount := steps
	if rollbackCount == 0 {
		rollbackCount = len(applied)
	}

	for i := 0; i < rollbackCount && i < len(applied); i++ {
		migrationName := applied[i]
		fmt.Printf("Rolling back migration: %s\n", migrationName)
		if err := m.rollbackMigration(migrationName); err != nil {
			return fmt.Errorf("failed to rollback %s: %w", migrationName, err)
		}
	}

	return nil
}

/* Status shows migration status */
func (m *Migrator) Status() error {
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	migrationsDir := "migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	fmt.Println("Migration Status:")
	fmt.Println("=================")
	for _, file := range files {
		migrationName := filepath.Base(file)
		status := "pending"
		if contains(applied, migrationName) {
			status = "applied"
		}
		fmt.Printf("  %s: %s\n", migrationName, status)
	}

	return nil
}

/* CreateMigration creates a new migration file */
func (m *Migrator) CreateMigration(name string) error {
	timestamp := fmt.Sprintf("%03d", len(m.getMigrationFiles())+1)
	filename := fmt.Sprintf("%s_%s.sql", timestamp, strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	filepath := filepath.Join("migrations", filename)

	content := fmt.Sprintf(`-- Migration: %s
-- Description: %s

-- Add your migration SQL here

`, name, name)

	return os.WriteFile(filepath, []byte(content), 0644)
}

/* Helper functions */
func (m *Migrator) ensureMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS neuronip.migrations (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`
	_, err := m.pool.Exec(nil, query)
	return err
}

func (m *Migrator) getAppliedMigrations() ([]string, error) {
	query := `SELECT name FROM neuronip.migrations ORDER BY applied_at`
	rows, err := m.pool.Query(nil, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applied []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			applied = append(applied, name)
		}
	}

	return applied, nil
}

func (m *Migrator) applyMigration(filepath, name string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	// Execute migration
	if _, err := m.pool.Exec(nil, string(content)); err != nil {
		return err
	}

	// Record migration
	query := `INSERT INTO neuronip.migrations (name) VALUES ($1)`
	_, err = m.pool.Exec(nil, query, name)
	return err
}

func (m *Migrator) rollbackMigration(name string) error {
	// In production, you'd have rollback SQL files
	// For now, just remove from migrations table
	query := `DELETE FROM neuronip.migrations WHERE name = $1`
	_, err := m.pool.Exec(nil, query, name)
	return err
}

func (m *Migrator) getMigrationFiles() []string {
	files, _ := filepath.Glob("migrations/*.sql")
	return files
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
