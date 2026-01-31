-- Migration: 057_budgets.sql
-- Description: Budget management for cost controls

-- Budgets table
CREATE TABLE IF NOT EXISTS neuronip.budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_name VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    limit_amount DOUBLE PRECISION NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    allocation JSONB, -- Budget allocation by team/project
    alerts_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budgets_user ON neuronip.budgets(user_id);
CREATE INDEX idx_budgets_period ON neuronip.budgets(period_start, period_end);

-- Integration marketplace
CREATE TABLE IF NOT EXISTS neuronip.integration_marketplace (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    description TEXT,
    version VARCHAR(50) NOT NULL,
    rating_average DOUBLE PRECISION DEFAULT 0,
    rating_count INTEGER DEFAULT 0,
    install_count INTEGER DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'draft', -- 'draft', 'published', 'deprecated'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integration_marketplace_category ON neuronip.integration_marketplace(category);
CREATE INDEX idx_integration_marketplace_status ON neuronip.integration_marketplace(status) WHERE status = 'published';

-- Integration installations
CREATE TABLE IF NOT EXISTS neuronip.integration_installations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES neuronip.integration_marketplace(id),
    user_id VARCHAR(255) NOT NULL,
    config JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'inactive', 'error'
    installed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integration_installations_integration ON neuronip.integration_installations(integration_id);
CREATE INDEX idx_integration_installations_user ON neuronip.integration_installations(user_id);

COMMENT ON TABLE neuronip.budgets IS 'Budget definitions for cost control';
COMMENT ON TABLE neuronip.integration_marketplace IS 'Integration marketplace catalog';
COMMENT ON TABLE neuronip.integration_installations IS 'Installed integrations';
