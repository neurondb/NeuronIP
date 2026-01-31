package knowledgegraph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* AlgorithmsService provides graph algorithms */
type AlgorithmsService struct {
	pool *pgxpool.Pool
}

/* NewAlgorithmsService creates a new algorithms service */
func NewAlgorithmsService(pool *pgxpool.Pool) *AlgorithmsService {
	return &AlgorithmsService{pool: pool}
}

/* CalculatePageRank calculates PageRank for entities using entity_links. Uses iterative updates with damping factor 0.85. */
func (as *AlgorithmsService) CalculatePageRank(ctx context.Context, maxIterations int) (map[uuid.UUID]float64, error) {
	results := make(map[uuid.UUID]float64)

	rows, err := as.pool.Query(ctx, `
		SELECT source_entity_id, target_entity_id FROM neuronip.entity_links
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to read entity_links: %w", err)
	}
	defer rows.Close()

	type edge struct{ from, to uuid.UUID }
	var edges []edge
	nodes := make(map[uuid.UUID]struct{})
	for rows.Next() {
		var from, to uuid.UUID
		if err := rows.Scan(&from, &to); err != nil {
			return nil, err
		}
		edges = append(edges, edge{from, to})
		nodes[from] = struct{}{}
		nodes[to] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	N := float64(len(nodes))
	if N == 0 {
		return results, nil
	}

	outDegree := make(map[uuid.UUID]int)
	incoming := make(map[uuid.UUID][]uuid.UUID)
	for _, e := range edges {
		outDegree[e.from]++
		incoming[e.to] = append(incoming[e.to], e.from)
	}

	for id := range nodes {
		results[id] = 1.0 / N
	}

	const damping = 0.85
	for iter := 0; iter < maxIterations; iter++ {
		next := make(map[uuid.UUID]float64)
		for id := range nodes {
			next[id] = (1.0 - damping) / N
		}
		for id, sources := range incoming {
			for _, src := range sources {
				deg := outDegree[src]
				if deg > 0 {
					next[id] += damping * (results[src] / float64(deg))
				}
			}
		}
		results = next
	}

	return results, nil
}

/* DetectCommunities detects connected components in the graph (each component gets a distinct integer id). Uses BFS over entity_links. */
func (as *AlgorithmsService) DetectCommunities(ctx context.Context) (map[uuid.UUID]int, error) {
	communities := make(map[uuid.UUID]int)

	rows, err := as.pool.Query(ctx, `
		SELECT source_entity_id, target_entity_id FROM neuronip.entity_links
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to read entity_links: %w", err)
	}
	defer rows.Close()

	adj := make(map[uuid.UUID][]uuid.UUID)
	allNodes := make(map[uuid.UUID]struct{})
	for rows.Next() {
		var from, to uuid.UUID
		if err := rows.Scan(&from, &to); err != nil {
			return nil, err
		}
		adj[from] = append(adj[from], to)
		adj[to] = append(adj[to], from)
		allNodes[from] = struct{}{}
		allNodes[to] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	visited := make(map[uuid.UUID]bool)
	componentID := 0
	for node := range allNodes {
		if visited[node] {
			continue
		}
		// BFS to mark all nodes in this component
		queue := []uuid.UUID{node}
		visited[node] = true
		communities[node] = componentID
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for _, neighbor := range adj[cur] {
				if !visited[neighbor] {
					visited[neighbor] = true
					communities[neighbor] = componentID
					queue = append(queue, neighbor)
				}
			}
		}
		componentID++
	}

	return communities, nil
}

/* FindShortestPath finds shortest path between entities */
func (as *AlgorithmsService) FindShortestPath(ctx context.Context, startEntityID, endEntityID uuid.UUID) ([]uuid.UUID, error) {
	// Shortest path algorithm (Dijkstra or BFS)
	// Use recursive CTE in PostgreSQL

	query := `
		WITH RECURSIVE path_search AS (
			SELECT source_entity_id, target_entity_id, ARRAY[source_entity_id] as path, 1 as depth
			FROM neuronip.entity_links
			WHERE source_entity_id = $1
			
			UNION ALL
			
			SELECT el.source_entity_id, el.target_entity_id, ps.path || el.target_entity_id, ps.depth + 1
			FROM neuronip.entity_links el
			JOIN path_search ps ON el.source_entity_id = ps.target_entity_id
			WHERE el.target_entity_id != ALL(ps.path)
				AND ps.depth < 10
		)
		SELECT path
		FROM path_search
		WHERE target_entity_id = $2
		ORDER BY depth
		LIMIT 1`

	var path []uuid.UUID
	err := as.pool.QueryRow(ctx, query, startEntityID, endEntityID).Scan(&path)
	if err != nil {
		return nil, fmt.Errorf("path not found: %w", err)
	}

	return path, nil
}
