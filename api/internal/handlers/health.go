package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/mcp"
)

/* HealthHandler handles health check requests */
type HealthHandler struct {
	pool      *pgxpool.Pool
	mcpClient *mcp.Client
	startTime time.Time
}

/* NewHealthHandler creates a new health handler */
func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		pool:      pool,
		mcpClient: nil,
		startTime: time.Now(),
	}
}

/* NewHealthHandlerWithMCP creates a new health handler with MCP client */
func NewHealthHandlerWithMCP(pool *pgxpool.Pool, mcpClient *mcp.Client) *HealthHandler {
	return &HealthHandler{
		pool:      pool,
		mcpClient: mcpClient,
		startTime: time.Now(),
	}
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

	// Comprehensive database health checks
	if h.pool != nil {
		healthChecker := db.NewHealthChecker(h.pool)
		healthy, results := healthChecker.IsHealthy(ctx)
		
		if !healthy {
			response.Status = "unhealthy"
		}
		
		// Add detailed check results
		for checkName, result := range results {
			status := "healthy"
			if !result.Healthy {
				status = "error"
				if response.Status == "ok" {
					response.Status = "unhealthy"
				}
			}
			
			message := result.Message
			if result.Latency > 0 {
				message = fmt.Sprintf("%s (latency: %v)", message, result.Latency)
			}
			if result.Connections != nil {
				message = fmt.Sprintf("%s [Pool: %d/%d acquired, %d idle]", 
					message, result.Connections.AcquiredConns, result.Connections.MaxConns, result.Connections.IdleConns)
			}
			
			response.Checks[fmt.Sprintf("database_%s", checkName)] = CheckStatus{
				Status:  status,
				Message: message,
			}
		}
	} else {
		response.Status = "unhealthy"
		response.Checks["database"] = CheckStatus{
			Status:  "error",
			Message: "Database pool not initialized",
		}
	}

	// Check MCP client connectivity
	if h.mcpClient != nil {
		// Try to list tools to verify MCP is working
		mcpCtx, mcpCancel := context.WithTimeout(ctx, 5*time.Second)
		defer mcpCancel()

		if _, err := h.mcpClient.ListTools(mcpCtx); err != nil {
			response.Checks["mcp"] = CheckStatus{
				Status:  "error",
				Message: fmt.Sprintf("MCP client error: %v", err),
			}
			if response.Status == "ok" {
				response.Status = "warning"
			}
		} else {
			response.Checks["mcp"] = CheckStatus{
				Status:  "healthy",
				Message: "MCP client is connected and responding",
			}
		}
	} else {
		// Check if MCP binary path exists (optional service)
		mcpPath := os.Getenv("NEURONMCP_BINARY_PATH")
		if mcpPath == "" {
			mcpPath = "/usr/local/bin/neurondb-mcp"
		}
		if _, err := os.Stat(mcpPath); err == nil {
			response.Checks["mcp"] = CheckStatus{
				Status:  "unavailable",
				Message: "MCP binary found but client not initialized",
			}
		} else {
			response.Checks["mcp"] = CheckStatus{
				Status:  "unavailable",
				Message: "MCP binary not found or not configured",
			}
		}
	}

	// Add uptime information
	uptime := time.Since(h.startTime)
	response.Checks["uptime"] = CheckStatus{
		Status:  "healthy",
		Message: fmt.Sprintf("Server uptime: %s", uptime.Round(time.Second).String()),
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
