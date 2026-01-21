-- Migration: RBAC Enhancements
-- Description: Adds organization, workspace, and resource-level permissions

-- Organizations: Top-level organizational units
CREATE TABLE IF NOT EXISTS neuronip.organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.organizations IS 'Organizations for multi-tenant RBAC';

CREATE INDEX IF NOT EXISTS idx_organizations_name ON neuronip.organizations(name);

-- Workspaces: Workspaces within organizations
CREATE TABLE IF NOT EXISTS neuronip.workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES neuronip.organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, name)
);
COMMENT ON TABLE neuronip.workspaces IS 'Workspaces within organizations';

CREATE INDEX IF NOT EXISTS idx_workspaces_org ON neuronip.workspaces(organization_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_name ON neuronip.workspaces(name);

-- Organization members: Users in organizations
CREATE TABLE IF NOT EXISTS neuronip.organization_members (
    organization_id UUID NOT NULL REFERENCES neuronip.organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);
COMMENT ON TABLE neuronip.organization_members IS 'Organization membership and roles';

CREATE INDEX IF NOT EXISTS idx_org_members_user ON neuronip.organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_org ON neuronip.organization_members(organization_id);

-- Workspace members: Users in workspaces
CREATE TABLE IF NOT EXISTS neuronip.workspace_members (
    workspace_id UUID NOT NULL REFERENCES neuronip.workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'member', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);
COMMENT ON TABLE neuronip.workspace_members IS 'Workspace membership and roles';

CREATE INDEX IF NOT EXISTS idx_workspace_members_user ON neuronip.workspace_members(user_id);
CREATE INDEX IF NOT EXISTS idx_workspace_members_workspace ON neuronip.workspace_members(workspace_id);

-- Resource permissions: Fine-grained resource-level permissions
CREATE TABLE IF NOT EXISTS neuronip.resource_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES neuronip.organizations(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES neuronip.workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES neuronip.users(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL, -- e.g., 'dataset', 'agent', 'workflow'
    resource_id UUID NOT NULL,
    permission TEXT NOT NULL, -- e.g., 'read', 'write', 'execute'
    granted_by UUID REFERENCES neuronip.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, workspace_id, user_id, resource_type, resource_id, permission)
);
COMMENT ON TABLE neuronip.resource_permissions IS 'Fine-grained resource-level permissions';

CREATE INDEX IF NOT EXISTS idx_resource_perms_user ON neuronip.resource_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_resource_perms_resource ON neuronip.resource_permissions(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_resource_perms_org ON neuronip.resource_permissions(organization_id);
CREATE INDEX IF NOT EXISTS idx_resource_perms_workspace ON neuronip.resource_permissions(workspace_id);

-- Custom roles: Custom roles with fine-grained permissions
CREATE TABLE IF NOT EXISTS neuronip.custom_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES neuronip.organizations(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES neuronip.workspaces(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    permissions TEXT[] NOT NULL, -- Array of permission strings
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(COALESCE(organization_id, '00000000-0000-0000-0000-000000000000'::uuid), 
           COALESCE(workspace_id, '00000000-0000-0000-0000-000000000000'::uuid), 
           name)
);
COMMENT ON TABLE neuronip.custom_roles IS 'Custom roles with fine-grained permissions';

CREATE INDEX IF NOT EXISTS idx_custom_roles_org ON neuronip.custom_roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_custom_roles_workspace ON neuronip.custom_roles(workspace_id);

-- Role assignments: Assign custom roles to users
CREATE TABLE IF NOT EXISTS neuronip.role_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES neuronip.custom_roles(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES neuronip.organizations(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES neuronip.workspaces(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES neuronip.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, role_id, COALESCE(organization_id, '00000000-0000-0000-0000-000000000000'::uuid), 
           COALESCE(workspace_id, '00000000-0000-0000-0000-000000000000'::uuid))
);
COMMENT ON TABLE neuronip.role_assignments IS 'Custom role assignments to users';

CREATE INDEX IF NOT EXISTS idx_role_assignments_user ON neuronip.role_assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_role_assignments_role ON neuronip.role_assignments(role_id);

-- Permission inheritance: Track permission inheritance relationships
CREATE TABLE IF NOT EXISTS neuronip.permission_inheritance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_permission_id UUID REFERENCES neuronip.resource_permissions(id) ON DELETE CASCADE,
    child_permission_id UUID REFERENCES neuronip.resource_permissions(id) ON DELETE CASCADE,
    inheritance_type TEXT NOT NULL CHECK (inheritance_type IN ('delegation', 'inheritance')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(parent_permission_id, child_permission_id)
);
COMMENT ON TABLE neuronip.permission_inheritance IS 'Permission inheritance and delegation';

-- Update triggers
CREATE OR REPLACE FUNCTION neuronip.update_organizations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_organizations_updated_at ON neuronip.organizations;
CREATE TRIGGER trigger_update_organizations_updated_at
    BEFORE UPDATE ON neuronip.organizations
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_organizations_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_workspaces_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_workspaces_updated_at ON neuronip.workspaces;
CREATE TRIGGER trigger_update_workspaces_updated_at
    BEFORE UPDATE ON neuronip.workspaces
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_workspaces_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_custom_roles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_custom_roles_updated_at ON neuronip.custom_roles;
CREATE TRIGGER trigger_update_custom_roles_updated_at
    BEFORE UPDATE ON neuronip.custom_roles
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_custom_roles_updated_at();
