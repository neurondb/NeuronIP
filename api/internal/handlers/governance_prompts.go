package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/governance"
)

/* PromptTemplateHandler handles prompt template governance requests */
type PromptTemplateHandler struct {
	service *governance.PromptTemplateService
}

/* NewPromptTemplateHandler creates a new prompt template handler */
func NewPromptTemplateHandler(service *governance.PromptTemplateService) *PromptTemplateHandler {
	return &PromptTemplateHandler{service: service}
}

/* CreatePromptTemplate handles POST /api/v1/governance/prompts */
func (h *PromptTemplateHandler) CreatePromptTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name             string                 `json:"name"`
		Version          string                 `json:"version"`
		TemplateText     string                 `json:"template_text"`
		Variables        []string               `json:"variables"`
		Description      *string                `json:"description,omitempty"`
		ParentTemplateID *uuid.UUID             `json:"parent_template_id,omitempty"`
		Metadata         map[string]interface{} `json:"metadata,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		WriteErrorResponse(w, errors.Unauthorized("User ID required"))
		return
	}

	prompt, err := h.service.CreatePromptTemplate(r.Context(), req.Name, req.Version, req.TemplateText, req.Variables, req.Description, userID, req.ParentTemplateID, req.Metadata)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, prompt)
}

/* GetPromptTemplate handles GET /api/v1/governance/prompts/{id} */
func (h *PromptTemplateHandler) GetPromptTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid prompt ID"))
		return
	}
	prompt, err := h.service.GetPromptTemplate(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, prompt)
}

/* ListPromptTemplates handles GET /api/v1/governance/prompts */
func (h *PromptTemplateHandler) ListPromptTemplates(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	prompts, err := h.service.ListPrompts(r.Context(), limit)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, prompts)
}

/* ApprovePromptTemplate handles POST /api/v1/governance/prompts/{id}/approve */
func (h *PromptTemplateHandler) ApprovePromptTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid prompt ID"))
		return
	}
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		WriteErrorResponse(w, errors.Unauthorized("User ID required"))
		return
	}
	if err := h.service.ApprovePrompt(r.Context(), id, userID); err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

/* ApprovalWorkflowHandler handles approval workflow requests */
type ApprovalWorkflowHandler struct {
	service *governance.ApprovalWorkflowService
}

/* NewApprovalWorkflowHandler creates a new approval workflow handler */
func NewApprovalWorkflowHandler(service *governance.ApprovalWorkflowService) *ApprovalWorkflowHandler {
	return &ApprovalWorkflowHandler{service: service}
}

/* CreateWorkflow handles POST /api/v1/governance/approvals */
func (h *ApprovalWorkflowHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var wf governance.ApprovalWorkflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	result, err := h.service.CreateWorkflow(r.Context(), wf)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, result)
}

/* GetWorkflow handles GET /api/v1/governance/approvals/{id} */
func (h *ApprovalWorkflowHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}
	wf, err := h.service.GetWorkflow(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, wf)
}

/* SubmitApproval handles POST /api/v1/governance/approvals/{id}/submit */
func (h *ApprovalWorkflowHandler) SubmitApproval(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}
	var req struct {
		StageNumber int    `json:"stage_number"`
		Decision    string `json:"decision"`
		Comments    string `json:"comments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		WriteErrorResponse(w, errors.Unauthorized("User ID required"))
		return
	}
	if err := h.service.SubmitApproval(r.Context(), workflowID, req.StageNumber, userID, req.Decision, req.Comments); err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
}

/* UIRLSHandler handles UI-driven RLS policy requests */
type UIRLSHandler struct {
	service *governance.UIRLSService
}

/* NewUIRLSHandler creates a new UI RLS handler */
func NewUIRLSHandler(service *governance.UIRLSService) *UIRLSHandler {
	return &UIRLSHandler{service: service}
}

/* CreateRLSPolicy handles POST /api/v1/governance/ui-rls/policies */
func (h *UIRLSHandler) CreateRLSPolicy(w http.ResponseWriter, r *http.Request) {
	var policy governance.RLSPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		WriteErrorResponse(w, errors.Unauthorized("User ID required"))
		return
	}
	policy.CreatedBy = userID
	result, err := h.service.CreateRLSPolicy(r.Context(), policy)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, result)
}

/* GetRLSPolicies handles GET /api/v1/governance/ui-rls/policies */
func (h *UIRLSHandler) GetRLSPolicies(w http.ResponseWriter, r *http.Request) {
	schema := r.URL.Query().Get("schema")
	table := r.URL.Query().Get("table")
	if schema == "" || table == "" {
		WriteErrorResponse(w, errors.BadRequest("schema and table query params required"))
		return
	}
	policies, err := h.service.GetRLSPolicies(r.Context(), schema, table)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, policies)
}

/* ToggleRLSPolicy handles POST /api/v1/governance/ui-rls/policies/{id}/toggle */
func (h *UIRLSHandler) ToggleRLSPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid policy ID"))
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if err := h.service.TogglePolicy(r.Context(), id, req.Enabled); err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"enabled": req.Enabled})
}
