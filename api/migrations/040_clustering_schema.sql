-- Migration: 040_clustering_schema.sql
-- Description: Schema for cluster coordination and horizontal scaling

-- Cluster nodes table
CREATE TABLE IF NOT EXISTS neuronip.cluster_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id VARCHAR(255) UNIQUE NOT NULL,
    node_type VARCHAR(50) NOT NULL, -- 'api', 'worker', 'scheduler', etc.
    hostname VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    port INTEGER NOT NULL,
    region VARCHAR(100),
    zone VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'draining', 'inactive'
    capabilities JSONB, -- Node capabilities (CPU, memory, features)
    metadata JSONB,
    last_heartbeat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    registered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cluster_nodes_status ON neuronip.cluster_nodes(status);
CREATE INDEX idx_cluster_nodes_type ON neuronip.cluster_nodes(node_type);
CREATE INDEX idx_cluster_nodes_heartbeat ON neuronip.cluster_nodes(last_heartbeat);

-- Cluster shards table
CREATE TABLE IF NOT EXISTS neuronip.cluster_shards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shard_key VARCHAR(255) NOT NULL,
    shard_type VARCHAR(50) NOT NULL, -- 'hash', 'range', 'directory'
    node_id VARCHAR(255) NOT NULL REFERENCES neuronip.cluster_nodes(node_id),
    range_start VARCHAR(255),
    range_end VARCHAR(255),
    directory_path VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cluster_shards_key ON neuronip.cluster_shards(shard_key);
CREATE INDEX idx_cluster_shards_node ON neuronip.cluster_shards(node_id);
CREATE INDEX idx_cluster_shards_type ON neuronip.cluster_shards(shard_type);

-- Distributed locks table
CREATE TABLE IF NOT EXISTS neuronip.distributed_locks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lock_key VARCHAR(255) UNIQUE NOT NULL,
    lock_holder VARCHAR(255) NOT NULL, -- node_id
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_distributed_locks_key ON neuronip.distributed_locks(lock_key);
CREATE INDEX idx_distributed_locks_expires ON neuronip.distributed_locks(expires_at);

-- Cluster metrics table
CREATE TABLE IF NOT EXISTS neuronip.cluster_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id VARCHAR(255) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    metric_unit VARCHAR(50),
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cluster_metrics_node ON neuronip.cluster_metrics(node_id);
CREATE INDEX idx_cluster_metrics_name ON neuronip.cluster_metrics(metric_name);
CREATE INDEX idx_cluster_metrics_timestamp ON neuronip.cluster_metrics(timestamp);

-- Request routing table
CREATE TABLE IF NOT EXISTS neuronip.request_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_key VARCHAR(255) NOT NULL, -- e.g., tenant_id, user_id, etc.
    target_node_id VARCHAR(255) NOT NULL REFERENCES neuronip.cluster_nodes(node_id),
    routing_strategy VARCHAR(50) NOT NULL, -- 'hash', 'round_robin', 'least_connections', 'sticky'
    priority INTEGER DEFAULT 0,
    weight INTEGER DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_routes_key ON neuronip.request_routes(route_key);
CREATE INDEX idx_request_routes_node ON neuronip.request_routes(target_node_id);
CREATE INDEX idx_request_routes_strategy ON neuronip.request_routes(routing_strategy);

-- Auto-scaling policies
CREATE TABLE IF NOT EXISTS neuronip.auto_scaling_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_name VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL, -- 'api', 'worker', 'database', etc.
    min_instances INTEGER NOT NULL DEFAULT 1,
    max_instances INTEGER NOT NULL DEFAULT 10,
    target_metric VARCHAR(100) NOT NULL, -- 'cpu', 'memory', 'requests_per_second', etc.
    target_value DOUBLE PRECISION NOT NULL,
    scale_up_threshold DOUBLE PRECISION NOT NULL,
    scale_down_threshold DOUBLE PRECISION NOT NULL,
    cooldown_seconds INTEGER NOT NULL DEFAULT 300,
    enabled BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auto_scaling_policies_type ON neuronip.auto_scaling_policies(resource_type);
CREATE INDEX idx_auto_scaling_policies_enabled ON neuronip.auto_scaling_policies(enabled);

-- Auto-scaling events
CREATE TABLE IF NOT EXISTS neuronip.auto_scaling_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES neuronip.auto_scaling_policies(id),
    action VARCHAR(50) NOT NULL, -- 'scale_up', 'scale_down'
    current_instances INTEGER NOT NULL,
    target_instances INTEGER NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    reason TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'executing', 'completed', 'failed'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_auto_scaling_events_policy ON neuronip.auto_scaling_events(policy_id);
CREATE INDEX idx_auto_scaling_events_status ON neuronip.auto_scaling_events(status);
CREATE INDEX idx_auto_scaling_events_created ON neuronip.auto_scaling_events(created_at);

-- Function to clean up expired locks
CREATE OR REPLACE FUNCTION neuronip.cleanup_expired_locks()
RETURNS void AS $$
BEGIN
    DELETE FROM neuronip.distributed_locks
    WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- Function to update node heartbeat
CREATE OR REPLACE FUNCTION neuronip.update_node_heartbeat(p_node_id VARCHAR)
RETURNS void AS $$
BEGIN
    UPDATE neuronip.cluster_nodes
    SET last_heartbeat = NOW(), updated_at = NOW()
    WHERE node_id = p_node_id;
END;
$$ LANGUAGE plpgsql;

-- Function to get active nodes
CREATE OR REPLACE FUNCTION neuronip.get_active_nodes(p_node_type VARCHAR DEFAULT NULL)
RETURNS TABLE (
    id UUID,
    node_id VARCHAR,
    node_type VARCHAR,
    hostname VARCHAR,
    ip_address INET,
    port INTEGER,
    region VARCHAR,
    zone VARCHAR,
    status VARCHAR,
    capabilities JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cn.id,
        cn.node_id,
        cn.node_type,
        cn.hostname,
        cn.ip_address,
        cn.port,
        cn.region,
        cn.zone,
        cn.status,
        cn.capabilities
    FROM neuronip.cluster_nodes cn
    WHERE cn.status = 'active'
        AND cn.last_heartbeat > NOW() - INTERVAL '30 seconds'
        AND (p_node_type IS NULL OR cn.node_type = p_node_type)
    ORDER BY cn.last_heartbeat DESC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON TABLE neuronip.cluster_nodes IS 'Cluster node registry for horizontal scaling';
COMMENT ON TABLE neuronip.cluster_shards IS 'Data sharding configuration';
COMMENT ON TABLE neuronip.distributed_locks IS 'Distributed locking mechanism';
COMMENT ON TABLE neuronip.cluster_metrics IS 'Cluster-wide metrics collection';
COMMENT ON TABLE neuronip.request_routes IS 'Request routing configuration';
COMMENT ON TABLE neuronip.auto_scaling_policies IS 'Auto-scaling policy definitions';
COMMENT ON TABLE neuronip.auto_scaling_events IS 'Auto-scaling event history';
