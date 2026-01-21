package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"golang.org/x/crypto/bcrypt"
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

	// Generate API key
	keyPrefix := "sk_live_"
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		WriteError(w, errors.InternalServer("Failed to generate API key"))
		return
	}
	apiKey := keyPrefix + base64.URLEncoding.EncodeToString(keyBytes)

	// Hash the API key for storage
	hasher := sha256.New()
	hasher.Write([]byte(apiKey))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	// Bcrypt hash for secure storage
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(keyHash), bcrypt.DefaultCost)
	if err != nil {
		WriteError(w, errors.InternalServer("Failed to hash API key"))
		return
	}

	keyID := uuid.New()

	// Insert API key into database
	query := `
		INSERT INTO neuronip.api_keys 
		(id, key_hash, key_prefix, user_id, name, rate_limit, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, key_prefix, user_id, name, rate_limit, created_at, expires_at, last_used_at`

	var key db.APIKey
	err = h.queries.DB.QueryRow(
		r.Context(),
		query,
		keyID, string(hashedKey), keyPrefix, userIDStr, req.Name, req.RateLimit,
	).Scan(
		&key.ID, &key.KeyPrefix, &key.UserID, &key.Name,
		&key.RateLimit, &key.CreatedAt, &key.ExpiresAt, &key.LastUsedAt,
	)
	if err != nil {
		WriteError(w, err)
		return
	}

	key.KeyHash = string(hashedKey)

	// Return the API key (only shown once)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        key.ID,
		"api_key":   apiKey, // Only returned once
		"key_prefix": key.KeyPrefix,
		"name":      key.Name,
		"rate_limit": key.RateLimit,
		"created_at": key.CreatedAt,
		"warning":   "Save this API key now. It will not be shown again.",
	})
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
