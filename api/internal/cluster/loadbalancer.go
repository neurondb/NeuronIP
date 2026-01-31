package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* LoadBalancer manages request routing and load balancing */
type LoadBalancer struct {
	pool       *pgxpool.Pool
	roundRobin map[string]int
	mu         sync.RWMutex
}

/* NewLoadBalancer creates a new load balancer */
func NewLoadBalancer(pool *pgxpool.Pool) *LoadBalancer {
	return &LoadBalancer{
		pool:       pool,
		roundRobin: make(map[string]int),
	}
}

/* RouteRequest routes a request to an appropriate node */
func (lb *LoadBalancer) RouteRequest(ctx context.Context, routeKey string, strategy string) (*Route, error) {
	switch strategy {
	case "hash":
		return lb.routeByHash(ctx, routeKey)
	case "round_robin":
		return lb.routeRoundRobin(ctx, routeKey)
	case "least_connections":
		return lb.routeLeastConnections(ctx, routeKey)
	case "sticky":
		return lb.routeSticky(ctx, routeKey)
	default:
		return lb.routeRoundRobin(ctx, routeKey)
	}
}

/* routeByHash routes using hash-based routing */
func (lb *LoadBalancer) routeByHash(ctx context.Context, routeKey string) (*Route, error) {
	query := `
		SELECT target_node_id, routing_strategy, priority, weight
		FROM neuronip.request_routes
		WHERE route_key = $1 AND routing_strategy = 'hash' AND status = 'active'
		ORDER BY priority DESC, weight DESC
		LIMIT 1`

	var route Route
	err := lb.pool.QueryRow(ctx, query, routeKey).Scan(
		&route.TargetNodeID, &route.Strategy, &route.Priority, &route.Weight,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to route by hash: %w", err)
	}

	route.RouteKey = routeKey
	return &route, nil
}

/* routeRoundRobin routes using round-robin */
func (lb *LoadBalancer) routeRoundRobin(ctx context.Context, routeKey string) (*Route, error) {
	query := `
		SELECT target_node_id, routing_strategy, priority, weight
		FROM neuronip.request_routes
		WHERE route_key = $1 AND routing_strategy = 'round_robin' AND status = 'active'
		ORDER BY priority DESC, weight DESC`

	rows, err := lb.pool.Query(ctx, query, routeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to route round robin: %w", err)
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var route Route
		err := rows.Scan(&route.TargetNodeID, &route.Strategy, &route.Priority, &route.Weight)
		if err != nil {
			continue
		}
		route.RouteKey = routeKey
		routes = append(routes, route)
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes found for key: %s", routeKey)
	}

	// Get current round-robin index
	lb.mu.Lock()
	index := lb.roundRobin[routeKey]
	lb.roundRobin[routeKey] = (index + 1) % len(routes)
	lb.mu.Unlock()

	return &routes[index], nil
}

/* routeLeastConnections routes to node with least connections */
func (lb *LoadBalancer) routeLeastConnections(ctx context.Context, routeKey string) (*Route, error) {
	// Get active connections per node
	query := `
		SELECT 
			rr.target_node_id,
			rr.routing_strategy,
			rr.priority,
			rr.weight,
			COALESCE(cm.metric_value, 0) as connection_count
		FROM neuronip.request_routes rr
		LEFT JOIN (
			SELECT node_id, metric_value
			FROM neuronip.cluster_metrics
			WHERE metric_name = 'active_connections'
				AND timestamp > NOW() - INTERVAL '1 minute'
			ORDER BY timestamp DESC
			LIMIT 1
		) cm ON rr.target_node_id = cm.node_id
		WHERE rr.route_key = $1 
			AND rr.routing_strategy = 'least_connections' 
			AND rr.status = 'active'
		ORDER BY connection_count ASC, priority DESC, weight DESC
		LIMIT 1`

	var route Route
	var connectionCount float64
	err := lb.pool.QueryRow(ctx, query, routeKey).Scan(
		&route.TargetNodeID, &route.Strategy, &route.Priority, &route.Weight, &connectionCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to route by least connections: %w", err)
	}

	route.RouteKey = routeKey
	return &route, nil
}

/* routeSticky routes using sticky sessions */
func (lb *LoadBalancer) routeSticky(ctx context.Context, routeKey string) (*Route, error) {
	// For sticky routing, we'd typically use a session ID or user ID
	// This is a simplified version
	return lb.routeByHash(ctx, routeKey)
}

/* CreateRoute creates a new routing rule */
func (lb *LoadBalancer) CreateRoute(ctx context.Context, routeKey string, targetNodeID string, strategy string, priority int, weight int) (*Route, error) {
	routeID := uuid.New()
	query := `
		INSERT INTO neuronip.request_routes 
		(id, route_key, target_node_id, routing_strategy, priority, weight, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'active', NOW(), NOW())
		RETURNING id, route_key, target_node_id, routing_strategy, priority, weight, status, created_at, updated_at`

	var route Route
	err := lb.pool.QueryRow(ctx, query, routeID, routeKey, targetNodeID, strategy, priority, weight).Scan(
		&route.ID, &route.RouteKey, &route.TargetNodeID, &route.Strategy,
		&route.Priority, &route.Weight, &route.Status, &route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}

	return &route, nil
}

/* ListRoutes lists all routes for a given route key */
func (lb *LoadBalancer) ListRoutes(ctx context.Context, routeKey string) ([]Route, error) {
	query := `
		SELECT id, route_key, target_node_id, routing_strategy, priority, weight, status, created_at, updated_at
		FROM neuronip.request_routes
		WHERE route_key = $1
		ORDER BY priority DESC, weight DESC`

	rows, err := lb.pool.Query(ctx, query, routeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var route Route
		err := rows.Scan(
			&route.ID, &route.RouteKey, &route.TargetNodeID, &route.Strategy,
			&route.Priority, &route.Weight, &route.Status, &route.CreatedAt, &route.UpdatedAt,
		)
		if err != nil {
			continue
		}
		routes = append(routes, route)
	}

	return routes, nil
}

/* Route represents a request routing rule */
type Route struct {
	ID           uuid.UUID  `json:"id"`
	RouteKey     string     `json:"route_key"`
	TargetNodeID string     `json:"target_node_id"`
	Strategy     string     `json:"strategy"`
	Priority     int        `json:"priority"`
	Weight       int        `json:"weight"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
