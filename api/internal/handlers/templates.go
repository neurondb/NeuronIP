package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/workflows"
)

/* TemplateHandler handles workflow template requests */
type TemplateHandler struct {
	templateService *workflows.TemplateService
}

/* NewTemplateHandler creates a new template handler */
func NewTemplateHandler(templateService *workflows.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

/* CreateTemplate handles POST /api/v1/workflows/templates */
func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var template workflows.WorkflowTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.templateService.CreateTemplate(r.Context(), template)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetTemplate handles GET /api/v1/workflows/templates/{id} */
func (h *TemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid template ID"))
		return
	}

	template, err := h.templateService.GetTemplate(r.Context(), templateID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

/* ListTemplates handles GET /api/v1/workflows/templates */
func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	isPublicStr := r.URL.Query().Get("is_public")
	limit := 50

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	var isPublic *bool
	if isPublicStr != "" {
		if parsed, err := strconv.ParseBool(isPublicStr); err == nil {
			isPublic = &parsed
		}
	}

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	templates, err := h.templateService.ListTemplates(r.Context(), categoryPtr, isPublic, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

/* InstantiateTemplate handles POST /api/v1/workflows/templates/{id}/instantiate */
func (h *TemplateHandler) InstantiateTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid template ID"))
		return
	}

	var req struct {
		Name       string                 `json:"name"`
		Parameters map[string]interface{} `json:"parameters,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	definition, err := h.templateService.InstantiateTemplate(r.Context(), templateID, req.Name, req.Parameters)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(definition)
}
