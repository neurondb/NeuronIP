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

/* HallucinationDetectionService provides hallucination detection and tracking */
type HallucinationDetectionService struct {
	pool *pgxpool.Pool
}

/* NewHallucinationDetectionService creates a new hallucination detection service */
func NewHallucinationDetectionService(pool *pgxpool.Pool) *HallucinationDetectionService {
	return &HallucinationDetectionService{pool: pool}
}

/* HallucinationSignal represents a hallucination risk signal */
type HallucinationSignal struct {
	ID                uuid.UUID              `json:"id"`
	QueryID           *uuid.UUID              `json:"query_id,omitempty"`
	AgentRunID        *uuid.UUID              `json:"agent_run_id,omitempty"`
	ResponseID        *uuid.UUID              `json:"response_id,omitempty"`
	ConfidenceScore   float64                 `json:"confidence_score"`
	RiskLevel         string                 `json:"risk_level"` // low, medium, high, critical
	CitationAccuracy  float64                 `json:"citation_accuracy"` // 0-1, how accurate citations are
	EvidenceStrength  float64                 `json:"evidence_strength"` // 0-1, how strong supporting evidence is
	Flags             []string               `json:"flags,omitempty"` // low_confidence, missing_citations, weak_evidence, etc.
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

/* RecordHallucinationSignal records a hallucination risk signal */
func (s *HallucinationDetectionService) RecordHallucinationSignal(ctx context.Context, queryID *uuid.UUID, agentRunID *uuid.UUID, responseID *uuid.UUID, confidenceScore float64, citationAccuracy float64, evidenceStrength float64, flags []string, metadata map[string]interface{}) (*HallucinationSignal, error) {
	id := uuid.New()
	metadataJSON, _ := json.Marshal(metadata)
	flagsJSON, _ := json.Marshal(flags)

	// Determine risk level
	riskLevel := "low"
	if confidenceScore < 0.5 || citationAccuracy < 0.5 || evidenceStrength < 0.5 {
		riskLevel = "critical"
	} else if confidenceScore < 0.7 || citationAccuracy < 0.7 || evidenceStrength < 0.7 {
		riskLevel = "high"
	} else if confidenceScore < 0.85 || citationAccuracy < 0.85 || evidenceStrength < 0.85 {
		riskLevel = "medium"
	}

	query := `
		INSERT INTO neuronip.hallucination_signals 
		(id, query_id, agent_run_id, response_id, confidence_score, risk_level,
		 citation_accuracy, evidence_strength, flags, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		RETURNING id, query_id, agent_run_id, response_id, confidence_score, risk_level,
		          citation_accuracy, evidence_strength, flags, metadata, created_at`

	var signal HallucinationSignal
	var queryIDVal, agentRunIDVal, responseIDVal sql.NullString
	var flagsJSONRaw, metadataJSONRaw json.RawMessage

	err := s.pool.QueryRow(ctx, query, id, queryID, agentRunID, responseID,
		confidenceScore, riskLevel, citationAccuracy, evidenceStrength, flagsJSON, metadataJSON).Scan(
		&signal.ID, &queryIDVal, &agentRunIDVal, &responseIDVal,
		&signal.ConfidenceScore, &signal.RiskLevel, &signal.CitationAccuracy,
		&signal.EvidenceStrength, &flagsJSONRaw, &metadataJSONRaw, &signal.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record hallucination signal: %w", err)
	}

	if queryIDVal.Valid {
		qID, _ := uuid.Parse(queryIDVal.String)
		signal.QueryID = &qID
	}
	if agentRunIDVal.Valid {
		aID, _ := uuid.Parse(agentRunIDVal.String)
		signal.AgentRunID = &aID
	}
	if responseIDVal.Valid {
		rID, _ := uuid.Parse(responseIDVal.String)
		signal.ResponseID = &rID
	}
	if flagsJSONRaw != nil {
		json.Unmarshal(flagsJSONRaw, &signal.Flags)
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &signal.Metadata)
	}

	return &signal, nil
}

/* GetHallucinationSignals retrieves hallucination signals */
func (s *HallucinationDetectionService) GetHallucinationSignals(ctx context.Context, queryID *uuid.UUID, agentRunID *uuid.UUID, riskLevel *string, limit int) ([]HallucinationSignal, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, query_id, agent_run_id, response_id, confidence_score, risk_level,
		       citation_accuracy, evidence_strength, flags, metadata, created_at
		FROM neuronip.hallucination_signals
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

	if riskLevel != nil {
		query += fmt.Sprintf(" AND risk_level = $%d", argIndex)
		args = append(args, *riskLevel)
		argIndex++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get hallucination signals: %w", err)
	}
	defer rows.Close()

	var signals []HallucinationSignal
	for rows.Next() {
		var signal HallucinationSignal
		var queryIDVal, agentRunIDVal, responseIDVal sql.NullString
		var flagsJSONRaw, metadataJSONRaw json.RawMessage

		err := rows.Scan(
			&signal.ID, &queryIDVal, &agentRunIDVal, &responseIDVal,
			&signal.ConfidenceScore, &signal.RiskLevel, &signal.CitationAccuracy,
			&signal.EvidenceStrength, &flagsJSONRaw, &metadataJSONRaw, &signal.CreatedAt,
		)
		if err != nil {
			continue
		}

		if queryIDVal.Valid {
			qID, _ := uuid.Parse(queryIDVal.String)
			signal.QueryID = &qID
		}
		if agentRunIDVal.Valid {
			aID, _ := uuid.Parse(agentRunIDVal.String)
			signal.AgentRunID = &aID
		}
		if responseIDVal.Valid {
			rID, _ := uuid.Parse(responseIDVal.String)
			signal.ResponseID = &rID
		}
		if flagsJSONRaw != nil {
			json.Unmarshal(flagsJSONRaw, &signal.Flags)
		}
		if metadataJSONRaw != nil {
			json.Unmarshal(metadataJSONRaw, &signal.Metadata)
		}

		signals = append(signals, signal)
	}

	return signals, nil
}

/* GetHallucinationStats gets aggregated hallucination statistics */
func (s *HallucinationDetectionService) GetHallucinationStats(ctx context.Context, timeRange string) (map[string]interface{}, error) {
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
			COUNT(*) as total_signals,
			AVG(confidence_score) as avg_confidence,
			AVG(citation_accuracy) as avg_citation_accuracy,
			AVG(evidence_strength) as avg_evidence_strength,
			SUM(CASE WHEN risk_level = 'critical' THEN 1 ELSE 0 END) as critical_count,
			SUM(CASE WHEN risk_level = 'high' THEN 1 ELSE 0 END) as high_count,
			SUM(CASE WHEN risk_level = 'medium' THEN 1 ELSE 0 END) as medium_count,
			SUM(CASE WHEN risk_level = 'low' THEN 1 ELSE 0 END) as low_count
		FROM neuronip.hallucination_signals
		WHERE created_at > NOW() - INTERVAL '%s'`, interval)

	var stats map[string]interface{}
	var totalSignals sql.NullInt64
	var avgConfidence, avgCitationAccuracy, avgEvidenceStrength sql.NullFloat64
	var criticalCount, highCount, mediumCount, lowCount sql.NullInt64

	err := s.pool.QueryRow(ctx, query).Scan(
		&totalSignals, &avgConfidence, &avgCitationAccuracy, &avgEvidenceStrength,
		&criticalCount, &highCount, &mediumCount, &lowCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get hallucination stats: %w", err)
	}

	stats = make(map[string]interface{})
	if totalSignals.Valid {
		stats["total_signals"] = totalSignals.Int64
	}
	if avgConfidence.Valid {
		stats["avg_confidence"] = avgConfidence.Float64
	}
	if avgCitationAccuracy.Valid {
		stats["avg_citation_accuracy"] = avgCitationAccuracy.Float64
	}
	if avgEvidenceStrength.Valid {
		stats["avg_evidence_strength"] = avgEvidenceStrength.Float64
	}
	if criticalCount.Valid {
		stats["critical_count"] = criticalCount.Int64
	}
	if highCount.Valid {
		stats["high_count"] = highCount.Int64
	}
	if mediumCount.Valid {
		stats["medium_count"] = mediumCount.Int64
	}
	if lowCount.Valid {
		stats["low_count"] = lowCount.Int64
	}

	return stats, nil
}
