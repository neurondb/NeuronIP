package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* RaftNode implements Raft consensus protocol for leader election */
type RaftNode struct {
	pool       *pgxpool.Pool
	nodeID     string
	term       int64
	state      string // "follower", "candidate", "leader"
	leaderID   *string
	votedFor   *string
	mu         sync.RWMutex
	stopCh     chan struct{}
}

/* NewRaftNode creates a new Raft node */
func NewRaftNode(pool *pgxpool.Pool, nodeID string) *RaftNode {
	return &RaftNode{
		pool:   pool,
		nodeID: nodeID,
		term:   0,
		state:  "follower",
		stopCh: make(chan struct{}),
	}
}

/* Start starts the Raft node */
func (r *RaftNode) Start(ctx context.Context) error {
	// Initialize Raft state in database
	if err := r.initializeRaftState(ctx); err != nil {
		return fmt.Errorf("failed to initialize raft state: %w", err)
	}

	// Start election timer
	go r.electionTimer(ctx)

	// Start heartbeat if leader
	go r.heartbeatLoop(ctx)

	return nil
}

/* Stop stops the Raft node */
func (r *RaftNode) Stop() {
	close(r.stopCh)
}

/* IsLeader returns true if this node is the leader */
func (r *RaftNode) IsLeader() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state == "leader"
}

/* GetLeader returns the current leader node ID */
func (r *RaftNode) GetLeader() *string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.leaderID
}

/* RequestVote handles vote requests from candidates */
func (r *RaftNode) RequestVote(ctx context.Context, candidateID string, term int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If term is greater, update term and become follower
	if term > r.term {
		r.term = term
		r.state = "follower"
		r.votedFor = nil
		r.leaderID = nil
	}

	// If already voted for someone else in this term, reject
	if r.votedFor != nil && *r.votedFor != candidateID && r.term == term {
		return false, nil
	}

	// Vote for candidate
	r.votedFor = &candidateID
	r.term = term

	// Update database
	if err := r.updateRaftState(ctx); err != nil {
		return false, err
	}

	return true, nil
}

/* AppendEntries handles append entries from leader */
func (r *RaftNode) AppendEntries(ctx context.Context, leaderID string, term int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If term is greater or equal, accept and become follower
	if term >= r.term {
		r.term = term
		r.state = "follower"
		r.leaderID = &leaderID
		r.votedFor = nil

		// Update database
		if err := r.updateRaftState(ctx); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

/* StartElection starts a leader election */
func (r *RaftNode) StartElection(ctx context.Context) error {
	r.mu.Lock()
	r.term++
	r.state = "candidate"
	r.votedFor = &r.nodeID
	currentTerm := r.term
	r.mu.Unlock()

	// Update database
	if err := r.updateRaftState(ctx); err != nil {
		return err
	}

	// Get all active nodes
	nodes, err := r.getActiveNodes(ctx)
	if err != nil {
		return err
	}

	// Count votes (including self)
	votes := 1
	neededVotes := (len(nodes) / 2) + 1

	// Request votes from other nodes
	for _, node := range nodes {
		if node.NodeID == r.nodeID {
			continue
		}

		voted, err := r.requestVoteFromNode(ctx, node.NodeID, currentTerm)
		if err != nil {
			continue // Skip failed nodes
		}

		if voted {
			votes++
		}

		// If we have majority, become leader
		if votes >= neededVotes {
			r.mu.Lock()
			r.state = "leader"
			r.leaderID = &r.nodeID
			r.mu.Unlock()

			if err := r.updateRaftState(ctx); err != nil {
				return err
			}

			return nil
		}
	}

	// Didn't get majority, remain candidate or become follower
	r.mu.Lock()
	if r.state == "candidate" && r.term == currentTerm {
		r.state = "follower"
		r.votedFor = nil
	}
	r.mu.Unlock()

	return nil
}

/* electionTimer runs the election timer */
func (r *RaftNode) electionTimer(ctx context.Context) {
	ticker := time.NewTicker(150 * time.Millisecond + time.Duration(uuid.New().ID()%100)*time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.mu.RLock()
			isLeader := r.state == "leader"
			r.mu.RUnlock()

			if !isLeader {
				// Check if we haven't heard from leader
				lastHeartbeat, err := r.getLastHeartbeat(ctx)
				if err == nil && time.Since(lastHeartbeat) > 300*time.Millisecond {
					r.StartElection(ctx)
				}
			}
		}
	}
}

/* heartbeatLoop sends heartbeats if this node is the leader */
func (r *RaftNode) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopCh:
			return
		case <-ticker.C:
			if r.IsLeader() {
				r.sendHeartbeat(ctx)
			}
		}
	}
}

/* sendHeartbeat sends heartbeat to all followers */
func (r *RaftNode) sendHeartbeat(ctx context.Context) {
	nodes, err := r.getActiveNodes(ctx)
	if err != nil {
		return
	}

	r.mu.RLock()
	currentTerm := r.term
	r.mu.RUnlock()

	for _, node := range nodes {
		if node.NodeID == r.nodeID {
			continue
		}

		// Send append entries (heartbeat)
		r.appendEntriesToNode(ctx, node.NodeID, currentTerm)
	}
}

/* initializeRaftState initializes Raft state in database */
func (r *RaftNode) initializeRaftState(ctx context.Context) error {
	query := `
		INSERT INTO neuronip.raft_state (node_id, term, state, leader_id, voted_for, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (node_id) DO UPDATE SET
			updated_at = NOW()`

	r.mu.RLock()
	leaderID := r.leaderID
	votedFor := r.votedFor
	term := r.term
	state := r.state
	r.mu.RUnlock()

	_, err := r.pool.Exec(ctx, query, r.nodeID, term, state, leaderID, votedFor)
	return err
}

/* updateRaftState updates Raft state in database */
func (r *RaftNode) updateRaftState(ctx context.Context) error {
	query := `
		UPDATE neuronip.raft_state
		SET term = $1, state = $2, leader_id = $3, voted_for = $4, updated_at = NOW()
		WHERE node_id = $5`

	r.mu.RLock()
	leaderID := r.leaderID
	votedFor := r.votedFor
	term := r.term
	state := r.state
	r.mu.RUnlock()

	_, err := r.pool.Exec(ctx, query, term, state, leaderID, votedFor, r.nodeID)
	return err
}

/* getActiveNodes gets all active nodes */
func (r *RaftNode) getActiveNodes(ctx context.Context) ([]Node, error) {
	query := `
		SELECT node_id, node_type, hostname, ip_address, port, region, zone, status, capabilities
		FROM neuronip.cluster_nodes
		WHERE status = 'active' AND last_heartbeat > NOW() - INTERVAL '30 seconds'`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var node Node
		var capabilitiesJSON json.RawMessage

		err := rows.Scan(
			&node.NodeID, &node.NodeType, &node.Hostname, &node.IPAddress, &node.Port,
			&node.Region, &node.Zone, &node.Status, &capabilitiesJSON,
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

/* requestVoteFromNode requests vote from a specific node */
func (r *RaftNode) requestVoteFromNode(ctx context.Context, nodeID string, term int64) (bool, error) {
	// In a real implementation, this would make an RPC call to the node
	// For now, we'll use database-based coordination
	query := `
		SELECT neuronip.request_vote($1, $2, $3)`

	var voted bool
	err := r.pool.QueryRow(ctx, query, nodeID, r.nodeID, term).Scan(&voted)
	return voted, err
}

/* appendEntriesToNode sends append entries to a specific node */
func (r *RaftNode) appendEntriesToNode(ctx context.Context, nodeID string, term int64) error {
	// In a real implementation, this would make an RPC call to the node
	// For now, we'll use database-based coordination
	query := `
		SELECT neuronip.append_entries($1, $2, $3)`

	_, err := r.pool.Exec(ctx, query, nodeID, r.nodeID, term)
	return err
}

/* getLastHeartbeat gets the last heartbeat time */
func (r *RaftNode) getLastHeartbeat(ctx context.Context) (time.Time, error) {
	query := `
		SELECT COALESCE(MAX(updated_at), '1970-01-01'::timestamp)
		FROM neuronip.raft_state
		WHERE leader_id = $1 AND state = 'leader'`

	var lastHeartbeat time.Time
	err := r.pool.QueryRow(ctx, query, r.nodeID).Scan(&lastHeartbeat)
	return lastHeartbeat, err
}
