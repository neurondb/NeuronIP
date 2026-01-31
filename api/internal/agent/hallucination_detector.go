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

/* HallucinationDetector provides hallucination risk scoring for agent responses */
type HallucinationDetector struct {
	pool *pgxpool.Pool
}

/* NewHallucinationDetector creates a new hallucination detector */
func NewHallucinationDetector(pool *pgxpool.Pool) *HallucinationDetector {
	return &HallucinationDetector{pool: pool}
}

/* HallucinationRisk represents hallucination risk assessment */
type HallucinationRisk struct {
	ID             uuid.UUID              `json:"id"`
	TraceID        uuid.UUID              `json:"trace_id"`
	ResponseID     *uuid.UUID             `json:"response_id,omitempty"`
	RiskScore      float64                `json:"risk_score"` // 0.0 to 1.0, higher = more risk
	RiskFactors    []RiskFactor           `json:"risk_factors"`
	Confidence     float64                `json:"confidence"`
	Recommendation string                 `json:"recommendation"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	EvaluatedAt    time.Time              `json:"evaluated_at"`
}

/* RiskFactor represents a factor contributing to hallucination risk */
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Severity    float64 `json:"severity"` // 0.0 to 1.0
	Description string  `json:"description"`
}

/* DetectHallucination detects hallucination risk in an agent response */
func (hd *HallucinationDetector) DetectHallucination(ctx context.Context, traceID uuid.UUID, responseText string, evidenceCoverage *EvidenceCoverage) (*HallucinationRisk, error) {
	risk := &HallucinationRisk{
		ID:          uuid.New(),
		TraceID:     traceID,
		RiskFactors: []RiskFactor{},
		EvaluatedAt: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Check various risk factors
	coverageRisk := hd.assessCoverageRisk(evidenceCoverage)
	confidenceRisk := hd.assessConfidenceRisk(responseText)
	consistencyRisk := hd.assessConsistencyRisk(ctx, traceID, responseText)
	specificityRisk := hd.assessSpecificityRisk(responseText)

	risk.RiskFactors = []RiskFactor{
		{Factor: "low_evidence_coverage", Severity: coverageRisk, Description: "Response has low evidence coverage"},
		{Factor: "high_confidence_claims", Severity: confidenceRisk, Description: "Response contains high-confidence claims without evidence"},
		{Factor: "inconsistency", Severity: consistencyRisk, Description: "Response is inconsistent with previous responses"},
		{Factor: "low_specificity", Severity: specificityRisk, Description: "Response is vague or non-specific"},
	}

	// Calculate overall risk score (weighted average)
	risk.RiskScore = (coverageRisk*0.4 + confidenceRisk*0.3 + consistencyRisk*0.2 + specificityRisk*0.1)

	// Set confidence
	risk.Confidence = 0.8 // Default confidence

	// Generate recommendation
	if risk.RiskScore < 0.3 {
		risk.Recommendation = "low_risk"
	} else if risk.RiskScore < 0.6 {
		risk.Recommendation = "moderate_risk_review_recommended"
	} else {
		risk.Recommendation = "high_risk_manual_review_required"
	}

	// Store risk assessment
	if err := hd.storeRisk(ctx, risk); err != nil {
		return nil, fmt.Errorf("failed to store risk: %w", err)
	}

	return risk, nil
}

/* assessCoverageRisk assesses risk based on evidence coverage */
func (hd *HallucinationDetector) assessCoverageRisk(coverage *EvidenceCoverage) float64 {
	if coverage == nil {
		return 1.0 // High risk if no coverage data
	}

	// Inverse of coverage score
	return 1.0 - coverage.CoverageScore
}

/* assessConfidenceRisk assesses risk based on confidence language */
func (hd *HallucinationDetector) assessConfidenceRisk(responseText string) float64 {
	// Look for high-confidence language without evidence markers
	highConfidenceWords := []string{"definitely", "certainly", "absolutely", "always", "never", "all", "every"}
	evidenceMarkers := []string{"according to", "based on", "source", "reference", "evidence"}

	hasHighConfidence := false
	hasEvidenceMarker := false

	responseLower := strings.ToLower(responseText)

	for _, word := range highConfidenceWords {
		if strings.Contains(responseLower, word) {
			hasHighConfidence = true
			break
		}
	}

	for _, marker := range evidenceMarkers {
		if strings.Contains(responseLower, marker) {
			hasEvidenceMarker = true
			break
		}
	}

	if hasHighConfidence && !hasEvidenceMarker {
		return 0.7 // Moderate to high risk
	}

	return 0.2 // Low risk
}

/* assessConsistencyRisk assesses risk based on consistency with previous responses */
func (hd *HallucinationDetector) assessConsistencyRisk(ctx context.Context, traceID uuid.UUID, responseText string) float64 {
	// Get previous responses from this trace
	query := `
		SELECT response_text
		FROM neuronip.agent_trace_spans
		WHERE trace_id = $1 AND response_text IS NOT NULL
		ORDER BY start_time DESC
		LIMIT 5`

	rows, err := hd.pool.Query(ctx, query, traceID)
	if err != nil {
		return 0.3 // Default if can't query
	}
	defer rows.Close()

	var previousResponses []string
	for rows.Next() {
		var prevText string
		if err := rows.Scan(&prevText); err != nil {
			continue
		}
		previousResponses = append(previousResponses, prevText)
	}

	if len(previousResponses) == 0 {
		return 0.2 // Low risk for first response
	}

	// Check for contradictions with previous responses
	currentWords := tokenize(strings.ToLower(responseText))
	contradictionCount := 0
	totalComparisons := 0

	negationPairs := map[string]string{
		"is":     "is not",
		"can":    "cannot",
		"will":   "will not",
		"does":   "does not",
		"has":    "has not",
		"true":   "false",
		"yes":    "no",
		"always": "never",
	}

	for _, prev := range previousResponses {
		prevLower := strings.ToLower(prev)
		for positive, negative := range negationPairs {
			hasPosInCurrent := strings.Contains(responseText, positive)
			hasNegInCurrent := strings.Contains(responseText, negative)
			hasPosInPrev := strings.Contains(prevLower, positive)
			hasNegInPrev := strings.Contains(prevLower, negative)

			if (hasPosInCurrent && hasNegInPrev) || (hasNegInCurrent && hasPosInPrev) {
				contradictionCount++
			}
			totalComparisons++
		}
	}

	// Calculate overlap to detect topic drift
	overlapSum := 0.0
	for _, prev := range previousResponses {
		prevWords := tokenize(strings.ToLower(prev))
		overlap := wordOverlap(currentWords, prevWords)
		overlapSum += overlap
	}
	avgOverlap := overlapSum / float64(len(previousResponses))

	// Low overlap = topic drift = higher risk
	driftRisk := 0.0
	if avgOverlap < 0.1 {
		driftRisk = 0.5
	} else if avgOverlap < 0.3 {
		driftRisk = 0.3
	}

	// Contradiction risk
	contradictionRisk := 0.0
	if contradictionCount > 3 {
		contradictionRisk = 0.7
	} else if contradictionCount > 1 {
		contradictionRisk = 0.4
	}

	return (driftRisk + contradictionRisk) / 2.0
}

/* tokenize splits text into words */
func tokenize(s string) []string {
	words := []string{}
	current := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			current += string(r)
		} else if current != "" {
			words = append(words, current)
			current = ""
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

/* wordOverlap calculates overlap ratio */
func wordOverlap(words1, words2 []string) float64 {
	if len(words1) == 0 {
		return 0.0
	}
	set2 := make(map[string]bool)
	for _, w := range words2 {
		set2[w] = true
	}
	overlap := 0
	for _, w := range words1 {
		if set2[w] {
			overlap++
		}
	}
	return float64(overlap) / float64(len(words1))
}

/* assessSpecificityRisk assesses risk based on response specificity */
func (hd *HallucinationDetector) assessSpecificityRisk(responseText string) float64 {
	// Check for vague language
	vagueWords := []string{"maybe", "perhaps", "possibly", "might", "could", "some", "various", "several"}

	responseLower := strings.ToLower(responseText)
	vagueCount := 0

	for _, word := range vagueWords {
		if strings.Contains(responseLower, word) {
			vagueCount++
		}
	}

	// More vague words = higher risk
	if vagueCount > 3 {
		return 0.6
	} else if vagueCount > 1 {
		return 0.4
	}

	return 0.2
}

/* storeRisk stores hallucination risk assessment */
func (hd *HallucinationDetector) storeRisk(ctx context.Context, risk *HallucinationRisk) error {
	factorsJSON, _ := json.Marshal(risk.RiskFactors)
	metadataJSON, _ := json.Marshal(risk.Metadata)

	query := `
		INSERT INTO neuronip.agent_hallucination_risks 
		(id, trace_id, response_id, risk_score, risk_factors, confidence, recommendation, metadata, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := hd.pool.Exec(ctx, query,
		risk.ID, risk.TraceID, risk.ResponseID, risk.RiskScore, factorsJSON,
		risk.Confidence, risk.Recommendation, metadataJSON, risk.EvaluatedAt,
	)
	return err
}

/* GetRisk retrieves hallucination risk for a trace */
func (hd *HallucinationDetector) GetRisk(ctx context.Context, traceID uuid.UUID) (*HallucinationRisk, error) {
	query := `
		SELECT id, trace_id, response_id, risk_score, risk_factors, confidence, recommendation, metadata, evaluated_at
		FROM neuronip.agent_hallucination_risks
		WHERE trace_id = $1
		ORDER BY evaluated_at DESC
		LIMIT 1
	`

	var risk HallucinationRisk
	var responseID *uuid.UUID
	var factorsJSON, metadataJSON json.RawMessage

	err := hd.pool.QueryRow(ctx, query, traceID).Scan(
		&risk.ID, &risk.TraceID, &responseID, &risk.RiskScore, &factorsJSON,
		&risk.Confidence, &risk.Recommendation, &metadataJSON, &risk.EvaluatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk: %w", err)
	}

	risk.ResponseID = responseID
	if factorsJSON != nil {
		json.Unmarshal(factorsJSON, &risk.RiskFactors)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &risk.Metadata)
	}

	return &risk, nil
}

// Helper functions removed - using strings package (ToLower, Contains) instead
