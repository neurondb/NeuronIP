package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/semantic"
)

/* SemanticHandler handles semantic layer requests */
type SemanticHandler struct {
	semanticService  *semantic.Service
	approvalService  *semantic.ApprovalService
	ownershipService *semantic.MetricOwnershipService
	lineageService   *semantic.LineageService
}

/* NewSemanticHandler creates a new semantic handler */
func NewSemanticHandler(
	semanticService *semantic.Service,
	approvalService *semantic.ApprovalService,
	ownershipService *semantic.MetricOwnershipService,
	lineageService *semantic.LineageService,
) *SemanticHandler {
	return &SemanticHandler{
		semanticService:  semanticService,
		approvalService:  approvalService,
		ownershipService: ownershipService,
		lineageService:   lineageService,
	}
}

/* GetTimeGrains handles GET /api/v1/metrics/time-grains */
func (h *SemanticHandler) GetTimeGrains(w http.ResponseWriter, r *http.Request) {
	grains, err := h.semanticService.GetTimeGrains(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grains)
}

/* GetMetricTimeGrains handles GET /api/v1/metrics/{id}/time-grains */
func (h *SemanticHandler) GetMetricTimeGrains(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	grains, err := h.semanticService.GetMetricTimeGrains(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grains)
}

/* AddMetricTimeGrain handles POST /api/v1/metrics/{id}/time-grains */
func (h *SemanticHandler) AddMetricTimeGrain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req struct {
		TimeGrainID uuid.UUID `json:"time_grain_id"`
		IsDefault   bool      `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	err = h.semanticService.AddMetricTimeGrain(r.Context(), metricID, req.TimeGrainID, req.IsDefault)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

/* DeleteMetricTimeGrain handles DELETE /api/v1/metrics/{id}/time-grains/{time_grain_id} */
func (h *SemanticHandler) DeleteMetricTimeGrain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	timeGrainID, err := uuid.Parse(vars["time_grain_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid time grain ID"))
		return
	}

	err = h.semanticService.DeleteMetricTimeGrain(r.Context(), metricID, timeGrainID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* GetMetricFilters handles GET /api/v1/metrics/{id}/filters */
func (h *SemanticHandler) GetMetricFilters(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	filters, err := h.semanticService.GetMetricFilters(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filters)
}

/* AddMetricFilter handles POST /api/v1/metrics/{id}/filters */
func (h *SemanticHandler) AddMetricFilter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req struct {
		FilterExpression string  `json:"filter_expression"`
		FilterName       *string `json:"filter_name,omitempty"`
		IsDefault        bool    `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.FilterExpression == "" {
		WriteErrorResponse(w, errors.ValidationFailed("filter_expression is required", nil))
		return
	}

	err = h.semanticService.AddMetricFilter(r.Context(), metricID, req.FilterExpression, req.FilterName, req.IsDefault)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

/* GetMetricGlossaryTerms handles GET /api/v1/metrics/{id}/glossary */
func (h *SemanticHandler) GetMetricGlossaryTerms(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	terms, err := h.semanticService.GetMetricGlossaryTerms(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(terms)
}

/* LinkMetricToGlossary handles POST /api/v1/metrics/{id}/glossary */
func (h *SemanticHandler) LinkMetricToGlossary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req struct {
		GlossaryTermID uuid.UUID `json:"glossary_term_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	err = h.semanticService.LinkMetricToGlossary(r.Context(), metricID, req.GlossaryTermID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

/* UnlinkMetricFromGlossary handles DELETE /api/v1/metrics/{id}/glossary/{term_id} */
func (h *SemanticHandler) UnlinkMetricFromGlossary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	termID, err := uuid.Parse(vars["term_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid glossary term ID"))
		return
	}

	err = h.semanticService.UnlinkMetricFromGlossary(r.Context(), metricID, termID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* GetMetricApprovals handles GET /api/v1/metrics/{id}/approvals */
func (h *SemanticHandler) GetMetricApprovals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	approvals, err := h.approvalService.GetMetricApprovals(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(approvals)
}

/* CreateMetricApproval handles POST /api/v1/metrics/{id}/approvals */
func (h *SemanticHandler) CreateMetricApproval(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req struct {
		ApproverID string  `json:"approver_id"`
		Comments   *string `json:"comments,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.ApproverID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("approver_id is required", nil))
		return
	}

	approval, err := h.approvalService.CreateApproval(r.Context(), metricID, req.ApproverID, req.Comments)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(approval)
}

/* ApproveMetric handles POST /api/v1/metrics/approvals/{id}/approve */
func (h *SemanticHandler) ApproveMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid approval ID"))
		return
	}

	var req struct {
		ApproverID string  `json:"approver_id"`
		Comments   *string `json:"comments,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.ApproverID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("approver_id is required", nil))
		return
	}

	err = h.approvalService.ApproveMetric(r.Context(), approvalID, req.ApproverID, req.Comments)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* RejectMetric handles POST /api/v1/metrics/approvals/{id}/reject */
func (h *SemanticHandler) RejectMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid approval ID"))
		return
	}

	var req struct {
		ApproverID string `json:"approver_id"`
		Comments   string `json:"comments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.ApproverID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("approver_id is required", nil))
		return
	}

	err = h.approvalService.RejectMetric(r.Context(), approvalID, req.ApproverID, req.Comments)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* RequestChanges handles POST /api/v1/metrics/approvals/{id}/request-changes */
func (h *SemanticHandler) RequestChanges(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid approval ID"))
		return
	}

	var req struct {
		ApproverID string `json:"approver_id"`
		Comments   string `json:"comments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.ApproverID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("approver_id is required", nil))
		return
	}

	err = h.approvalService.RequestChanges(r.Context(), approvalID, req.ApproverID, req.Comments)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* GetMetricLineage handles GET /api/v1/metrics/{id}/lineage */
func (h *SemanticHandler) GetMetricLineage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	maxDepth := 3
	if depthStr := r.URL.Query().Get("max_depth"); depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil {
			maxDepth = d
		}
	}

	lineage, err := h.lineageService.GetMetricLineage(r.Context(), id, maxDepth)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lineage)
}

/* GetImpactAnalysis handles GET /api/v1/metrics/{id}/impact-analysis */
func (h *SemanticHandler) GetImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	impact, err := h.lineageService.GetImpactAnalysis(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(impact)
}

/* UpdateMetricOwner handles PUT /api/v1/metrics/{id}/owner */
func (h *SemanticHandler) UpdateMetricOwner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req struct {
		OwnerID string `json:"owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.OwnerID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("owner_id is required", nil))
		return
	}

	err = h.ownershipService.UpdateMetricOwner(r.Context(), id, req.OwnerID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* GetApprovalQueue handles GET /api/v1/metrics/approvals/queue */
func (h *SemanticHandler) GetApprovalQueue(w http.ResponseWriter, r *http.Request) {
	approverID := r.URL.Query().Get("approver_id")
	var approverIDPtr *string
	if approverID != "" {
		approverIDPtr = &approverID
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	queue, err := h.approvalService.GetApprovalQueue(r.Context(), approverIDPtr, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(queue)
}

/* GetMetricOwners handles GET /api/v1/metrics/{id}/owners */
func (h *SemanticHandler) GetMetricOwners(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	owners, err := h.ownershipService.GetMetricOwners(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(owners)
}

/* AddMetricOwner handles POST /api/v1/metrics/{id}/owners */
func (h *SemanticHandler) AddMetricOwner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req struct {
		OwnerID   string `json:"owner_id"`
		OwnerType string `json:"owner_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.OwnerID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("owner_id is required", nil))
		return
	}

	if req.OwnerType == "" {
		req.OwnerType = "secondary"
	}

	err = h.ownershipService.AddMetricOwner(r.Context(), id, req.OwnerID, req.OwnerType)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

/* RemoveMetricOwner handles DELETE /api/v1/metrics/{id}/owners/{owner_id} */
func (h *SemanticHandler) RemoveMetricOwner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	ownerID := vars["owner_id"]
	if ownerID == "" {
		WriteErrorResponse(w, errors.BadRequest("Invalid owner ID"))
		return
	}

	err = h.ownershipService.RemoveMetricOwner(r.Context(), id, ownerID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
