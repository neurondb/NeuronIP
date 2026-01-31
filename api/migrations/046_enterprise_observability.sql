-- Migration: 046_enterprise_observability.sql
-- Description: Enterprise observability enhancements

-- Distributed tracing spans
CREATE TABLE IF NOT EXISTS neuronip.trace_spans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id VARCHAR(255) NOT NULL,
    span_id VARCHAR(255) NOT NULL,
    parent_span_id VARCHAR(255),
    operation_name VARCHAR(255) NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_ms INTEGER NOT NULL,
    tags JSONB,
    logs JSONB,
    status VARCHAR(50), -- 'ok', 'error'
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trace_spans_trace ON neuronip.trace_spans(trace_id);
CREATE INDEX idx_trace_spans_span ON neuronip.trace_spans(span_id);
CREATE INDEX idx_trace_spans_service ON neuronip.trace_spans(service_name);
CREATE INDEX idx_trace_spans_start_time ON neuronip.trace_spans(start_time);

-- Alert rules
CREATE TABLE IF NOT EXISTS neuronip.alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_name VARCHAR(255) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    condition_type VARCHAR(50) NOT NULL, -- 'threshold', 'rate_change', 'anomaly'
    condition_config JSONB NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning', -- 'critical', 'warning', 'info'
    enabled BOOLEAN NOT NULL DEFAULT true,
    notification_channels JSONB, -- Where to send alerts
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_rules_enabled ON neuronip.alert_rules(enabled) WHERE enabled = true;
CREATE INDEX idx_alert_rules_metric ON neuronip.alert_rules(metric_name);

-- Alert events
CREATE TABLE IF NOT EXISTS neuronip.alert_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES neuronip.alert_rules(id),
    status VARCHAR(50) NOT NULL DEFAULT 'firing', -- 'firing', 'resolved', 'acknowledged'
    metric_value DOUBLE PRECISION NOT NULL,
    threshold_value DOUBLE PRECISION,
    message TEXT,
    acknowledged_by VARCHAR(255),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_events_rule ON neuronip.alert_events(rule_id);
CREATE INDEX idx_alert_events_status ON neuronip.alert_events(status);
CREATE INDEX idx_alert_events_created ON neuronip.alert_events(created_at);

-- Observability dashboards
CREATE TABLE IF NOT EXISTS neuronip.observability_dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_name VARCHAR(255) NOT NULL,
    description TEXT,
    dashboard_config JSONB NOT NULL, -- Dashboard layout and widgets
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_observability_dashboards_created_by ON neuronip.observability_dashboards(created_by);

COMMENT ON TABLE neuronip.trace_spans IS 'Distributed tracing spans';
COMMENT ON TABLE neuronip.alert_rules IS 'Alert rule definitions';
COMMENT ON TABLE neuronip.alert_events IS 'Alert event history';
COMMENT ON TABLE neuronip.observability_dashboards IS 'Observability dashboard configurations';
