package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Context key for database selection */
type contextKey string

const databaseContextKey contextKey = "database_pool"

/* WithDatabase adds a database pool to the context */
func WithDatabase(ctx context.Context, pool *pgxpool.Pool) context.Context {
	return context.WithValue(ctx, databaseContextKey, pool)
}

/* GetDatabaseFromContext gets the database pool from context */
func GetDatabaseFromContext(ctx context.Context) (*pgxpool.Pool, bool) {
	pool, ok := ctx.Value(databaseContextKey).(*pgxpool.Pool)
	return pool, ok
}

/* GetDatabaseNameFromContext gets the database name from context (from session) */
func GetDatabaseNameFromContext(ctx context.Context) string {
	if dbVal := ctx.Value("database"); dbVal != nil {
		if dbStr, ok := dbVal.(string); ok && dbStr != "" {
			return dbStr
		}
	}
	return "neuronip" // default
}
