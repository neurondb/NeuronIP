-- Migration: Agent Hub
-- Description: Adds agent templates, tools registry, memory policies

-- Agent templates: Pre-defined agent templates
CREATE TABLE IF NOT EXISTS neuronip.agent_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT, -- e.g., 'support', 'analytics', 'automation'
    template_config JSONB NOT NULL, -- Agent configuration
    tools TEXT[], -- Required tools
    memory_policy_id UUID REFERENCES neuronip.memory_policies(id) ON DELETE SET NULL,
    is_public BOOLEAN NOT NULL DEFAULT true,
    usage_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_templates IS 'Pre-defined agent templates';

CREATE INDEX IF NOT EXISTS idx_agent_templates_category ON neuronip.agent_templates(category);
CREATE INDEX IF NOT EXISTS idx_agent_templates_public ON neuronip.agent_templates(is_public);
CREATE INDEX IF NOT EXISTS idx_agent_templates_usage ON neuronip.agent_templates(usage_count DESC);

-- Tools registry: Available tools for agents
CREATE TABLE IF NOT EXISTS neuronip.tools_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    tool_type TEXT NOT NULL CHECK (tool_type IN ('function', 'api', 'database', 'workflow', 'custom')),
    tool_config JSONB NOT NULL, -- Tool configuration
    parameters JSONB, -- Tool parameters schema
    category TEXT,
    tags TEXT[],
    is_available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.tools_registry IS 'Registry of available tools for agents';

CREATE INDEX IF NOT EXISTS idx_tools_registry_name ON neuronip.tools_registry(name);
CREATE INDEX IF NOT EXISTS idx_tools_registry_type ON neuronip.tools_registry(tool_type);
CREATE INDEX IF NOT EXISTS idx_tools_registry_category ON neuronip.tools_registry(category);
CREATE INDEX IF NOT EXISTS idx_tools_registry_tags ON neuronip.tools_registry USING gin(tags);

-- Memory policies: Memory retention and summarization policies
CREATE TABLE IF NOT EXISTS neuronip.memory_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    retention_days INTEGER, -- How long to keep memories
    summarization_enabled BOOLEAN NOT NULL DEFAULT true,
    summarization_threshold INTEGER, -- Number of memories before summarization
    max_memories INTEGER, -- Maximum number of memories
    compression_enabled BOOLEAN NOT NULL DEFAULT false,
    policy_config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.memory_policies IS 'Memory retention and summarization policies';

CREATE INDEX IF NOT EXISTS idx_memory_policies_name ON neuronip.memory_policies(name);

-- Agent tools: Many-to-many relationship between agents and tools
CREATE TABLE IF NOT EXISTS neuronip.agent_tools (
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    tool_id UUID NOT NULL REFERENCES neuronip.tools_registry(id) ON DELETE CASCADE,
    tool_config JSONB, -- Agent-specific tool configuration
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (agent_id, tool_id)
);
COMMENT ON TABLE neuronip.agent_tools IS 'Agent-tool associations';

CREATE INDEX IF NOT EXISTS idx_agent_tools_agent ON neuronip.agent_tools(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_tools_tool ON neuronip.agent_tools(tool_id);

-- Update triggers
CREATE OR REPLACE FUNCTION neuronip.update_agent_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_agent_templates_updated_at ON neuronip.agent_templates;
CREATE TRIGGER trigger_update_agent_templates_updated_at
    BEFORE UPDATE ON neuronip.agent_templates
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_agent_templates_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_tools_registry_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_tools_registry_updated_at ON neuronip.tools_registry;
CREATE TRIGGER trigger_update_tools_registry_updated_at
    BEFORE UPDATE ON neuronip.tools_registry
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_tools_registry_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_memory_policies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_memory_policies_updated_at ON neuronip.memory_policies;
CREATE TRIGGER trigger_update_memory_policies_updated_at
    BEFORE UPDATE ON neuronip.memory_policies
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_memory_policies_updated_at();
