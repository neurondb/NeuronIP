-- Migration: Multi-Tenancy Schema
-- Description: Adds tables for multi-tenant isolation (per-tenant schema or database)

-- Tenants: Tenant metadata and configuration
CREATE TABLE IF NOT EXISTS neuronip.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    schema_name TEXT, -- For schema-per-tenant mode
    database_name TEXT, -- For database-per-tenant mode
    isolated BOOLEAN NOT NULL DEFAULT true,
    tenancy_mode TEXT NOT NULL CHECK (tenancy_mode IN ('schema', 'database')) DEFAULT 'schema',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.tenants IS 'Multi-tenant isolation configuration';

CREATE INDEX IF NOT EXISTS idx_tenants_name ON neuronip.tenants(name);
CREATE INDEX IF NOT EXISTS idx_tenants_schema ON neuronip.tenants(schema_name);
CREATE INDEX IF NOT EXISTS idx_tenants_database ON neuronip.tenants(database_name);

-- Tenant users: Map users to tenants
CREATE TABLE IF NOT EXISTS neuronip.tenant_users (
    tenant_id UUID NOT NULL REFERENCES neuronip.tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'member', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, user_id)
);
COMMENT ON TABLE neuronip.tenant_users IS 'User-tenant associations';

CREATE INDEX IF NOT EXISTS idx_tenant_users_user ON neuronip.tenant_users(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant ON neuronip.tenant_users(tenant_id);

-- Tenant resources: Track resources per tenant
CREATE TABLE IF NOT EXISTS neuronip.tenant_resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES neuronip.tenants(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL, -- e.g., 'dataset', 'agent', 'workflow'
    resource_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, resource_type, resource_id)
);
COMMENT ON TABLE neuronip.tenant_resources IS 'Resource-tenant associations for isolation';

CREATE INDEX IF NOT EXISTS idx_tenant_resources_tenant ON neuronip.tenant_resources(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_resources_resource ON neuronip.tenant_resources(resource_type, resource_id);

-- Isolation audit: Track isolation verification tests
CREATE TABLE IF NOT EXISTS neuronip.isolation_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES neuronip.tenants(id) ON DELETE CASCADE,
    test_type TEXT NOT NULL, -- e.g., 'data_isolation', 'schema_isolation', 'access_isolation'
    status TEXT NOT NULL CHECK (status IN ('passed', 'failed', 'warning')),
    details JSONB,
    tested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.isolation_audit IS 'Isolation verification test results';

CREATE INDEX IF NOT EXISTS idx_isolation_audit_tenant ON neuronip.isolation_audit(tenant_id);
CREATE INDEX IF NOT EXISTS idx_isolation_audit_status ON neuronip.isolation_audit(status);
CREATE INDEX IF NOT EXISTS idx_isolation_audit_tested ON neuronip.isolation_audit(tested_at DESC);

-- Update trigger
CREATE OR REPLACE FUNCTION neuronip.update_tenants_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_tenants_updated_at ON neuronip.tenants;
CREATE TRIGGER trigger_update_tenants_updated_at
    BEFORE UPDATE ON neuronip.tenants
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_tenants_updated_at();
