package notion

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/blocks"
	"github.com/neurondb/NeuronIP/api/internal/handlers"
)

/* TemplatesHandler handles page and database template HTTP requests */
type TemplatesHandler struct {
	templates *blocks.TemplatesService
}

/* NewTemplatesHandler creates a new templates handler */
func NewTemplatesHandler(templates *blocks.TemplatesService) *TemplatesHandler {
	return &TemplatesHandler{templates: templates}
}

/* ListPageTemplates handles GET /api/v1/notion-ui/templates/pages */
func (h *TemplatesHandler) ListPageTemplates(w http.ResponseWriter, r *http.Request) {
	var workspaceID *uuid.UUID
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if id, err := uuid.Parse(wsStr); err == nil {
			workspaceID = &id
		}
	}
	list, err := h.templates.ListPageTemplates(r.Context(), workspaceID)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}
	handlers.WriteJSON(w, http.StatusOK, list)
}

/* ListDatabaseTemplates handles GET /api/v1/notion-ui/templates/databases */
func (h *TemplatesHandler) ListDatabaseTemplates(w http.ResponseWriter, r *http.Request) {
	var workspaceID *uuid.UUID
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if id, err := uuid.Parse(wsStr); err == nil {
			workspaceID = &id
		}
	}
	list, err := h.templates.ListDatabaseTemplates(r.Context(), workspaceID)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}
	handlers.WriteJSON(w, http.StatusOK, list)
}
