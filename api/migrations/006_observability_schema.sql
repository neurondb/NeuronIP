-- Migration: Observability Schema
-- Description: Adds tables for usage analytics, cost tracking, and alerting

-- Usage metrics: Time-series usage data
CREATE TABLE IF NOT EXISTS neuronip.usage_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT,
    resource_type TEXT NOT NULL CHECK (resource_type IN ('query', 'agent', 'workflow', 'api_call', 'model_inference')),
    resource_id TEXT,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL,
    unit TEXT,
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.usage_metrics IS 'Time-series usage metrics';

CREATE INDEX IF NOT EXISTS idx_usage_metrics_user ON neuronip.usage_metrics(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_metrics_resource ON neuronip.usage_metrics(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_usage_metrics_timestamp ON neuronip.usage_metrics(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_usage_metrics_name ON neuronip.usage_metrics(metric_name);

-- Cost tracking: Cost records
CREATE TABLE IF NOT EXISTS neuronip.cost_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    cost_amount NUMERIC NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    cost_category TEXT CHECK (cost_category IN ('compute', 'storage', 'api_calls', 'model_inference', 'data_transfer', 'other')),
    billing_period_start TIMESTAMPTZ NOT NULL,
    billing_period_end TIMESTAMPTZ NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.cost_tracking IS 'Cost tracking records';

CREATE INDEX IF NOT EXISTS idx_cost_tracking_user ON neuronip.cost_tracking(user_id);
CREATE INDEX IF NOT EXISTS idx_cost_tracking_period ON neuronip.cost_tracking(billing_period_start, billing_period_end);
CREATE INDEX IF NOT EXISTS idx_cost_tracking_category ON neuronip.cost_tracking(cost_category);

-- Alerts: Alert definitions and history
CREATE TABLE IF NOT EXISTS neuronip.alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    alert_type TEXT NOT NULL CHECK (alert_type IN ('threshold', 'anomaly', 'error_rate', 'cost', 'performance')),
    condition JSONB NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('info', 'warning', 'error', 'critical')) DEFAULT 'warning',
    enabled BOOLEAN NOT NULL DEFAULT true,
    notification_channels TEXT[] DEFAULT '{}',
    last_triggered_at TIMESTAMPTZ,
    trigger_count INTEGER DEFAULT 0,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.alerts IS 'Alert definitions';

CREATE INDEX IF NOT EXISTS idx_alerts_type ON neuronip.alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_alerts_enabled ON neuronip.alerts(enabled);

-- Alert history: Alert trigger history
CREATE TABLE IF NOT EXISTS neuronip.alert_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID NOT NULL REFERENCES neuronip.alerts(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('triggered', 'resolved', 'acknowledged')) DEFAULT 'triggered',
    message TEXT NOT NULL,
    context JSONB DEFAULT '{}',
    resolved_at TIMESTAMPTZ,
    resolved_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.alert_history IS 'Alert trigger history';

CREATE INDEX IF NOT EXISTS idx_alert_history_alert ON neuronip.alert_history(alert_id);
CREATE INDEX IF NOT EXISTS idx_alert_history_status ON neuronip.alert_history(status);
CREATE INDEX IF NOT EXISTS idx_alert_history_created_at ON neuronip.alert_history(created_at DESC);
