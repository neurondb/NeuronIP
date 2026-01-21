package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* RBACHandler handles RBAC requests */
type RBACHandler struct {
	rbacService      *auth.RBACService
	workspaceService *auth.WorkspaceService
}

/* NewRBACHandler creates a new RBAC handler */
func NewRBACHandler(rbacService *auth.RBACService, workspaceService *auth.WorkspaceService) *RBACHandler {
	return &RBACHandler{
		rbacService:      rbacService,
		workspaceService: workspaceService,
	}
}

/* CreateWorkspace handles creating a workspace */
func (h *RBACHandler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrganizationID uuid.UUID `json:"organization_id"`
		Name           string    `json:"name"`
		Description    *string   `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	workspace, err := h.workspaceService.CreateWorkspace(r.Context(), req.OrganizationID, req.Name, req.Description)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workspace)
}

/* ListWorkspaces handles listing workspaces */
func (h *RBACHandler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("organization_id")
	if orgIDStr == "" {
		WriteErrorResponse(w, errors.BadRequest("organization_id is required"))
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid organization ID"))
		return
	}

	workspaces, err := h.workspaceService.ListWorkspaces(r.Context(), orgID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaces)
}

/* CheckPermission handles permission checks */
func (h *RBACHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	var req struct {
		Permission string                `json:"permission"`
		Scope      auth.PermissionScope  `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	hasPermission, err := h.rbacService.HasPermissionWithScope(
		r.Context(),
		userIDStr,
		auth.Permission(req.Permission),
		req.Scope,
	)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"has_permission": hasPermission,
	})
}
