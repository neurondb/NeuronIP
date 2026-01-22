package observability

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* RetrievalMetricsService provides retrieval metrics tracking */
type RetrievalMetricsService struct {
	pool *pgxpool.Pool
}

/* NewRetrievalMetricsService creates a new retrieval metrics service */
func NewRetrievalMetricsService(pool *pgxpool.Pool) *RetrievalMetricsService {
	return &RetrievalMetricsService{pool: pool}
}

/* RetrievalMetric represents a retrieval metric */
type RetrievalMetric struct {
	ID                uuid.UUID              `json:"id"`
	QueryID           *uuid.UUID              `json:"query_id,omitempty"`
	AgentRunID        *uuid.UUID              `json:"agent_run_id,omitempty"`
	RetrievalType     string                 `json:"retrieval_type"` // semantic, keyword, hybrid
	DocumentsRetrieved int                   `json:"documents_retrieved"`
	DocumentsUsed     int                   `json:"documents_used"`
	HitRate           float64                `json:"hit_rate"` // documents_used / documents_retrieved
	EvidenceCoverage  float64                `json:"evidence_coverage"` // percentage of query covered by evidence
	AvgSimilarity     float64                `json:"avg_similarity"`
	RetrievalLatencyMs int64                 `json:"retrieval_latency_ms"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

/* RecordRetrieval records a retrieval metric */
func (s *RetrievalMetricsService) RecordRetrieval(ctx context.Context, queryID *uuid.UUID, agentRunID *uuid.UUID, retrievalType string, documentsRetrieved int, documentsUsed int, avgSimilarity float64, latencyMs int64, metadata map[string]interface{}) error {
	id := uuid.New()
	metadataJSON, _ := json.Marshal(metadata)

	hitRate := 0.0
	if documentsRetrieved > 0 {
		hitRate = float64(documentsUsed) / float64(documentsRetrieved)
	}

	// Calculate evidence coverage (simplified - in production would analyze content overlap)
	evidenceCoverage := hitRate * 0.8 // Simplified calculation

	query := `
		INSERT INTO neuronip.retrieval_metrics 
		(id, query_id, agent_run_id, retrieval_type, documents_retrieved, documents_used, 
		 hit_rate, evidence_coverage, avg_similarity, retrieval_latency_ms, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())`

	_, err := s.pool.Exec(ctx, query, id, queryID, agentRunID, retrievalType,
		documentsRetrieved, documentsUsed, hitRate, evidenceCoverage, avgSimilarity, latencyMs, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to record retrieval: %w", err)
	}

	return nil
}

/* GetRetrievalMetrics retrieves retrieval metrics */
func (s *RetrievalMetricsService) GetRetrievalMetrics(ctx context.Context, queryID *uuid.UUID, agentRunID *uuid.UUID, limit int) ([]RetrievalMetric, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, query_id, agent_run_id, retrieval_type, documents_retrieved, documents_used,
		       hit_rate, evidence_coverage, avg_similarity, retrieval_latency_ms, metadata, created_at
		FROM neuronip.retrieval_metrics
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if queryID != nil {
		query += fmt.Sprintf(" AND query_id = $%d", argIndex)
		args = append(args, *queryID)
		argIndex++
	}

	if agentRunID != nil {
		query += fmt.Sprintf(" AND agent_run_id = $%d", argIndex)
		args = append(args, *agentRunID)
		argIndex++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get retrieval metrics: %w", err)
	}
	defer rows.Close()

	var metrics []RetrievalMetric
	for rows.Next() {
		var metric RetrievalMetric
		var queryIDVal, agentRunIDVal sql.NullString
		var metadataJSON json.RawMessage

		err := rows.Scan(
			&metric.ID, &queryIDVal, &agentRunIDVal, &metric.RetrievalType,
			&metric.DocumentsRetrieved, &metric.DocumentsUsed, &metric.HitRate,
			&metric.EvidenceCoverage, &metric.AvgSimilarity, &metric.RetrievalLatencyMs,
			&metadataJSON, &metric.CreatedAt,
		)
		if err != nil {
			continue
		}

		if queryIDVal.Valid {
			qID, _ := uuid.Parse(queryIDVal.String)
			metric.QueryID = &qID
		}
		if agentRunIDVal.Valid {
			aID, _ := uuid.Parse(agentRunIDVal.String)
			metric.AgentRunID = &aID
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &metric.Metadata)
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

/* GetRetrievalStats gets aggregated retrieval statistics */
func (s *RetrievalMetricsService) GetRetrievalStats(ctx context.Context, timeRange string) (map[string]interface{}, error) {
	var interval string
	switch timeRange {
	case "1h", "1hour":
		interval = "1 hour"
	case "24h", "1day":
		interval = "24 hours"
	case "7d", "1week":
		interval = "7 days"
	default:
		interval = "24 hours"
	}

	query := fmt.Sprintf(`
		SELECT 
			AVG(hit_rate) as avg_hit_rate,
			AVG(evidence_coverage) as avg_evidence_coverage,
			AVG(avg_similarity) as avg_similarity,
			AVG(retrieval_latency_ms) as avg_latency_ms,
			COUNT(*) as total_retrievals,
			SUM(CASE WHEN hit_rate < 0.5 THEN 1 ELSE 0 END) as low_hit_rate_count
		FROM neuronip.retrieval_metrics
		WHERE created_at > NOW() - INTERVAL '%s'`, interval)

	var stats map[string]interface{}
	var avgHitRate, avgEvidenceCoverage, avgSimilarity, avgLatencyMs sql.NullFloat64
	var totalRetrievals, lowHitRateCount sql.NullInt64

	err := s.pool.QueryRow(ctx, query).Scan(
		&avgHitRate, &avgEvidenceCoverage, &avgSimilarity, &avgLatencyMs,
		&totalRetrievals, &lowHitRateCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get retrieval stats: %w", err)
	}

	stats = make(map[string]interface{})
	if avgHitRate.Valid {
		stats["avg_hit_rate"] = avgHitRate.Float64
	}
	if avgEvidenceCoverage.Valid {
		stats["avg_evidence_coverage"] = avgEvidenceCoverage.Float64
	}
	if avgSimilarity.Valid {
		stats["avg_similarity"] = avgSimilarity.Float64
	}
	if avgLatencyMs.Valid {
		stats["avg_latency_ms"] = avgLatencyMs.Float64
	}
	if totalRetrievals.Valid {
		stats["total_retrievals"] = totalRetrievals.Int64
	}
	if lowHitRateCount.Valid {
		stats["low_hit_rate_count"] = lowHitRateCount.Int64
	}

	return stats, nil
}
