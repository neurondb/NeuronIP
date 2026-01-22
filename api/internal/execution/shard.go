package execution

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ShardService provides sharding functionality */
type ShardService struct {
	pool *pgxpool.Pool
}

/* NewShardService creates a new shard service */
func NewShardService(pool *pgxpool.Pool) *ShardService {
	return &ShardService{pool: pool}
}

/* QueryShard represents a query shard configuration */
type QueryShard struct {
	ID            uuid.UUID              `json:"id"`
	TableName     string                 `json:"table_name"`
	SchemaName    string                 `json:"schema_name"`
	ShardKey      string                 `json:"shard_key"`
	ShardCount    int                    `json:"shard_count"`
	ShardStrategy string                 `json:"shard_strategy"` // hash, range, list
	ShardConfig   map[string]interface{} `json:"shard_config,omitempty"`
	Enabled       bool                   `json:"enabled"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

/* CreateShard creates a shard configuration */
func (s *ShardService) CreateShard(ctx context.Context, schemaName, tableName, shardKey string, shardCount int, shardStrategy string, shardConfig map[string]interface{}) (*QueryShard, error) {
	id := uuid.New()
	configJSON, _ := json.Marshal(shardConfig)
	now := "NOW()"

	query := `
		INSERT INTO neuronip.query_shards 
		(id, schema_name, table_name, shard_key, shard_count, shard_strategy, shard_config, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8, $9)
		ON CONFLICT (schema_name, table_name) 
		DO UPDATE SET 
			shard_key = EXCLUDED.shard_key,
			shard_count = EXCLUDED.shard_count,
			shard_strategy = EXCLUDED.shard_strategy,
			shard_config = EXCLUDED.shard_config,
			updated_at = EXCLUDED.updated_at
		RETURNING id, schema_name, table_name, shard_key, shard_count, shard_strategy, shard_config, enabled, created_at, updated_at`

	var shard QueryShard
	var configJSONRaw json.RawMessage
	var createdAt, updatedAt string

	err := s.pool.QueryRow(ctx, query, id, schemaName, tableName, shardKey, shardCount, shardStrategy, configJSON, now, now).Scan(
		&shard.ID, &shard.SchemaName, &shard.TableName, &shard.ShardKey,
		&shard.ShardCount, &shard.ShardStrategy, &configJSONRaw, &shard.Enabled,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shard: %w", err)
	}

	if configJSONRaw != nil {
		json.Unmarshal(configJSONRaw, &shard.ShardConfig)
	}
	shard.CreatedAt = createdAt
	shard.UpdatedAt = updatedAt

	return &shard, nil
}

/* GetShard gets shard configuration for a table */
func (s *ShardService) GetShard(ctx context.Context, schemaName, tableName string) (*QueryShard, error) {
	query := `
		SELECT id, schema_name, table_name, shard_key, shard_count, shard_strategy, shard_config, enabled, created_at, updated_at
		FROM neuronip.query_shards
		WHERE schema_name = $1 AND table_name = $2 AND enabled = true`

	var shard QueryShard
	var configJSONRaw json.RawMessage
	var createdAt, updatedAt string

	err := s.pool.QueryRow(ctx, query, schemaName, tableName).Scan(
		&shard.ID, &shard.SchemaName, &shard.TableName, &shard.ShardKey,
		&shard.ShardCount, &shard.ShardStrategy, &configJSONRaw, &shard.Enabled,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No sharding configured
		}
		return nil, fmt.Errorf("failed to get shard: %w", err)
	}

	if configJSONRaw != nil {
		json.Unmarshal(configJSONRaw, &shard.ShardConfig)
	}
	shard.CreatedAt = createdAt
	shard.UpdatedAt = updatedAt

	return &shard, nil
}

/* RouteToShard routes a query to the appropriate shard based on shard key value */
func (s *ShardService) RouteToShard(ctx context.Context, schemaName, tableName, shardKeyValue string) (int, error) {
	shard, err := s.GetShard(ctx, schemaName, tableName)
	if err != nil {
		return -1, err
	}
	if shard == nil {
		return -1, nil // No sharding
	}

	switch shard.ShardStrategy {
	case "hash":
		return s.hashShard(shardKeyValue, shard.ShardCount), nil
	case "range":
		return s.rangeShard(shardKeyValue, shard), nil
	case "list":
		return s.listShard(shardKeyValue, shard), nil
	default:
		return -1, fmt.Errorf("unknown shard strategy: %s", shard.ShardStrategy)
	}
}

/* hashShard routes to shard using hash */
func (s *ShardService) hashShard(shardKeyValue string, shardCount int) int {
	h := fnv.New32a()
	h.Write([]byte(shardKeyValue))
	hash := h.Sum32()
	return int(hash % uint32(shardCount))
}

/* rangeShard routes to shard using range */
func (s *ShardService) rangeShard(shardKeyValue string, shard *QueryShard) int {
	// Get range configuration from shard config
	ranges, ok := shard.ShardConfig["ranges"].([]interface{})
	if !ok || len(ranges) == 0 {
		// Default: divide evenly
		return s.hashShard(shardKeyValue, shard.ShardCount)
	}

	// Find which range the value falls into
	for i, rangeItem := range ranges {
		rangeMap, ok := rangeItem.(map[string]interface{})
		if !ok {
			continue
		}

		min, ok1 := rangeMap["min"].(string)
		max, ok2 := rangeMap["max"].(string)
		if !ok1 || !ok2 {
			continue
		}

		if shardKeyValue >= min && shardKeyValue <= max {
			return i
		}
	}

	// Default to last shard
	return shard.ShardCount - 1
}

/* listShard routes to shard using list */
func (s *ShardService) listShard(shardKeyValue string, shard *QueryShard) int {
	// Get list configuration from shard config
	lists, ok := shard.ShardConfig["lists"].([]interface{})
	if !ok || len(lists) == 0 {
		// Default: hash
		return s.hashShard(shardKeyValue, shard.ShardCount)
	}

	// Find which list contains the value
	for i, listItem := range lists {
		list, ok := listItem.([]interface{})
		if !ok {
			continue
		}

		for _, item := range list {
			itemStr := fmt.Sprintf("%v", item)
			if itemStr == shardKeyValue {
				return i
			}
		}
	}

	// Default to first shard
	return 0
}

/* BuildShardedQuery builds a query for a specific shard */
func (s *ShardService) BuildShardedQuery(query string, shardIndex int, shard *QueryShard) string {
	// Replace table name with sharded table name
	// In production, actual sharded tables would be named like: table_name_shard_0, table_name_shard_1, etc.
	shardTableName := fmt.Sprintf("%s_shard_%d", shard.TableName, shardIndex)
	
	// Replace table reference in query
	oldTableRef := fmt.Sprintf("%s.%s", shard.SchemaName, shard.TableName)
	newTableRef := fmt.Sprintf("%s.%s", shard.SchemaName, shardTableName)
	
	return strings.ReplaceAll(query, oldTableRef, newTableRef)
}

/* AggregateShardResults aggregates results from multiple shards */
func (s *ShardService) AggregateShardResults(results []map[string]interface{}) (map[string]interface{}, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to aggregate")
	}

	// Simple aggregation: merge all results
	aggregated := make(map[string]interface{})
	for _, result := range results {
		for k, v := range result {
			if existing, exists := aggregated[k]; exists {
				// Handle numeric aggregation
				if num, ok := v.(float64); ok {
					if existingNum, ok := existing.(float64); ok {
						aggregated[k] = existingNum + num
					} else {
						aggregated[k] = num
					}
				} else {
					// For non-numeric, keep the first value
					if _, exists := aggregated[k]; !exists {
						aggregated[k] = v
					}
				}
			} else {
				aggregated[k] = v
			}
		}
	}

	return aggregated, nil
}
