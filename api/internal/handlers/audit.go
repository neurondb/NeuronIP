package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/neurondb/NeuronIP/api/internal/audit"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* AuditHandler handles audit and activity requests */
type AuditHandler struct {
	auditService *audit.AuditService
}

/* NewAuditHandler creates a new audit handler */
func NewAuditHandler(auditService *audit.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

/* GetAuditEvents handles GET /api/v1/audit/events */
func (h *AuditHandler) GetAuditEvents(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("event_type")
	entityType := r.URL.Query().Get("entity_type")
	userID := r.URL.Query().Get("user_id")
	limit := 100

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	events, err := h.auditService.GetAuditEvents(r.Context(), eventType, entityType, userID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

/* GetActivityTimeline handles GET /api/v1/audit/activity */
func (h *AuditHandler) GetActivityTimeline(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	limit := 100

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	events, err := h.auditService.GetActivityTimeline(r.Context(), userID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timeline": events,
		"count":    len(events),
	})
}

/* GetComplianceTrail handles GET /api/v1/audit/compliance-trail */
func (h *AuditHandler) GetComplianceTrail(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	limit := 100

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	events, err := h.auditService.GetComplianceTrail(r.Context(), entityType, entityID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trail": events,
		"count": len(events),
	})
}

/* SearchAuditEvents handles POST /api/v1/audit/search */
func (h *AuditHandler) SearchAuditEvents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("query is required", nil))
		return
	}

	if req.Limit <= 0 {
		req.Limit = 100
	}

	events, err := h.auditService.SearchAuditEvents(r.Context(), req.Query, req.Limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}
