package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/lineage"
)

/* DiscoveryHandler handles lineage discovery requests */
type DiscoveryHandler struct {
	discoveryService *lineage.DiscoveryService
}

/* NewDiscoveryHandler creates a new discovery handler */
func NewDiscoveryHandler(discoveryService *lineage.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{discoveryService: discoveryService}
}

/* RunDiscovery handles POST /api/v1/lineage/discovery/run */
func (h *DiscoveryHandler) RunDiscovery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RuleID *uuid.UUID `json:"rule_id,omitempty"`
	}

	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
			return
		}
	}

	discovered, err := h.discoveryService.RunDiscovery(r.Context(), req.RuleID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"discovered": discovered,
		"count":      len(discovered),
	})
}

/* CreateDiscoveryRule handles POST /api/v1/lineage/discovery/rules */
func (h *DiscoveryHandler) CreateDiscoveryRule(w http.ResponseWriter, r *http.Request) {
	var rule lineage.DiscoveryRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.discoveryService.CreateDiscoveryRule(r.Context(), rule)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* VerifyDiscoveredLineage handles POST /api/v1/lineage/discovery/{id}/verify */
func (h *DiscoveryHandler) VerifyDiscoveredLineage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	discoveredID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid discovered lineage ID"))
		return
	}

	if err := h.discoveryService.VerifyDiscoveredLineage(r.Context(), discoveredID); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "verified",
	})
}
