package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/webhooks"
)

/* WebhooksHandler handles webhook requests */
type WebhooksHandler struct {
	service *webhooks.Service
}

/* NewWebhooksHandler creates a new webhooks handler */
func NewWebhooksHandler(service *webhooks.Service) *WebhooksHandler {
	return &WebhooksHandler{service: service}
}

/* CreateWebhook creates a new webhook */
func (h *WebhooksHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var webhook webhooks.Webhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	// Get user ID from context
	userID := r.Header.Get("X-User-ID")
	if userID != "" {
		webhook.CreatedBy = &userID
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

/* GetWebhook retrieves a webhook */
func (h *WebhooksHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
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

/* ListWebhooks lists all webhooks */
func (h *WebhooksHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	enabledOnly := r.URL.Query().Get("enabled_only") == "true"
	
	webhooks, err := h.service.ListWebhooks(r.Context(), enabledOnly)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}

/* TriggerWebhook manually triggers a webhook */
func (h *WebhooksHandler) TriggerWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid webhook ID"))
		return
	}

	var req struct {
		EventType string                 `json:"event_type"`
		Payload   map[string]interface{} `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.DeliverWebhook(r.Context(), id, req.EventType, req.Payload); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
