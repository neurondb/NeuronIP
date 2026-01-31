package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* EvidenceTracker provides evidence coverage tracking for agent responses */
type EvidenceTracker struct {
	pool *pgxpool.Pool
}

/* NewEvidenceTracker creates a new evidence tracker */
func NewEvidenceTracker(pool *pgxpool.Pool) *EvidenceTracker {
	return &EvidenceTracker{pool: pool}
}

/* EvidenceCoverage represents evidence coverage for an agent response */
type EvidenceCoverage struct {
	ID                uuid.UUID              `json:"id"`
	TraceID           uuid.UUID              `json:"trace_id"`
	ResponseID        *uuid.UUID             `json:"response_id,omitempty"`
	TotalClaims       int                    `json:"total_claims"`
	SupportedClaims   int                    `json:"supported_claims"`
	UnsupportedClaims int                    `json:"unsupported_claims"`
	CoverageScore     float64                `json:"coverage_score"` // 0.0 to 1.0
	EvidenceSources   []EvidenceSource        `json:"evidence_sources"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	EvaluatedAt       time.Time              `json:"evaluated_at"`
}

/* EvidenceSource represents a source of evidence */
type EvidenceSource struct {
	ID          uuid.UUID              `json:"id"`
	SourceType  string                 `json:"source_type"` // "document", "database", "api", "knowledge_graph"
	SourceID    string                 `json:"source_id"`
	SourceName  string                 `json:"source_name"`
	Relevance   float64                `json:"relevance"` // 0.0 to 1.0
	Claims      []string               `json:"claims,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

/* TrackEvidence tracks evidence coverage for an agent response */
func (et *EvidenceTracker) TrackEvidence(ctx context.Context, traceID uuid.UUID, responseText string, evidenceSources []EvidenceSource) (*EvidenceCoverage, error) {
	coverage := &EvidenceCoverage{
		ID:              uuid.New(),
		TraceID:         traceID,
		EvidenceSources: evidenceSources,
		EvaluatedAt:      time.Now(),
		Metadata:         make(map[string]interface{}),
	}
	
	// Extract claims from response
	claims := et.extractClaims(responseText)
	coverage.TotalClaims = len(claims)
	
	// Match claims to evidence sources
	supportedClaims := et.matchClaimsToEvidence(claims, evidenceSources)
	coverage.SupportedClaims = len(supportedClaims)
	coverage.UnsupportedClaims = coverage.TotalClaims - coverage.SupportedClaims
	
	// Calculate coverage score
	if coverage.TotalClaims > 0 {
		coverage.CoverageScore = float64(coverage.SupportedClaims) / float64(coverage.TotalClaims)
	} else {
		coverage.CoverageScore = 0.0
	}
	
	// Store coverage
	if err := et.storeCoverage(ctx, coverage); err != nil {
		return nil, fmt.Errorf("failed to store coverage: %w", err)
	}
	
	return coverage, nil
}

/* extractClaims extracts claims from response text */
func (et *EvidenceTracker) extractClaims(text string) []string {
	// Simple claim extraction - in production, use NLP
	
	// Look for statements (sentences ending with period)
	sentences := []string{}
	current := ""
	for _, char := range text {
		if char == '.' || char == '!' || char == '?' {
			current += string(char)
			if len(current) > 10 { // Minimum claim length
				sentences = append(sentences, current)
			}
			current = ""
		} else {
			current += string(char)
		}
	}
	
	return sentences
}

/* matchClaimsToEvidence matches claims to evidence sources */
func (et *EvidenceTracker) matchClaimsToEvidence(claims []string, sources []EvidenceSource) []string {
	supported := []string{}
	
	for _, claim := range claims {
		for _, source := range sources {
			if source.Relevance > 0.5 {
				// Check if source supports this claim
				if et.sourceSupportsClaim(claim, source) {
					supported = append(supported, claim)
					break
				}
			}
		}
	}
	
	return supported
}

/* sourceSupportsClaim checks if a source supports a claim */
func (et *EvidenceTracker) sourceSupportsClaim(claim string, source EvidenceSource) bool {
	// Simple keyword matching - in production, use semantic similarity
	claimLower := strings.ToLower(claim)
	sourceNameLower := strings.ToLower(source.SourceName)
	
	// Check if source name contains keywords from claim
	claimWords := splitWords(claimLower)
	for _, word := range claimWords {
		if len(word) > 3 && strings.Contains(sourceNameLower, word) {
			return true
		}
	}
	
	return false
}

/* storeCoverage stores evidence coverage */
func (et *EvidenceTracker) storeCoverage(ctx context.Context, coverage *EvidenceCoverage) error {
	sourcesJSON, _ := json.Marshal(coverage.EvidenceSources)
	metadataJSON, _ := json.Marshal(coverage.Metadata)
	
	query := `
		INSERT INTO neuronip.agent_evidence_coverage 
		(id, trace_id, response_id, total_claims, supported_claims, unsupported_claims, coverage_score, evidence_sources, metadata, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	
	_, err := et.pool.Exec(ctx, query,
		coverage.ID, coverage.TraceID, coverage.ResponseID, coverage.TotalClaims,
		coverage.SupportedClaims, coverage.UnsupportedClaims, coverage.CoverageScore,
		sourcesJSON, metadataJSON, coverage.EvaluatedAt,
	)
	return err
}

/* GetCoverage retrieves evidence coverage for a trace */
func (et *EvidenceTracker) GetCoverage(ctx context.Context, traceID uuid.UUID) (*EvidenceCoverage, error) {
	query := `
		SELECT id, trace_id, response_id, total_claims, supported_claims, unsupported_claims, coverage_score, evidence_sources, metadata, evaluated_at
		FROM neuronip.agent_evidence_coverage
		WHERE trace_id = $1
		ORDER BY evaluated_at DESC
		LIMIT 1
	`
	
	var coverage EvidenceCoverage
	var responseID *uuid.UUID
	var sourcesJSON, metadataJSON json.RawMessage
	
	err := et.pool.QueryRow(ctx, query, traceID).Scan(
		&coverage.ID, &coverage.TraceID, &responseID, &coverage.TotalClaims,
		&coverage.SupportedClaims, &coverage.UnsupportedClaims, &coverage.CoverageScore,
		&sourcesJSON, &metadataJSON, &coverage.EvaluatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage: %w", err)
	}
	
	coverage.ResponseID = responseID
	if sourcesJSON != nil {
		json.Unmarshal(sourcesJSON, &coverage.EvidenceSources)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &coverage.Metadata)
	}
	
	return &coverage, nil
}

// Helper functions - using strings package
func splitWords(s string) []string {
	words := []string{}
	current := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			current += string(r)
		} else {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
