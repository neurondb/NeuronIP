package semantic

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/execution"
)

/* DistributedSearchService provides distributed vector search functionality */
type DistributedSearchService struct {
	pool           *pgxpool.Pool
	executor       *execution.DistributedExecutor
	nodeID         string
	shardCount     int
	shardMutex     sync.RWMutex
}

/* NewDistributedSearchService creates a new distributed search service */
func NewDistributedSearchService(pool *pgxpool.Pool, executor *execution.DistributedExecutor, nodeID string, shardCount int) *DistributedSearchService {
	return &DistributedSearchService{
		pool:       pool,
		executor:   executor,
		nodeID:     nodeID,
		shardCount: shardCount,
	}
}

/* DistributedSearchRequest represents a distributed search request */
type DistributedSearchRequest struct {
	QueryEmbedding string
	TableName      string
	EmbeddingColumn string
	TextColumn     string
	Limit          int
	Threshold      float64
	ShardIDs       []int // Optional: specific shards to search
}

/* DistributedSearchResult represents a distributed search result */
type DistributedSearchResult struct {
	Results    []map[string]interface{}
	ShardResults map[int][]map[string]interface{} // Results per shard
	TotalShards int
	QueriedShards int
}

/* SearchDistributed performs distributed vector search across shards */
func (dss *DistributedSearchService) SearchDistributed(ctx context.Context, req DistributedSearchRequest) (*DistributedSearchResult, error) {
	// Determine which shards to search
	shardsToSearch := req.ShardIDs
	if len(shardsToSearch) == 0 {
		// Search all shards
		shardsToSearch = make([]int, dss.shardCount)
		for i := 0; i < dss.shardCount; i++ {
			shardsToSearch[i] = i
		}
	}
	
	result := &DistributedSearchResult{
		ShardResults: make(map[int][]map[string]interface{}),
		TotalShards:  dss.shardCount,
		QueriedShards: len(shardsToSearch),
	}
	
	// Search each shard in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	for _, shardID := range shardsToSearch {
		wg.Add(1)
		go func(shard int) {
			defer wg.Done()
			
			shardResults, err := dss.searchShard(ctx, req, shard)
			if err != nil {
				// Log error but continue
				return
			}
			
			mu.Lock()
			result.ShardResults[shard] = shardResults
			result.Results = append(result.Results, shardResults...)
			mu.Unlock()
		}(shardID)
	}
	
	wg.Wait()
	
	// Merge and rank results
	result.Results = dss.mergeAndRankResults(result.Results, req.Limit, req.Threshold)
	
	return result, nil
}

/* searchShard searches a specific shard */
func (dss *DistributedSearchService) searchShard(ctx context.Context, req DistributedSearchRequest, shardID int) ([]map[string]interface{}, error) {
	// Build shard-specific table name
	shardTableName := fmt.Sprintf("%s_shard_%d", req.TableName, shardID)
	
	// Perform vector search on shard
	query := fmt.Sprintf(`
		SELECT *, 1 - (%s <=> $1::vector) as similarity
		FROM %s
		WHERE %s IS NOT NULL
			AND 1 - (%s <=> $1::vector) >= $2
		ORDER BY %s <=> $1::vector
		LIMIT $3
	`, req.EmbeddingColumn, shardTableName, req.EmbeddingColumn, req.EmbeddingColumn, req.EmbeddingColumn)
	
	rows, err := dss.pool.Query(ctx, query, req.QueryEmbedding, req.Threshold, req.Limit)
	if err != nil {
		// Shard might not exist, return empty results
		return []map[string]interface{}{}, nil
	}
	defer rows.Close()
	
	var results []map[string]interface{}
	fieldDescriptions := rows.FieldDescriptions()
	
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			continue
		}
		
		row := make(map[string]interface{})
		for i, desc := range fieldDescriptions {
			row[desc.Name] = values[i]
		}
		results = append(results, row)
	}
	
	return results, nil
}

/* mergeAndRankResults merges results from multiple shards and ranks them */
func (dss *DistributedSearchService) mergeAndRankResults(results []map[string]interface{}, limit int, threshold float64) []map[string]interface{} {
	// Remove duplicates and sort by similarity
	seen := make(map[string]bool)
	uniqueResults := []map[string]interface{}{}
	
	for _, result := range results {
		// Use ID as unique key
		if id, ok := result["id"].(string); ok {
			if !seen[id] {
				seen[id] = true
				uniqueResults = append(uniqueResults, result)
			}
		} else if id, ok := result["id"].(fmt.Stringer); ok {
			idStr := id.String()
			if !seen[idStr] {
				seen[idStr] = true
				uniqueResults = append(uniqueResults, result)
			}
		} else {
			uniqueResults = append(uniqueResults, result)
		}
	}
	
	// Sort by similarity (descending)
	// Simple bubble sort for now
	for i := 0; i < len(uniqueResults)-1; i++ {
		for j := i + 1; j < len(uniqueResults); j++ {
			simI, okI := getSimilarity(uniqueResults[i])
			simJ, okJ := getSimilarity(uniqueResults[j])
			
			if okI && okJ && simI < simJ {
				uniqueResults[i], uniqueResults[j] = uniqueResults[j], uniqueResults[i]
			}
		}
	}
	
	// Apply limit
	if limit > 0 && len(uniqueResults) > limit {
		uniqueResults = uniqueResults[:limit]
	}
	
	return uniqueResults
}

/* getSimilarity extracts similarity score from result */
func getSimilarity(result map[string]interface{}) (float64, bool) {
	if sim, ok := result["similarity"].(float64); ok {
		return sim, true
	}
	if sim, ok := result["similarity"].(float32); ok {
		return float64(sim), true
	}
	return 0.0, false
}

/* CreateShard creates a new search shard */
func (dss *DistributedSearchService) CreateShard(ctx context.Context, tableName string, shardID int) error {
	shardTableName := fmt.Sprintf("%s_shard_%d", tableName, shardID)
	
	// Create shard table (simplified - in production, would copy schema)
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (LIKE %s INCLUDING ALL)
	`, shardTableName, tableName)
	
	_, err := dss.pool.Exec(ctx, query)
	return err
}
