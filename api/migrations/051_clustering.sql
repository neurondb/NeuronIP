-- Migration: 051_clustering.sql
-- Description: Enhanced clustering with distributed task queue, Raft consensus, and multi-region support

-- Cluster tasks table for distributed task queue
CREATE TABLE IF NOT EXISTS neuronip.cluster_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_type VARCHAR(100) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    payload JSONB NOT NULL,
    metadata JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    assigned_node_id VARCHAR(255),
    result JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_cluster_tasks_status ON neuronip.cluster_tasks(status);
CREATE INDEX idx_cluster_tasks_type ON neuronip.cluster_tasks(task_type);
CREATE INDEX idx_cluster_tasks_priority ON neuronip.cluster_tasks(priority DESC, created_at ASC);
CREATE INDEX idx_cluster_tasks_assigned ON neuronip.cluster_tasks(assigned_node_id) WHERE status = 'processing';
CREATE INDEX idx_cluster_tasks_created ON neuronip.cluster_tasks(created_at);

-- Raft consensus state
CREATE TABLE IF NOT EXISTS neuronip.raft_state (
    node_id VARCHAR(255) PRIMARY KEY,
    term BIGINT NOT NULL DEFAULT 0,
    state VARCHAR(50) NOT NULL DEFAULT 'follower', -- 'follower', 'candidate', 'leader'
    leader_id VARCHAR(255),
    voted_for VARCHAR(255),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_raft_state_term ON neuronip.raft_state(term);
CREATE INDEX idx_raft_state_leader ON neuronip.raft_state(leader_id) WHERE state = 'leader';

-- Regions table for multi-region support
CREATE TABLE IF NOT EXISTS neuronip.regions (
    region_code VARCHAR(50) PRIMARY KEY,
    region_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'inactive'
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_regions_status ON neuronip.regions(status);

-- Region replication tracking
CREATE TABLE IF NOT EXISTS neuronip.region_replication (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_region VARCHAR(50) NOT NULL REFERENCES neuronip.regions(region_code),
    target_region VARCHAR(50) NOT NULL REFERENCES neuronip.regions(region_code),
    data_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'syncing', 'completed', 'failed'
    last_sync_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_region_replication_source ON neuronip.region_replication(source_region);
CREATE INDEX idx_region_replication_target ON neuronip.region_replication(target_region);
CREATE INDEX idx_region_replication_status ON neuronip.region_replication(status);

-- Cluster health monitoring
CREATE TABLE IF NOT EXISTS neuronip.cluster_health (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id VARCHAR(255),
    health_status VARCHAR(50) NOT NULL, -- 'healthy', 'degraded', 'unhealthy'
    metrics JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cluster_health_node ON neuronip.cluster_health(node_id);
CREATE INDEX idx_cluster_health_timestamp ON neuronip.cluster_health(timestamp DESC);

-- Function to request vote (Raft consensus)
CREATE OR REPLACE FUNCTION neuronip.request_vote(
    p_node_id VARCHAR,
    p_candidate_id VARCHAR,
    p_term BIGINT
)
RETURNS BOOLEAN AS $$
DECLARE
    v_current_term BIGINT;
    v_voted_for VARCHAR;
BEGIN
    -- Get current state
    SELECT term, voted_for INTO v_current_term, v_voted_for
    FROM neuronip.raft_state
    WHERE node_id = p_node_id;

    -- If no state exists, create it
    IF v_current_term IS NULL THEN
        INSERT INTO neuronip.raft_state (node_id, term, state, voted_for, updated_at)
        VALUES (p_node_id, p_term, 'follower', p_candidate_id, NOW());
        RETURN TRUE;
    END IF;

    -- If term is greater, update and vote
    IF p_term > v_current_term THEN
        UPDATE neuronip.raft_state
        SET term = p_term, state = 'follower', voted_for = p_candidate_id, updated_at = NOW()
        WHERE node_id = p_node_id;
        RETURN TRUE;
    END IF;

    -- If same term and not voted, vote
    IF p_term = v_current_term AND (v_voted_for IS NULL OR v_voted_for = p_candidate_id) THEN
        UPDATE neuronip.raft_state
        SET voted_for = p_candidate_id, updated_at = NOW()
        WHERE node_id = p_node_id;
        RETURN TRUE;
    END IF;

    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- Function to append entries (Raft consensus)
CREATE OR REPLACE FUNCTION neuronip.append_entries(
    p_node_id VARCHAR,
    p_leader_id VARCHAR,
    p_term BIGINT
)
RETURNS BOOLEAN AS $$
DECLARE
    v_current_term BIGINT;
BEGIN
    -- Get current term
    SELECT term INTO v_current_term
    FROM neuronip.raft_state
    WHERE node_id = p_node_id;

    -- If no state exists, create it
    IF v_current_term IS NULL THEN
        INSERT INTO neuronip.raft_state (node_id, term, state, leader_id, updated_at)
        VALUES (p_node_id, p_term, 'follower', p_leader_id, NOW());
        RETURN TRUE;
    END IF;

    -- If term is greater or equal, accept
    IF p_term >= v_current_term THEN
        UPDATE neuronip.raft_state
        SET term = p_term, state = 'follower', leader_id = p_leader_id, voted_for = NULL, updated_at = NOW()
        WHERE node_id = p_node_id;
        RETURN TRUE;
    END IF;

    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON TABLE neuronip.cluster_tasks IS 'Distributed task queue for cluster processing';
COMMENT ON TABLE neuronip.raft_state IS 'Raft consensus protocol state';
COMMENT ON TABLE neuronip.regions IS 'Multi-region configuration';
COMMENT ON TABLE neuronip.region_replication IS 'Cross-region replication tracking';
COMMENT ON TABLE neuronip.cluster_health IS 'Cluster health monitoring';
