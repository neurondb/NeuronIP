package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* MembershipManager manages node membership and health monitoring */
type MembershipManager struct {
	pool *pgxpool.Pool
}

/* NewMembershipManager creates a new membership manager */
func NewMembershipManager(pool *pgxpool.Pool) *MembershipManager {
	return &MembershipManager{pool: pool}
}

/* MarkNodeDraining marks a node as draining (not accepting new requests) */
func (m *MembershipManager) MarkNodeDraining(ctx context.Context, nodeID string) error {
	query := `
		UPDATE neuronip.cluster_nodes 
		SET status = 'draining', updated_at = NOW()
		WHERE node_id = $1`
	_, err := m.pool.Exec(ctx, query, nodeID)
	return err
}

/* MarkNodeInactive marks a node as inactive */
func (m *MembershipManager) MarkNodeInactive(ctx context.Context, nodeID string) error {
	query := `
		UPDATE neuronip.cluster_nodes 
		SET status = 'inactive', updated_at = NOW()
		WHERE node_id = $1`
	_, err := m.pool.Exec(ctx, query, nodeID)
	return err
}

/* CleanupStaleNodes removes nodes that haven't sent heartbeat in specified duration */
func (m *MembershipManager) CleanupStaleNodes(ctx context.Context, staleThreshold time.Duration) error {
	query := `
		UPDATE neuronip.cluster_nodes 
		SET status = 'inactive', updated_at = NOW()
		WHERE status = 'active' 
			AND last_heartbeat < NOW() - INTERVAL '%d seconds'`
	
	thresholdSeconds := int(staleThreshold.Seconds())
	query = fmt.Sprintf(query, thresholdSeconds)
	
	_, err := m.pool.Exec(ctx, query)
	return err
}

/* GetNodeHealth retrieves health status of a node */
func (m *MembershipManager) GetNodeHealth(ctx context.Context, nodeID string) (*NodeHealth, error) {
	query := `
		SELECT 
			node_id, 
			status, 
			last_heartbeat,
			NOW() - last_heartbeat as time_since_heartbeat
		FROM neuronip.cluster_nodes
		WHERE node_id = $1`

	var health NodeHealth
	var timeSinceHeartbeat time.Duration

	err := m.pool.QueryRow(ctx, query, nodeID).Scan(
		&health.NodeID, &health.Status, &health.LastHeartbeat, &timeSinceHeartbeat,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get node health: %w", err)
	}

	health.TimeSinceHeartbeat = timeSinceHeartbeat
	health.IsHealthy = health.Status == "active" && timeSinceHeartbeat < 30*time.Second

	return &health, nil
}

/* GetClusterHealth retrieves overall cluster health */
func (m *MembershipManager) GetClusterHealth(ctx context.Context) (*ClusterHealth, error) {
	query := `
		SELECT 
			COUNT(*) as total_nodes,
			SUM(CASE WHEN status = 'active' AND last_heartbeat > NOW() - INTERVAL '30 seconds' THEN 1 ELSE 0 END) as healthy_nodes,
			SUM(CASE WHEN status = 'draining' THEN 1 ELSE 0 END) as draining_nodes,
			SUM(CASE WHEN status = 'inactive' THEN 1 ELSE 0 END) as inactive_nodes
		FROM neuronip.cluster_nodes`

	var health ClusterHealth
	err := m.pool.QueryRow(ctx, query).Scan(
		&health.TotalNodes, &health.HealthyNodes, &health.DrainingNodes, &health.InactiveNodes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster health: %w", err)
	}

	health.IsHealthy = health.HealthyNodes > 0
	if health.TotalNodes > 0 {
		health.HealthPercentage = float64(health.HealthyNodes) / float64(health.TotalNodes) * 100
	}

	return &health, nil
}

/* UpdateNodeCapabilities updates node capabilities */
func (m *MembershipManager) UpdateNodeCapabilities(ctx context.Context, nodeID string, capabilities map[string]interface{}) error {
	capabilitiesJSON, _ := json.Marshal(capabilities)
	query := `
		UPDATE neuronip.cluster_nodes 
		SET capabilities = $1, updated_at = NOW()
		WHERE node_id = $2`
	_, err := m.pool.Exec(ctx, query, capabilitiesJSON, nodeID)
	return err
}

/* NodeHealth represents node health status */
type NodeHealth struct {
	NodeID             string        `json:"node_id"`
	Status             string        `json:"status"`
	LastHeartbeat      time.Time     `json:"last_heartbeat"`
	TimeSinceHeartbeat time.Duration `json:"time_since_heartbeat"`
	IsHealthy          bool          `json:"is_healthy"`
}

/* ClusterHealth represents overall cluster health */
type ClusterHealth struct {
	TotalNodes       int     `json:"total_nodes"`
	HealthyNodes     int     `json:"healthy_nodes"`
	DrainingNodes     int     `json:"draining_nodes"`
	InactiveNodes    int     `json:"inactive_nodes"`
	IsHealthy        bool    `json:"is_healthy"`
	HealthPercentage float64 `json:"health_percentage"`
}
