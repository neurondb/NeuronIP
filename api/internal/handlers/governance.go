package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/warehouse"
)

/* GovernanceHandler handles query governance requests */
type GovernanceHandler struct {
	service *warehouse.GovernanceService
}

/* NewGovernanceHandler creates a new governance handler */
func NewGovernanceHandler(service *warehouse.GovernanceService) *GovernanceHandler {
	return &GovernanceHandler{service: service}
}

/* ValidateQuery handles query validation */
func (h *GovernanceHandler) ValidateQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QueryText string `json:"query_text"`
		UserRole  string `json:"user_role"`
		UserID    string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = req.UserID
	}

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	result, err := h.service.ValidateQuery(r.Context(), req.QueryText, req.UserRole, userIDPtr)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* SanitizeQuery handles query sanitization */
func (h *GovernanceHandler) SanitizeQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QueryText string `json:"query_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	sanitized := h.service.SanitizeQuery(req.QueryText)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"sanitized_query": sanitized,
	})
}
