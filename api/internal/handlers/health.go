package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* HealthHandler handles health check requests */
type HealthHandler struct {
	pool *pgxpool.Pool
}

/* NewHealthHandler creates a new health handler */
func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

/* HealthResponse represents the health check response */
type HealthResponse struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckStatus `json:"checks,omitempty"`
}

/* CheckStatus represents the status of a health check */
type CheckStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

/* ServeHTTP handles health check requests */
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "ok",
		Service:   "neuronip-api",
		Timestamp: time.Now(),
		Checks:    make(map[string]CheckStatus),
	}

	// Check database connectivity
	if h.pool != nil {
		if err := h.pool.Ping(ctx); err != nil {
			response.Status = "unhealthy"
			response.Checks["database"] = CheckStatus{
				Status:  "error",
				Message: err.Error(),
			}
		} else {
			stats := h.pool.Stat()
			response.Checks["database"] = CheckStatus{
				Status:  "healthy",
				Message: fmt.Sprintf("Pool: %d/%d connections", stats.AcquiredConns(), stats.MaxConns()),
			}
		}
	} else {
		response.Status = "unhealthy"
		response.Checks["database"] = CheckStatus{
			Status:  "error",
			Message: "Database pool not initialized",
		}
	}

	// Determine HTTP status code
	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
