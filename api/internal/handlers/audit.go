package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

	filters := audit.AuditFilters{
		UserID:       &userID,
		ActionType:   &eventType,
		ResourceType: &entityType,
	}
	events, err := h.auditService.GetAuditLogs(r.Context(), filters, limit)
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

	filters := audit.AuditFilters{UserID: &userID}
	events, err := h.auditService.GetAuditLogs(r.Context(), filters, limit)
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

	filters := audit.AuditFilters{
		ResourceType: &entityType,
		ResourceID:   &entityID,
	}
	events, err := h.auditService.GetAuditLogs(r.Context(), filters, limit)
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

	// Simple search - in production, implement full-text search
	filters := audit.AuditFilters{}
	events, err := h.auditService.GetAuditLogs(r.Context(), filters, req.Limit)
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

/* ExportAuditEvents handles GET /api/v1/audit/export - returns audit events as CSV or JSON download */
func (h *AuditHandler) ExportAuditEvents(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("event_type")
	entityType := r.URL.Query().Get("entity_type")
	userID := r.URL.Query().Get("user_id")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	limit := 5000
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var filters audit.AuditFilters
	if userID != "" {
		filters.UserID = &userID
	}
	if eventType != "" {
		filters.ActionType = &eventType
	}
	if entityType != "" {
		filters.ResourceType = &entityType
	}
	events, err := h.auditService.GetAuditLogs(r.Context(), filters, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	filename := "audit_events_" + time.Now().Format("2006-01-02") + "." + format
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		cw := csv.NewWriter(w)
		_ = cw.Write([]string{"id", "action_type", "resource_type", "resource_id", "user_id", "action", "details", "created_at"})
		for _, e := range events {
			details := ""
			if e.Details != nil {
				if b, err := json.Marshal(e.Details); err == nil {
					details = string(b)
				}
			}
			_ = cw.Write([]string{
				e.ID.String(), e.ActionType, safeStr(e.ResourceType), safeStr(e.ResourceID), safeStr(e.UserID),
				e.Action, details, e.CreatedAt.Format(time.RFC3339),
			})
		}
		cw.Flush()
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"events": events, "count": len(events)})
	}
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
