package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* DatabaseTestRequest represents a database connection test request */
type DatabaseTestRequest struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode,omitempty"`
}

/* DatabaseTestResponse represents a database connection test response */
type DatabaseTestResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Version   string `json:"version,omitempty"`
}

/* TestDatabaseConnection tests a PostgreSQL database connection */
func TestDatabaseConnection(w http.ResponseWriter, r *http.Request) {
	var req DatabaseTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	// Validate required fields
	if req.Host == "" || req.Port == "" || req.Database == "" || req.User == "" {
		WriteErrorResponse(w, errors.BadRequest("Host, port, database, and user are required"))
		return
	}

	// Default SSL mode
	if req.SSLMode == "" {
		req.SSLMode = "disable"
	}

	// Build connection string
	connString := buildConnectionString(req)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Now()
	
	// Create a temporary pool to test connection
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid connection string: "+err.Error()))
		return
	}

	// Set minimal pool settings for testing
	config.MaxConns = 1
	config.MinConns = 0

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DatabaseTestResponse{
			Success: false,
			Message: "Failed to create connection pool: " + err.Error(),
		})
		return
	}
	defer pool.Close()

	// Test ping
	if err := pool.Ping(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DatabaseTestResponse{
			Success: false,
			Message: "Connection failed: " + err.Error(),
		})
		return
	}

	// Get PostgreSQL version
	var version string
	err = pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		version = "Unknown"
	}

	latency := time.Since(startTime).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DatabaseTestResponse{
		Success:   true,
		Message:   "Connection successful",
		LatencyMs: latency,
		Version:   version,
	})
}

/* buildConnectionString builds a PostgreSQL connection string */
func buildConnectionString(req DatabaseTestRequest) string {
	connStr := "host=" + req.Host +
		" port=" + req.Port +
		" user=" + req.User +
		" password=" + req.Password +
		" dbname=" + req.Database +
		" sslmode=" + req.SSLMode

	return connStr
}
