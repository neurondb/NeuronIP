package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* APIKeyHandler handles API key management requests */
type APIKeyHandler struct {
	queries *db.Queries
}

/* NewAPIKeyHandler creates a new API key handler */
func NewAPIKeyHandler(queries *db.Queries) *APIKeyHandler {
	return &APIKeyHandler{queries: queries}
}

/* ListAPIKeys handles listing user API keys */
func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	// Query API keys for user
	query := `SELECT id, key_hash, key_prefix, user_id, name, rate_limit, last_used_at, expires_at, created_at 
	          FROM neuronip.api_keys WHERE user_id = $1 ORDER BY created_at DESC`
	
	rows, err := h.queries.DB.Query(r.Context(), query, userIDStr)
	if err != nil {
		WriteError(w, err)
		return
	}
	defer rows.Close()

	var keys []db.APIKey
	for rows.Next() {
		var key db.APIKey
		if err := rows.Scan(
			&key.ID, &key.KeyHash, &key.KeyPrefix, &key.UserID, &key.Name,
			&key.RateLimit, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt,
		); err != nil {
			WriteError(w, err)
			return
		}
		keys = append(keys, key)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

/* CreateAPIKey handles creating an API key */
func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	var req struct {
		Name      string `json:"name"`
		RateLimit int    `json:"rate_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.RateLimit == 0 {
		req.RateLimit = 100
	}

	// API key creation logic would go here
	// For now, return placeholder
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "API key creation not fully implemented"})
}

/* UpdateAPIKey handles updating an API key */
func (h *APIKeyHandler) UpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid API key ID"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* DeleteAPIKey handles deleting an API key */
func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid API key ID"))
		return
	}

	query := `DELETE FROM neuronip.api_keys WHERE id = $1`
	if _, err := h.queries.DB.Exec(r.Context(), query, keyID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
