package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/integrations"
)

/* IntegrationHandler handles integration requests */
type IntegrationHandler struct {
	integrationsService *integrations.IntegrationsService
	helpdeskService     *integrations.HelpdeskService
	webhookService      *integrations.WebhookService
}

/* NewIntegrationHandler creates a new integration handler */
func NewIntegrationHandler(integrationsService *integrations.IntegrationsService, helpdeskService *integrations.HelpdeskService, webhookService *integrations.WebhookService) *IntegrationHandler {
	return &IntegrationHandler{
		integrationsService: integrationsService,
		helpdeskService:     helpdeskService,
		webhookService:      webhookService,
	}
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

/* ListIntegrations handles GET /api/v1/integrations */
func (h *IntegrationHandler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	integrationType := r.URL.Query().Get("type")
	var integrationTypePtr *string
	if integrationType != "" {
		integrationTypePtr = &integrationType
	}

	integrations, err := h.integrationsService.ListIntegrations(r.Context(), integrationTypePtr)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(integrations)
}

/* GetIntegration handles GET /api/v1/integrations/{id} */
func (h *IntegrationHandler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid integration ID"))
		return
	}

	integration, err := h.integrationsService.GetIntegration(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(integration)
}

/* CreateIntegration handles POST /api/v1/integrations */
func (h *IntegrationHandler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	var req integrations.Integration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name is required", nil))
		return
	}
	if req.IntegrationType == "" {
		WriteErrorResponse(w, errors.ValidationFailed("integration_type is required", nil))
		return
	}

	integration, err := h.integrationsService.CreateIntegration(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(integration)
}

/* UpdateIntegration handles PUT /api/v1/integrations/{id} */
func (h *IntegrationHandler) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid integration ID"))
		return
	}

	var req integrations.Integration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	integration, err := h.integrationsService.UpdateIntegration(r.Context(), id, req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(integration)
}

/* DeleteIntegration handles DELETE /api/v1/integrations/{id} */
func (h *IntegrationHandler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid integration ID"))
		return
	}

	if err := h.integrationsService.DeleteIntegration(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* TestIntegration handles POST /api/v1/integrations/{id}/test */
func (h *IntegrationHandler) TestIntegration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid integration ID"))
		return
	}

	if err := h.integrationsService.TestIntegration(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Integration test passed",
	})
}

/* GetIntegrationHealth handles GET /api/v1/integrations/health */
func (h *IntegrationHandler) GetIntegrationHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.integrationsService.GetIntegrationHealth(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
