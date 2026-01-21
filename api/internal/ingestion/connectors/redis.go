package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* RedisConnector implements the Connector interface for Redis */
type RedisConnector struct {
	*ingestion.BaseConnector
	client *redis.Client
}

/* NewRedisConnector creates a new Redis connector */
func NewRedisConnector() *RedisConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "redis",
		Name:        "Redis",
		Description: "Redis in-memory data store connector",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("redis", metadata)

	return &RedisConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Redis */
func (r *RedisConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	host, _ := config["host"].(string)
	port, _ := config["port"].(float64)
	password, _ := config["password"].(string)
	database, _ := config["database"].(float64)

	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 6379
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%.0f", host, port),
		Password: password,
		DB:       int(database),
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	r.client = rdb
	r.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (r *RedisConnector) Disconnect(ctx context.Context) error {
	if r.client != nil {
		r.client.Close()
	}
	r.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (r *RedisConnector) TestConnection(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("not connected")
	}
	return r.client.Ping(ctx).Err()
}

/* DiscoverSchema discovers Redis schema (keys) */
func (r *RedisConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if r.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Redis doesn't have traditional schemas, so we organize by key patterns
	tables := []ingestion.TableSchema{}

	// Scan all keys
	iter := r.client.Scan(ctx, 0, "*", 0).Iterator()
	keyPatterns := make(map[string]bool)

	for iter.Next(ctx) {
		key := iter.Val()
		// Extract pattern (prefix before first : or *)
		pattern := r.extractPattern(key)
		keyPatterns[pattern] = true
	}
	// Redis iterator doesn't need explicit close

	// Create "tables" for each key pattern
	for pattern := range keyPatterns {
		// Get sample key to determine type
		var sampleKey string
		iter = r.client.Scan(ctx, 0, pattern+"*", 1).Iterator()
		if iter.Next(ctx) {
			sampleKey = iter.Val()
		}
		// Redis iterator doesn't need explicit close

		dataType := "string"
		if sampleKey != "" {
			keyType, _ := r.client.Type(ctx, sampleKey).Result()
			dataType = keyType
		}

		tables = append(tables, ingestion.TableSchema{
			Name: pattern,
			Columns: []ingestion.ColumnSchema{
				{Name: "key", DataType: "string"},
				{Name: "value", DataType: dataType},
			},
		})
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* extractPattern extracts a pattern from a Redis key */
func (r *RedisConnector) extractPattern(key string) string {
	// Use prefix before first colon or wildcard as pattern
	for i, char := range key {
		if char == ':' || char == '*' {
			return key[:i]
		}
	}
	return key
}

/* Sync performs a sync operation */
func (r *RedisConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if r.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := r.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	tables := options.Tables
	if len(tables) == 0 {
		for _, table := range schema.Tables {
			tables = append(tables, table.Name)
		}
	}

	for _, pattern := range tables {
		// Count keys matching pattern
		iter := r.client.Scan(ctx, 0, pattern+"*", 0).Iterator()
		count := int64(0)
		for iter.Next(ctx) {
			count++
		}
		// Redis iterator doesn't need explicit close

		result.RowsSynced += count
		result.TablesSynced = append(result.TablesSynced, pattern)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
