package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Coordinator manages cluster coordination and node membership */
type Coordinator struct {
	pool   *pgxpool.Pool
	nodeID string
}

/* NewCoordinator creates a new cluster coordinator */
func NewCoordinator(pool *pgxpool.Pool) (*Coordinator, error) {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		hostname, _ := os.Hostname()
		nodeID = fmt.Sprintf("%s-%s", hostname, uuid.New().String()[:8])
	}

	return &Coordinator{
		pool:   pool,
		nodeID: nodeID,
	}, nil
}

/* RegisterNode registers this node in the cluster */
func (c *Coordinator) RegisterNode(ctx context.Context, nodeType string, port int, capabilities map[string]interface{}) error {
	hostname, _ := os.Hostname()
	ipAddr := c.getLocalIP()

	capabilitiesJSON, _ := json.Marshal(capabilities)
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"hostname": hostname,
		"pid":      os.Getpid(),
	})

	query := `
		INSERT INTO neuronip.cluster_nodes 
		(node_id, node_type, hostname, ip_address, port, status, capabilities, metadata, last_heartbeat, registered_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW(), NOW())
		ON CONFLICT (node_id) DO UPDATE SET
			status = $6,
			capabilities = $7,
			metadata = $8,
			last_heartbeat = NOW(),
			updated_at = NOW()`

	_, err := c.pool.Exec(ctx, query, c.nodeID, nodeType, hostname, ipAddr, port, "active", capabilitiesJSON, metadataJSON)
	return err
}

/* UpdateHeartbeat updates the node heartbeat */
func (c *Coordinator) UpdateHeartbeat(ctx context.Context) error {
	query := `SELECT neuronip.update_node_heartbeat($1)`
	_, err := c.pool.Exec(ctx, query, c.nodeID)
	return err
}

/* GetActiveNodes retrieves active nodes in the cluster */
func (c *Coordinator) GetActiveNodes(ctx context.Context, nodeType string) ([]Node, error) {
	var query string
	var args []interface{}

	if nodeType != "" {
		query = `SELECT * FROM neuronip.get_active_nodes($1)`
		args = []interface{}{nodeType}
	} else {
		query = `SELECT * FROM neuronip.get_active_nodes(NULL)`
		args = []interface{}{}
	}

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get active nodes: %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var node Node
		var capabilitiesJSON json.RawMessage

		err := rows.Scan(
			&node.ID, &node.NodeID, &node.NodeType, &node.Hostname,
			&node.IPAddress, &node.Port, &node.Region, &node.Zone,
			&node.Status, &capabilitiesJSON,
		)
		if err != nil {
			continue
		}

		if capabilitiesJSON != nil {
			json.Unmarshal(capabilitiesJSON, &node.Capabilities)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

/* StartHeartbeat starts the heartbeat loop */
func (c *Coordinator) StartHeartbeat(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send initial heartbeat
	c.UpdateHeartbeat(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.UpdateHeartbeat(ctx)
		}
	}
}

/* DeregisterNode removes this node from the cluster */
func (c *Coordinator) DeregisterNode(ctx context.Context) error {
	query := `
		UPDATE neuronip.cluster_nodes 
		SET status = 'inactive', updated_at = NOW()
		WHERE node_id = $1`
	_, err := c.pool.Exec(ctx, query, c.nodeID)
	return err
}

/* GetNodeID returns the current node ID */
func (c *Coordinator) GetNodeID() string {
	return c.nodeID
}

/* getLocalIP gets the local IP address */
func (c *Coordinator) getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

/* Node represents a cluster node */
type Node struct {
	ID           uuid.UUID              `json:"id"`
	NodeID       string                 `json:"node_id"`
	NodeType     string                 `json:"node_type"`
	Hostname     string                 `json:"hostname"`
	IPAddress    string                 `json:"ip_address"`
	Port         int                    `json:"port"`
	Region       *string                `json:"region,omitempty"`
	Zone         *string                `json:"zone,omitempty"`
	Status       string                 `json:"status"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
}
