package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/integrations"
)

/* IntegrationHandler handles integration requests */
type IntegrationHandler struct {
	helpdeskService *integrations.HelpdeskService
}

/* NewIntegrationHandler creates a new integration handler */
func NewIntegrationHandler(helpdeskService *integrations.HelpdeskService) *IntegrationHandler {
	return &IntegrationHandler{helpdeskService: helpdeskService}
}

/* SyncHelpdeskRequest represents helpdesk sync request */
type SyncHelpdeskRequest struct {
	Config integrations.HelpdeskConfig `json:"config"`
}

/* SyncHelpdesk handles helpdesk sync requests */
func (h *IntegrationHandler) SyncHelpdesk(w http.ResponseWriter, r *http.Request) {
	var req SyncHelpdeskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Config.Provider == "" {
		WriteErrorResponse(w, errors.ValidationFailed("provider is required", nil))
		return
	}

	count, err := h.helpdeskService.SyncTickets(r.Context(), req.Config, nil)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"synced_count": count,
		"status":       "completed",
	})
}
