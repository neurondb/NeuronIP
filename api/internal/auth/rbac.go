package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

// Permission type and constants are defined in permissions.go

/* Role represents a user role with permissions */
type Role struct {
	Name        string
	Permissions []Permission
}

/* Default roles with permissions */
var (
	RoleAdmin = Role{
		Name:        "admin",
		Permissions: []Permission{PermissionAdmin},
	}

	RoleAnalyst = Role{
		Name: "analyst",
		Permissions: []Permission{
			PermissionSemanticSearch,
			PermissionWarehouseQuery,
			PermissionSupportRead,
			PermissionComplianceRead,
		},
	}

	RoleSupport = Role{
		Name: "support",
		Permissions: []Permission{
			PermissionSemanticSearch,
			PermissionSupportRead,
			PermissionSupportWrite,
		},
	}

	RoleDeveloper = Role{
		Name: "developer",
		Permissions: []Permission{
			PermissionSemanticSearch,
			PermissionSemanticCreate,
			PermissionWarehouseQuery,
			PermissionWorkflowExecute,
		},
	}
)

/* RoleRegistry maps role names to Role definitions */
var RoleRegistry = map[string]Role{
	"admin":     RoleAdmin,
	"analyst":   RoleAnalyst,
	"support":   RoleSupport,
	"developer": RoleDeveloper,
}

/* RBACService provides role-based access control */
type RBACService struct {
	queries *db.Queries
}

/* NewRBACService creates a new RBAC service */
func NewRBACService(queries *db.Queries) *RBACService {
	return &RBACService{queries: queries}
}

/* GetUserRole gets the role for a user from database */
func (s *RBACService) GetUserRole(ctx context.Context, userID string) (string, error) {
	// First, try to query user_roles table if it exists
	query := `
		SELECT role_name
		FROM neuronip.user_roles
		WHERE user_id = $1
		LIMIT 1`
	
	var role string
	err := s.queries.DB.QueryRow(ctx, query, userID).Scan(&role)
	if err == nil {
		return role, nil
	}
	
	// If user_roles table doesn't exist or no role found, try api_keys table
	// API keys may have role information
	apiKeyQuery := `
		SELECT role
		FROM neuronip.api_keys
		WHERE user_id = $1 AND role IS NOT NULL
		LIMIT 1`
	
	err = s.queries.DB.QueryRow(ctx, apiKeyQuery, userID).Scan(&role)
	if err == nil {
		return role, nil
	}
	
	// Default to "analyst" role if no role found
	return "analyst", nil
}

/* HasPermission checks if a user has a specific permission */
func (s *RBACService) HasPermission(ctx context.Context, userID string, permission Permission) (bool, error) {
	// Check admin permission first
	roleName, err := s.GetUserRole(ctx, userID)
	if err == nil {
		role, exists := RoleRegistry[roleName]
		if exists {
			for _, perm := range role.Permissions {
				if perm == PermissionAdmin {
					return true, nil
				}
			}
		}
	}

	// Check role-based permissions
	if roleName != "" {
		role, exists := RoleRegistry[roleName]
		if exists {
			for _, perm := range role.Permissions {
				if HasPermission(perm, permission) {
					return true, nil
				}
			}
		}
	}

	// Check custom roles
	customRoles, err := s.GetUserCustomRoles(ctx, userID)
	if err == nil {
		for _, role := range customRoles {
			for _, perm := range role.Permissions {
				if HasPermission(Permission(perm), permission) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

/* HasPermissionWithScope checks if a user has a permission with specific scope */
func (s *RBACService) HasPermissionWithScope(ctx context.Context, userID string, permission Permission, scope PermissionScope) (bool, error) {
	// Check organization-level permission
	if scope.OrganizationID != nil {
		hasOrgPerm, err := s.HasOrganizationPermission(ctx, userID, *scope.OrganizationID, permission)
		if err == nil && hasOrgPerm {
			return true, nil
		}
	}

	// Check workspace-level permission
	if scope.WorkspaceID != nil {
		hasWorkspacePerm, err := s.HasWorkspacePermission(ctx, userID, *scope.WorkspaceID, permission)
		if err == nil && hasWorkspacePerm {
			return true, nil
		}
	}

	// Check resource-level permission
	if scope.ResourceType != nil && scope.ResourceID != nil {
		hasResourcePerm, err := s.HasResourcePermission(ctx, userID, *scope.ResourceType, *scope.ResourceID, permission)
		if err == nil && hasResourcePerm {
			return true, nil
		}
	}

	// Fall back to global permission check
	return s.HasPermission(ctx, userID, permission)
}

/* HasOrganizationPermission checks organization-level permission */
func (s *RBACService) HasOrganizationPermission(ctx context.Context, userID, orgID string, permission Permission) (bool, error) {
	// Check if user is member of organization
	query := `SELECT role FROM neuronip.organization_members WHERE organization_id = $1 AND user_id = $2`
	var role string
	err := s.queries.DB.QueryRow(ctx, query, orgID, userID).Scan(&role)
	if err != nil {
		return false, nil
	}

	// Organization owners and admins have all permissions
	if role == "owner" || role == "admin" {
		return true, nil
	}

	return false, nil
}

/* HasWorkspacePermission checks workspace-level permission */
func (s *RBACService) HasWorkspacePermission(ctx context.Context, userID, workspaceID string, permission Permission) (bool, error) {
	// Check if user is member of workspace
	query := `SELECT role FROM neuronip.workspace_members WHERE workspace_id = $1 AND user_id = $2`
	var role string
	err := s.queries.DB.QueryRow(ctx, query, workspaceID, userID).Scan(&role)
	if err != nil {
		return false, nil
	}

	// Workspace admins have all permissions
	if role == "admin" {
		return true, nil
	}

	return false, nil
}

/* HasResourcePermission checks resource-level permission */
func (s *RBACService) HasResourcePermission(ctx context.Context, userID, resourceType, resourceID string, permission Permission) (bool, error) {
	query := `
		SELECT COUNT(*) FROM neuronip.resource_permissions
		WHERE user_id = $1 AND resource_type = $2 AND resource_id = $3 AND permission = $4
	`
	var count int
	err := s.queries.DB.QueryRow(ctx, query, userID, resourceType, resourceID, string(permission)).Scan(&count)
	if err != nil {
		return false, nil
	}
	return count > 0, nil
}

/* CustomRole represents a custom role */
type CustomRole struct {
	ID             uuid.UUID
	OrganizationID *uuid.UUID
	WorkspaceID    *uuid.UUID
	Name           string
	Description    *string
	Permissions    []string
}

/* GetUserCustomRoles gets custom roles assigned to a user */
func (s *RBACService) GetUserCustomRoles(ctx context.Context, userID string) ([]CustomRole, error) {
	query := `
		SELECT cr.id, cr.organization_id, cr.workspace_id, cr.name, cr.description, cr.permissions
		FROM neuronip.custom_roles cr
		JOIN neuronip.role_assignments ra ON cr.id = ra.role_id
		WHERE ra.user_id = $1
	`
	rows, err := s.queries.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []CustomRole
	for rows.Next() {
		var role CustomRole
		err := rows.Scan(&role.ID, &role.OrganizationID, &role.WorkspaceID, &role.Name, &role.Description, &role.Permissions)
		if err != nil {
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}

/* CheckPermissionMiddleware creates middleware that checks permissions */
func CheckPermissionMiddleware(rbac *RBACService, permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "User ID not found in context", http.StatusUnauthorized)
				return
			}

			hasPermission, err := rbac.HasPermission(r.Context(), userID, permission)
			if err != nil {
				http.Error(w, fmt.Sprintf("Permission check failed: %v", err), http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

/* RequirePermission is a helper to check permissions in handlers */
func RequirePermission(ctx context.Context, rbac *RBACService, permission Permission) error {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("user ID not found in context")
	}

	hasPermission, err := rbac.HasPermission(ctx, userID, permission)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}

	if !hasPermission {
		return fmt.Errorf("insufficient permissions: requires %s", permission)
	}

	return nil
}
