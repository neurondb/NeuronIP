package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/integrations"
)

/* IntegrationMarketplaceHandler handles integration marketplace requests */
type IntegrationMarketplaceHandler struct {
	marketplaceService *integrations.MarketplaceService
}

/* NewIntegrationMarketplaceHandler creates a new integration marketplace handler */
func NewIntegrationMarketplaceHandler(service *integrations.MarketplaceService) *IntegrationMarketplaceHandler {
	return &IntegrationMarketplaceHandler{marketplaceService: service}
}

/* ListIntegrations handles GET /api/v1/integrations/marketplace */
func (h *IntegrationMarketplaceHandler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	integrations, err := h.marketplaceService.ListIntegrations(r.Context(), category)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(integrations)
}

/* InstallIntegration handles POST /api/v1/integrations/marketplace/{id}/install */
func (h *IntegrationMarketplaceHandler) InstallIntegration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	integrationID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid integration ID"))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	var req struct {
		Config map[string]interface{} `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.marketplaceService.InstallIntegration(r.Context(), integrationID, userID, req.Config); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"integration_id": integrationID,
		"status":         "installed",
	})
}
