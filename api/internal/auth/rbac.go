package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* Permission represents a permission for RBAC */
type Permission string

const (
	PermissionSemanticSearch   Permission = "semantic:search"
	PermissionSemanticCreate   Permission = "semantic:create"
	PermissionWarehouseQuery   Permission = "warehouse:query"
	PermissionSupportRead      Permission = "support:read"
	PermissionSupportWrite     Permission = "support:write"
	PermissionWorkflowExecute  Permission = "workflow:execute"
	PermissionWorkflowManage   Permission = "workflow:manage"
	PermissionComplianceRead   Permission = "compliance:read"
	PermissionComplianceManage Permission = "compliance:manage"
	PermissionAdmin            Permission = "admin:*"
)

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
	roleName, err := s.GetUserRole(ctx, userID)
	if err != nil {
		return false, err
	}

	role, exists := RoleRegistry[roleName]
	if !exists {
		return false, fmt.Errorf("role not found: %s", roleName)
	}

	// Check if role has permission
	for _, perm := range role.Permissions {
		if perm == permission || perm == PermissionAdmin {
			return true, nil
		}
		// Check wildcard permissions
		if strings.HasSuffix(string(perm), ":*") {
			permPrefix := strings.TrimSuffix(string(perm), ":*")
			if strings.HasPrefix(string(permission), permPrefix+":") {
				return true, nil
			}
		}
	}

	return false, nil
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
