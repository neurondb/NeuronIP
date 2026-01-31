package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides cluster coordination and management */
type Service struct {
	pool          *pgxpool.Pool
	coordinator   *Coordinator
	membership    *MembershipManager
	taskQueue     *TaskQueue
	autoscaler    *AutoScaler
	regionManager *RegionManager
	raftNode      *RaftNode
}

/* NewService creates a new cluster service */
func NewService(pool *pgxpool.Pool) (*Service, error) {
	coordinator, err := NewCoordinator(pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create coordinator: %w", err)
	}

	service := &Service{
		pool:          pool,
		coordinator:   coordinator,
		membership:    NewMembershipManager(pool),
		taskQueue:     NewTaskQueue(pool),
		autoscaler:    NewAutoScaler(pool),
		regionManager: NewRegionManager(pool),
	}

	// Initialize Raft node
	service.raftNode = NewRaftNode(pool, coordinator.GetNodeID())

	return service, nil
}

/* Start starts the cluster service */
func (s *Service) Start(ctx context.Context, nodeType string, port int, capabilities map[string]interface{}) error {
	// Register node
	if err := s.coordinator.RegisterNode(ctx, nodeType, port, capabilities); err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}

	// Start heartbeat
	go s.coordinator.StartHeartbeat(ctx, 5*time.Second)

	// Start Raft consensus
	if err := s.raftNode.Start(ctx); err != nil {
		return fmt.Errorf("failed to start raft: %w", err)
	}

	// Start auto-scaling evaluation loop
	go s.autoscalingLoop(ctx)

	return nil
}

/* Stop stops the cluster service */
func (s *Service) Stop(ctx context.Context) error {
	// Deregister node
	if err := s.coordinator.DeregisterNode(ctx); err != nil {
		return err
	}

	// Stop Raft
	s.raftNode.Stop()

	return nil
}

/* GetNodeID returns the current node ID */
func (s *Service) GetNodeID() string {
	return s.coordinator.GetNodeID()
}

/* IsLeader returns true if this node is the leader */
func (s *Service) IsLeader() bool {
	return s.raftNode.IsLeader()
}

/* GetLeader returns the current leader node ID */
func (s *Service) GetLeader() *string {
	return s.raftNode.GetLeader()
}

/* GetClusterHealth returns cluster health */
func (s *Service) GetClusterHealth(ctx context.Context) (*ClusterHealth, error) {
	return s.membership.GetClusterHealth(ctx)
}

/* EnqueueTask enqueues a task */
func (s *Service) EnqueueTask(ctx context.Context, task Task) (*Task, error) {
	return s.taskQueue.EnqueueTask(ctx, task)
}

/* DequeueTask dequeues a task */
func (s *Service) DequeueTask(ctx context.Context, taskTypes []string) (*Task, error) {
	return s.taskQueue.DequeueTask(ctx, s.GetNodeID(), taskTypes)
}

/* GetActiveNodes returns active nodes */
func (s *Service) GetActiveNodes(ctx context.Context, nodeType string) ([]Node, error) {
	return s.coordinator.GetActiveNodes(ctx, nodeType)
}

/* CreateAutoScalingPolicy creates an auto-scaling policy */
func (s *Service) CreateAutoScalingPolicy(ctx context.Context, policy AutoScalingPolicy) (*AutoScalingPolicy, error) {
	return s.autoscaler.CreatePolicy(ctx, policy)
}

/* EvaluateAutoScaling evaluates auto-scaling policies */
func (s *Service) EvaluateAutoScaling(ctx context.Context, policyID string) (*ScalingDecision, error) {
	parsed, err := uuid.Parse(policyID)
	if err != nil {
		return nil, fmt.Errorf("invalid policy ID: %w", err)
	}
	return s.autoscaler.EvaluatePolicy(ctx, parsed)
}

/* RouteToRegion routes a request to the appropriate region */
func (s *Service) RouteToRegion(ctx context.Context, routeKey string, strategy string) (*string, error) {
	return s.regionManager.RouteToRegion(ctx, routeKey, strategy)
}

/* autoscalingLoop runs the auto-scaling evaluation loop */
func (s *Service) autoscalingLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Only leader evaluates auto-scaling
			if !s.IsLeader() {
				continue
			}

			// Get all enabled policies
			// Evaluate and execute scaling decisions
			// This would query policies and evaluate them
		}
	}
}
