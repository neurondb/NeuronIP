-- Migration: 052_observability_enhancements.sql
-- Description: Enhanced observability with distributed tracing, advanced metrics, and enterprise alerting

-- Aggregated logs table for centralized log collection
CREATE TABLE IF NOT EXISTS neuronip.aggregated_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level VARCHAR(20) NOT NULL CHECK (level IN ('debug', 'info', 'warn', 'error', 'fatal')),
    message TEXT NOT NULL,
    service VARCHAR(100) NOT NULL,
    component VARCHAR(100),
    trace_id VARCHAR(255),
    span_id VARCHAR(255),
    fields JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_aggregated_logs_level ON neuronip.aggregated_logs(level);
CREATE INDEX idx_aggregated_logs_service ON neuronip.aggregated_logs(service);
CREATE INDEX idx_aggregated_logs_timestamp ON neuronip.aggregated_logs(timestamp DESC);
CREATE INDEX idx_aggregated_logs_trace ON neuronip.aggregated_logs(trace_id) WHERE trace_id IS NOT NULL;

-- Enterprise alert rules
CREATE TABLE IF NOT EXISTS neuronip.enterprise_alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    rule_type VARCHAR(50) NOT NULL CHECK (rule_type IN ('threshold', 'anomaly', 'data_drift', 'custom')),
    metric VARCHAR(255) NOT NULL,
    condition VARCHAR(10) NOT NULL CHECK (condition IN ('gt', 'gte', 'lt', 'lte', 'eq', 'neq')),
    threshold DOUBLE PRECISION,
    enabled BOOLEAN NOT NULL DEFAULT true,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    notification_channels JSONB NOT NULL DEFAULT '[]',
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_enterprise_alert_rules_enabled ON neuronip.enterprise_alert_rules(enabled) WHERE enabled = true;
CREATE INDEX idx_enterprise_alert_rules_severity ON neuronip.enterprise_alert_rules(severity);

-- Enterprise alerts
CREATE TABLE IF NOT EXISTS neuronip.enterprise_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES neuronip.enterprise_alert_rules(id) ON DELETE CASCADE,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    message TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'resolved')),
    acknowledged_by VARCHAR(255),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolution TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_enterprise_alerts_rule ON neuronip.enterprise_alerts(rule_id);
CREATE INDEX idx_enterprise_alerts_status ON neuronip.enterprise_alerts(status);
CREATE INDEX idx_enterprise_alerts_severity ON neuronip.enterprise_alerts(severity);
CREATE INDEX idx_enterprise_alerts_created ON neuronip.enterprise_alerts(created_at DESC);

-- Alert history for deduplication
CREATE TABLE IF NOT EXISTS neuronip.alert_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID NOT NULL REFERENCES neuronip.enterprise_alerts(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL CHECK (action IN ('created', 'acknowledged', 'resolved', 'escalated')),
    user_id VARCHAR(255),
    details JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_history_alert ON neuronip.alert_history(alert_id);
CREATE INDEX idx_alert_history_timestamp ON neuronip.alert_history(timestamp DESC);

-- Distributed traces
CREATE TABLE IF NOT EXISTS neuronip.distributed_traces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id VARCHAR(255) NOT NULL,
    span_id VARCHAR(255) NOT NULL,
    parent_span_id VARCHAR(255),
    operation VARCHAR(255) NOT NULL,
    service VARCHAR(100) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    tags JSONB,
    logs JSONB,
    status VARCHAR(20) CHECK (status IN ('ok', 'error', 'timeout')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_distributed_traces_trace_id ON neuronip.distributed_traces(trace_id);
CREATE INDEX idx_distributed_traces_span_id ON neuronip.distributed_traces(span_id);
CREATE INDEX idx_distributed_traces_service ON neuronip.distributed_traces(service);
CREATE INDEX idx_distributed_traces_start_time ON neuronip.distributed_traces(start_time DESC);

-- Business metrics snapshot
CREATE TABLE IF NOT EXISTS neuronip.business_metrics_snapshot (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    queries_per_second DOUBLE PRECISION,
    avg_latency_seconds DOUBLE PRECISION,
    p99_latency_seconds DOUBLE PRECISION,
    p95_latency_seconds DOUBLE PRECISION,
    error_rate_percent DOUBLE PRECISION,
    cpu_usage_percent DOUBLE PRECISION,
    memory_usage_percent DOUBLE PRECISION,
    disk_usage_percent DOUBLE PRECISION,
    network_io_mbps DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_business_metrics_period ON neuronip.business_metrics_snapshot(period_start, period_end);
CREATE INDEX idx_business_metrics_created ON neuronip.business_metrics_snapshot(created_at DESC);

COMMENT ON TABLE neuronip.aggregated_logs IS 'Centralized log aggregation';
COMMENT ON TABLE neuronip.enterprise_alert_rules IS 'Enterprise alert rule definitions';
COMMENT ON TABLE neuronip.enterprise_alerts IS 'Enterprise alerts';
COMMENT ON TABLE neuronip.alert_history IS 'Alert history and deduplication';
COMMENT ON TABLE neuronip.distributed_traces IS 'Distributed tracing data';
COMMENT ON TABLE neuronip.business_metrics_snapshot IS 'Business metrics snapshots';
