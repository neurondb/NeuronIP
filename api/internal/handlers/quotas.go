package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/tenancy"
)

/* QuotaHandler handles resource quota requests */
type QuotaHandler struct {
	quotaService *tenancy.QuotaService
}

/* NewQuotaHandler creates a new quota handler */
func NewQuotaHandler(pool *pgxpool.Pool) *QuotaHandler {
	return &QuotaHandler{
		quotaService: tenancy.NewQuotaService(pool),
	}
}

/* SetQuota handles POST /api/v1/quotas */
func (h *QuotaHandler) SetQuota(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkspaceID  *uuid.UUID `json:"workspace_id,omitempty"`
		UserID       *string    `json:"user_id,omitempty"`
		ResourceType string     `json:"resource_type"`
		Limit        int64      `json:"limit"`
		Period       string     `json:"period"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.ResourceType == "" || req.Limit <= 0 || req.Period == "" {
		WriteErrorResponse(w, errors.ValidationFailed("resource_type, limit, and period are required", nil))
		return
	}

	quota, err := h.quotaService.SetQuota(r.Context(), req.WorkspaceID, req.UserID, req.ResourceType, req.Limit, req.Period)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(quota)
}

/* ListQuotas handles GET /api/v1/quotas */
func (h *QuotaHandler) ListQuotas(w http.ResponseWriter, r *http.Request) {
	var workspaceID *uuid.UUID
	if wsIDStr := r.URL.Query().Get("workspace_id"); wsIDStr != "" {
		if id, err := uuid.Parse(wsIDStr); err == nil {
			workspaceID = &id
		}
	}

	userID := r.URL.Query().Get("user_id")
	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	quotas, err := h.quotaService.ListQuotas(r.Context(), workspaceID, userIDPtr)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotas)
}

/* CheckQuota handles POST /api/v1/quotas/check */
func (h *QuotaHandler) CheckQuota(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkspaceID     *uuid.UUID `json:"workspace_id,omitempty"`
		UserID          *string    `json:"user_id,omitempty"`
		ResourceType    string     `json:"resource_type"`
		RequestedAmount int64      `json:"requested_amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.ResourceType == "" {
		WriteErrorResponse(w, errors.ValidationFailed("resource_type is required", nil))
		return
	}

	withinQuota, quota, err := h.quotaService.CheckQuota(r.Context(), req.WorkspaceID, req.UserID, req.ResourceType, req.RequestedAmount)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"within_quota": withinQuota,
		"quota":        quota,
	})
}
