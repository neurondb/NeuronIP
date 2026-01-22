package warehouse

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* CacheService provides query result caching */
type CacheService struct {
	pool *pgxpool.Pool
}

/* NewCacheService creates a new cache service */
func NewCacheService(pool *pgxpool.Pool) *CacheService {
	return &CacheService{pool: pool}
}

/* CacheEntry represents a cached query result */
type CacheEntry struct {
	ID           uuid.UUID              `json:"id"`
	CacheKey     string                 `json:"cache_key"`
	QueryHash    string                 `json:"query_hash"`
	QueryText    string                 `json:"query_text"`
	ResultData   []map[string]interface{} `json:"result_data"`
	TTL          time.Duration           `json:"ttl"`
	ExpiresAt    time.Time              `json:"expires_at"`
	HitCount     int                    `json:"hit_count"`
	CreatedAt    time.Time              `json:"created_at"`
	LastAccessedAt *time.Time           `json:"last_accessed_at,omitempty"`
}

/* GetCacheKey generates a cache key from query and parameters */
func (s *CacheService) GetCacheKey(queryText string, params map[string]interface{}) string {
	data := map[string]interface{}{
		"query": queryText,
		"params": params,
	}
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

/* GetCachedResult retrieves a cached query result */
func (s *CacheService) GetCachedResult(ctx context.Context, cacheKey string) (*CacheEntry, error) {
	query := `
		SELECT id, cache_key, query_hash, query_text, result_data, expires_at,
		       hit_count, created_at, last_accessed_at
		FROM neuronip.query_cache
		WHERE cache_key = $1 AND expires_at > NOW()
	`
	var entry CacheEntry
	var resultDataJSON []byte
	var lastAccessedAt *time.Time

	err := s.pool.QueryRow(ctx, query, cacheKey).Scan(
		&entry.ID, &entry.CacheKey, &entry.QueryHash, &entry.QueryText,
		&resultDataJSON, &entry.ExpiresAt, &entry.HitCount,
		&entry.CreatedAt, &lastAccessedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("cache miss: %w", err)
	}

	json.Unmarshal(resultDataJSON, &entry.ResultData)
	entry.LastAccessedAt = lastAccessedAt

	// Update hit count and last accessed
	updateQuery := `
		UPDATE neuronip.query_cache
		SET hit_count = hit_count + 1, last_accessed_at = NOW()
		WHERE id = $1
	`
	s.pool.Exec(ctx, updateQuery, entry.ID)

	return &entry, nil
}

/* SetCachedResult stores a query result in cache */
func (s *CacheService) SetCachedResult(ctx context.Context, queryText string, params map[string]interface{}, results []map[string]interface{}, ttl time.Duration) error {
	cacheKey := s.GetCacheKey(queryText, params)
	queryHash := s.GetCacheKey(queryText, nil) // Hash without params

	resultDataJSON, _ := json.Marshal(results)
	expiresAt := time.Now().Add(ttl)

	query := `
		INSERT INTO neuronip.query_cache (
			cache_key, query_hash, query_text, result_data, ttl_seconds,
			expires_at, hit_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, 0, NOW())
		ON CONFLICT (cache_key) DO UPDATE SET
			result_data = $4,
			ttl_seconds = $5,
			expires_at = $6,
			hit_count = 0
	`
	_, err := s.pool.Exec(ctx, query,
		cacheKey, queryHash, queryText, resultDataJSON,
		int(ttl.Seconds()), expiresAt,
	)
	return err
}

/* InvalidateCache invalidates cache entries matching criteria */
func (s *CacheService) InvalidateCache(ctx context.Context, invalidationRule InvalidationRule) error {
	var query string
	var args []interface{}

	switch invalidationRule.Type {
	case "query_hash":
		query = `DELETE FROM neuronip.query_cache WHERE query_hash = $1`
		args = []interface{}{invalidationRule.Value}
	case "schema":
		query = `DELETE FROM neuronip.query_cache WHERE query_text LIKE $1`
		args = []interface{}{fmt.Sprintf("%%FROM %s.%%", invalidationRule.Value)}
	case "table":
		query = `DELETE FROM neuronip.query_cache WHERE query_text LIKE $1`
		args = []interface{}{fmt.Sprintf("%%FROM %s%%", invalidationRule.Value)}
	case "time":
		query = `DELETE FROM neuronip.query_cache WHERE expires_at < NOW()`
		args = []interface{}{}
	default:
		return fmt.Errorf("unknown invalidation rule type: %s", invalidationRule.Type)
	}

	_, err := s.pool.Exec(ctx, query, args...)
	return err
}

/* InvalidationRule represents a cache invalidation rule */
type InvalidationRule struct {
	Type  string // "query_hash", "schema", "table", "time"
	Value string
}

/* GetCacheStats gets cache statistics */
func (s *CacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_entries,
			COUNT(CASE WHEN expires_at > NOW() THEN 1 END) as active_entries,
			SUM(hit_count) as total_hits,
			AVG(hit_count) as avg_hits_per_entry
		FROM neuronip.query_cache
	`
	var totalEntries, activeEntries int
	var totalHits *int
	var avgHits *float64

	err := s.pool.QueryRow(ctx, query).Scan(&totalEntries, &activeEntries, &totalHits, &avgHits)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total_entries": totalEntries,
		"active_entries": activeEntries,
	}
	if totalHits != nil {
		result["total_hits"] = *totalHits
	}
	if avgHits != nil {
		result["avg_hits_per_entry"] = *avgHits
	}

	return result, nil
}

/* WarmCache warms the cache with frequently used queries */
func (s *CacheService) WarmCache(ctx context.Context, queries []string) error {
	// Execute each query and cache the results
	// Default TTL for warmed cache entries is 1 hour
	defaultTTL := 1 * time.Hour
	
	for _, queryText := range queries {
		if queryText == "" {
			continue
		}
		
		// Execute query directly on the database
		rows, err := s.pool.Query(ctx, queryText)
		if err != nil {
			// Log error but continue with other queries
			continue
		}
		
		// Convert rows to result maps
		results := make([]map[string]interface{}, 0)
		columns := rows.FieldDescriptions()
		columnNames := make([]string, len(columns))
		for i, col := range columns {
			columnNames[i] = col.Name
		}
		
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			
			if err := rows.Scan(valuePtrs...); err != nil {
				continue
			}
			
			row := make(map[string]interface{})
			for i, colName := range columnNames {
				val := values[i]
				// Convert database types to JSON-serializable types
				if val != nil {
					row[colName] = val
				} else {
					row[colName] = nil
				}
			}
			results = append(results, row)
		}
		rows.Close()
		
		// Cache the results
		if len(results) > 0 {
			params := make(map[string]interface{})
			if err := s.SetCachedResult(ctx, queryText, params, results, defaultTTL); err != nil {
				// Log error but continue
				continue
			}
		}
	}
	
	return nil
}
