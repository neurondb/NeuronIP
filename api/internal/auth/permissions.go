package auth

/* Permission represents a fine-grained permission */
type Permission string

/* Permission constants */
const (
	// Semantic search permissions
	PermissionSemanticSearch   Permission = "semantic:search"
	PermissionSemanticCreate   Permission = "semantic:create"
	PermissionSemanticUpdate   Permission = "semantic:update"
	PermissionSemanticDelete   Permission = "semantic:delete"
	PermissionSemanticRead     Permission = "semantic:read"

	// Warehouse permissions
	PermissionWarehouseQuery   Permission = "warehouse:query"
	PermissionWarehouseRead    Permission = "warehouse:read"
	PermissionWarehouseWrite   Permission = "warehouse:write"
	PermissionWarehouseManage  Permission = "warehouse:manage"

	// Support permissions
	PermissionSupportRead      Permission = "support:read"
	PermissionSupportWrite     Permission = "support:write"
	PermissionSupportManage    Permission = "support:manage"

	// Workflow permissions
	PermissionWorkflowExecute  Permission = "workflow:execute"
	PermissionWorkflowRead     Permission = "workflow:read"
	PermissionWorkflowManage   Permission = "workflow:manage"

	// Compliance permissions
	PermissionComplianceRead   Permission = "compliance:read"
	PermissionComplianceManage Permission = "compliance:manage"

	// Agent permissions
	PermissionAgentRead        Permission = "agent:read"
	PermissionAgentExecute     Permission = "agent:execute"
	PermissionAgentManage      Permission = "agent:manage"

	// Data source permissions
	PermissionDataSourceRead   Permission = "datasource:read"
	PermissionDataSourceWrite  Permission = "datasource:write"
	PermissionDataSourceManage Permission = "datasource:manage"

	// Catalog permissions
	PermissionCatalogRead      Permission = "catalog:read"
	PermissionCatalogWrite     Permission = "catalog:write"
	PermissionCatalogManage    Permission = "catalog:manage"

	// Admin permissions
	PermissionAdmin            Permission = "admin:*"
	PermissionUserManage       Permission = "user:manage"
	PermissionSystemManage     Permission = "system:manage"
)

/* PermissionScope represents the scope of a permission */
type PermissionScope struct {
	OrganizationID *string
	WorkspaceID    *string
	ResourceType   *string // e.g., "dataset", "agent", "workflow"
	ResourceID     *string
}

/* HasPermission checks if a permission string matches */
func HasPermission(permission Permission, required Permission) bool {
	if permission == PermissionAdmin {
		return true
	}
	if permission == required {
		return true
	}
	// Check wildcard permissions (e.g., "semantic:*" matches "semantic:read")
	permStr := string(permission)
	reqStr := string(required)
	if len(permStr) > 0 && permStr[len(permStr)-1] == '*' {
		prefix := permStr[:len(permStr)-2] // Remove ":*"
		if len(reqStr) > len(prefix) && reqStr[:len(prefix)+1] == prefix+":" {
			return true
		}
	}
	return false
}

/* ParsePermission parses a permission string */
func ParsePermission(permStr string) Permission {
	return Permission(permStr)
}

/* PermissionSet represents a set of permissions */
type PermissionSet map[Permission]bool

/* NewPermissionSet creates a new permission set */
func NewPermissionSet(permissions ...Permission) PermissionSet {
	set := make(PermissionSet)
	for _, perm := range permissions {
		set[perm] = true
	}
	return set
}

/* Add adds a permission to the set */
func (ps PermissionSet) Add(permission Permission) {
	ps[permission] = true
}

/* Remove removes a permission from the set */
func (ps PermissionSet) Remove(permission Permission) {
	delete(ps, permission)
}

/* Has checks if the set contains a permission */
func (ps PermissionSet) Has(permission Permission) bool {
	if ps[PermissionAdmin] {
		return true
	}
	return ps[permission]
}

/* HasAny checks if the set contains any of the given permissions */
func (ps PermissionSet) HasAny(permissions ...Permission) bool {
	for _, perm := range permissions {
		if ps.Has(perm) {
			return true
		}
	}
	return false
}

/* HasAll checks if the set contains all of the given permissions */
func (ps PermissionSet) HasAll(permissions ...Permission) bool {
	for _, perm := range permissions {
		if !ps.Has(perm) {
			return false
		}
	}
	return true
}

/* List returns all permissions in the set */
func (ps PermissionSet) List() []Permission {
	perms := make([]Permission, 0, len(ps))
	for perm := range ps {
		perms = append(perms, perm)
	}
	return perms
}
