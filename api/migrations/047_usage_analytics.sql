-- Migration: 047_usage_analytics.sql
-- Description: Usage analytics and cost controls (extends existing usage_metrics and cost_tracking)

-- Resource quotas (enhanced)
CREATE TABLE IF NOT EXISTS neuronip.resource_quotas_enhanced (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID,
    user_id VARCHAR(255),
    resource_type VARCHAR(50) NOT NULL, -- 'compute', 'storage', 'api_calls', 'queries', 'agents'
    quota_limit BIGINT NOT NULL,
    current_usage BIGINT DEFAULT 0,
    period_type VARCHAR(20) NOT NULL DEFAULT 'monthly', -- 'hourly', 'daily', 'monthly'
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    auto_reset BOOLEAN NOT NULL DEFAULT true,
    enforcement_action VARCHAR(50) NOT NULL DEFAULT 'block', -- 'block', 'throttle', 'warn'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, user_id, resource_type, period_start)
);

CREATE INDEX idx_resource_quotas_workspace ON neuronip.resource_quotas_enhanced(workspace_id);
CREATE INDEX idx_resource_quotas_user ON neuronip.resource_quotas_enhanced(user_id);
CREATE INDEX idx_resource_quotas_type ON neuronip.resource_quotas_enhanced(resource_type);
CREATE INDEX idx_resource_quotas_period ON neuronip.resource_quotas_enhanced(period_start, period_end);

-- Budget alerts
CREATE TABLE IF NOT EXISTS neuronip.budget_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID,
    user_id VARCHAR(255),
    budget_name VARCHAR(255) NOT NULL,
    budget_amount DOUBLE PRECISION NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    alert_thresholds JSONB NOT NULL, -- e.g., {"50": "warning", "80": "critical", "100": "block"}
    current_spend DOUBLE PRECISION DEFAULT 0,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budget_alerts_workspace ON neuronip.budget_alerts(workspace_id);
CREATE INDEX idx_budget_alerts_user ON neuronip.budget_alerts(user_id);
CREATE INDEX idx_budget_alerts_enabled ON neuronip.budget_alerts(enabled) WHERE enabled = true;

-- Cost optimization recommendations
CREATE TABLE IF NOT EXISTS neuronip.cost_optimization_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID,
    user_id VARCHAR(255),
    recommendation_type VARCHAR(50) NOT NULL, -- 'rightsize', 'schedule', 'cache', 'archive'
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255),
    current_cost DOUBLE PRECISION,
    potential_savings DOUBLE PRECISION,
    recommendation_text TEXT NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium', -- 'low', 'medium', 'high'
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'applied', 'dismissed'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cost_optimization_workspace ON neuronip.cost_optimization_recommendations(workspace_id);
CREATE INDEX idx_cost_optimization_status ON neuronip.cost_optimization_recommendations(status);
CREATE INDEX idx_cost_optimization_priority ON neuronip.cost_optimization_recommendations(priority);

COMMENT ON TABLE neuronip.resource_quotas_enhanced IS 'Enhanced resource quota management';
COMMENT ON TABLE neuronip.budget_alerts IS 'Budget tracking and alerts';
COMMENT ON TABLE neuronip.cost_optimization_recommendations IS 'Cost optimization recommendations';
