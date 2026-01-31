package cluster

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ShardingManager manages data sharding strategies */
type ShardingManager struct {
	pool *pgxpool.Pool
}

/* NewShardingManager creates a new sharding manager */
func NewShardingManager(pool *pgxpool.Pool) *ShardingManager {
	return &ShardingManager{pool: pool}
}

/* CreateShard creates a new shard */
func (s *ShardingManager) CreateShard(ctx context.Context, shardKey string, shardType string, nodeID string, config map[string]interface{}) (*Shard, error) {
	shardID := uuid.New()
	
	var rangeStart, rangeEnd, directoryPath interface{}
	if config != nil {
		rangeStart = config["range_start"]
		rangeEnd = config["range_end"]
		directoryPath = config["directory_path"]
	}

	query := `
		INSERT INTO neuronip.cluster_shards 
		(id, shard_key, shard_type, node_id, range_start, range_end, directory_path, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'active', NOW(), NOW())
		RETURNING id, shard_key, shard_type, node_id, range_start, range_end, directory_path, status, created_at, updated_at`

	var shard Shard
	err := s.pool.QueryRow(ctx, query, shardID, shardKey, shardType, nodeID, rangeStart, rangeEnd, directoryPath).Scan(
		&shard.ID, &shard.ShardKey, &shard.ShardType, &shard.NodeID,
		&shard.RangeStart, &shard.RangeEnd, &shard.DirectoryPath,
		&shard.Status, &shard.CreatedAt, &shard.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shard: %w", err)
	}

	return &shard, nil
}

/* GetShardForKey determines which shard a key belongs to */
func (s *ShardingManager) GetShardForKey(ctx context.Context, shardKey string, key string, shardType string) (*Shard, error) {
	var shard Shard
	var query string

	switch shardType {
	case "hash":
		// Hash-based sharding
		hash := s.hashKey(key)
		query = `
			SELECT id, shard_key, shard_type, node_id, range_start, range_end, directory_path, status, created_at, updated_at
			FROM neuronip.cluster_shards
			WHERE shard_key = $1 AND shard_type = 'hash'
			ORDER BY ABS(CAST(range_start AS INTEGER) - $2)
			LIMIT 1`
		err := s.pool.QueryRow(ctx, query, shardKey, hash).Scan(
			&shard.ID, &shard.ShardKey, &shard.ShardType, &shard.NodeID,
			&shard.RangeStart, &shard.RangeEnd, &shard.DirectoryPath,
			&shard.Status, &shard.CreatedAt, &shard.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get shard: %w", err)
		}

	case "range":
		// Range-based sharding
		query = `
			SELECT id, shard_key, shard_type, node_id, range_start, range_end, directory_path, status, created_at, updated_at
			FROM neuronip.cluster_shards
			WHERE shard_key = $1 
				AND shard_type = 'range'
				AND ($2 >= range_start OR range_start IS NULL)
				AND ($2 <= range_end OR range_end IS NULL)
				AND status = 'active'
			LIMIT 1`
		err := s.pool.QueryRow(ctx, query, shardKey, key).Scan(
			&shard.ID, &shard.ShardKey, &shard.ShardType, &shard.NodeID,
			&shard.RangeStart, &shard.RangeEnd, &shard.DirectoryPath,
			&shard.Status, &shard.CreatedAt, &shard.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get shard: %w", err)
		}

	case "directory":
		// Directory-based sharding (exact match)
		query = `
			SELECT id, shard_key, shard_type, node_id, range_start, range_end, directory_path, status, created_at, updated_at
			FROM neuronip.cluster_shards
			WHERE shard_key = $1 
				AND shard_type = 'directory'
				AND directory_path = $2
				AND status = 'active'
			LIMIT 1`
		err := s.pool.QueryRow(ctx, query, shardKey, key).Scan(
			&shard.ID, &shard.ShardKey, &shard.ShardType, &shard.NodeID,
			&shard.RangeStart, &shard.RangeEnd, &shard.DirectoryPath,
			&shard.Status, &shard.CreatedAt, &shard.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get shard: %w", err)
		}

	default:
		return nil, fmt.Errorf("unknown shard type: %s", shardType)
	}

	return &shard, nil
}

/* ListShards lists all shards for a given shard key */
func (s *ShardingManager) ListShards(ctx context.Context, shardKey string) ([]Shard, error) {
	query := `
		SELECT id, shard_key, shard_type, node_id, range_start, range_end, directory_path, status, created_at, updated_at
		FROM neuronip.cluster_shards
		WHERE shard_key = $1
		ORDER BY created_at`

	rows, err := s.pool.Query(ctx, query, shardKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list shards: %w", err)
	}
	defer rows.Close()

	var shards []Shard
	for rows.Next() {
		var shard Shard
		err := rows.Scan(
			&shard.ID, &shard.ShardKey, &shard.ShardType, &shard.NodeID,
			&shard.RangeStart, &shard.RangeEnd, &shard.DirectoryPath,
			&shard.Status, &shard.CreatedAt, &shard.UpdatedAt,
		)
		if err != nil {
			continue
		}
		shards = append(shards, shard)
	}

	return shards, nil
}

/* hashKey hashes a key to a numeric value for hash-based sharding */
func (s *ShardingManager) hashKey(key string) int64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return int64(h.Sum64())
}

/* HashKeySHA256 hashes a key using SHA256 for consistent hashing */
func (s *ShardingManager) HashKeySHA256(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

/* Shard represents a data shard */
type Shard struct {
	ID            uuid.UUID  `json:"id"`
	ShardKey      string     `json:"shard_key"`
	ShardType     string     `json:"shard_type"`
	NodeID        string     `json:"node_id"`
	RangeStart    *string    `json:"range_start,omitempty"`
	RangeEnd      *string    `json:"range_end,omitempty"`
	DirectoryPath *string    `json:"directory_path,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
