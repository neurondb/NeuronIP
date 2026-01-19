package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/integrations"
)

/* WebhookHandler handles webhook requests */
type WebhookHandler struct {
	service *integrations.WebhookService
}

/* NewWebhookHandler creates a new webhook handler */
func NewWebhookHandler(service *integrations.WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

/* CreateWebhook handles POST /api/v1/webhooks */
func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var webhook integrations.Webhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateWebhook(r.Context(), webhook)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetWebhook handles GET /api/v1/webhooks/{id} */
func (h *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid webhook ID"))
		return
	}

	webhook, err := h.service.GetWebhook(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

/* ListWebhooks handles GET /api/v1/webhooks */
func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	webhooks, err := h.service.ListWebhooks(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}

/* TriggerWebhook handles POST /api/v1/webhooks/{id}/trigger */
func (h *WebhookHandler) TriggerWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid webhook ID"))
		return
	}

	var req struct {
		Event   string                 `json:"event"`
		Payload map[string]interface{} `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	err = h.service.TriggerWebhook(r.Context(), id, req.Event, req.Payload)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "triggered",
	})
}
