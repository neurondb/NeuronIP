package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* QualityScorer provides model output quality scoring functionality */
type QualityScorer struct {
	pool *pgxpool.Pool
}

/* NewQualityScorer creates a new quality scorer */
func NewQualityScorer(pool *pgxpool.Pool) *QualityScorer {
	return &QualityScorer{pool: pool}
}

/* QualityScore represents a quality score for model output */
type QualityScore struct {
	ID              uuid.UUID              `json:"id"`
	ModelID         uuid.UUID              `json:"model_id"`
	ModelVersion    string                 `json:"model_version"`
	OutputID        *uuid.UUID             `json:"output_id,omitempty"`
	Score           float64                `json:"score"` // 0.0 to 1.0
	ScoreComponents map[string]float64     `json:"score_components,omitempty"`
	Metrics         map[string]interface{} `json:"metrics,omitempty"`
	EvaluatedAt     time.Time              `json:"evaluated_at"`
	EvaluatedBy     *string                `json:"evaluated_by,omitempty"`
}

/* ScoreOutput scores the quality of a model output */
func (qs *QualityScorer) ScoreOutput(ctx context.Context, modelID uuid.UUID, modelVersion string, output interface{}, expectedOutput interface{}, metadata map[string]interface{}) (*QualityScore, error) {
	score := &QualityScore{
		ID:              uuid.New(),
		ModelID:         modelID,
		ModelVersion:    modelVersion,
		ScoreComponents: make(map[string]float64),
		Metrics:         make(map[string]interface{}),
		EvaluatedAt:     time.Now(),
	}

	// Calculate different quality components
	accuracy := qs.calculateAccuracy(output, expectedOutput)
	consistency := qs.calculateConsistency(output, metadata)
	relevance := qs.calculateRelevance(output, metadata)
	completeness := qs.calculateCompleteness(output, metadata)

	score.ScoreComponents["accuracy"] = accuracy
	score.ScoreComponents["consistency"] = consistency
	score.ScoreComponents["relevance"] = relevance
	score.ScoreComponents["completeness"] = completeness

	// Weighted average
	score.Score = (accuracy*0.4 + consistency*0.2 + relevance*0.2 + completeness*0.2)

	// Store score
	if err := qs.storeScore(ctx, score); err != nil {
		return nil, fmt.Errorf("failed to store score: %w", err)
	}

	return score, nil
}

/* calculateAccuracy calculates accuracy score */
func (qs *QualityScorer) calculateAccuracy(output, expected interface{}) float64 {
	// Simple string comparison for now
	// In production, use more sophisticated comparison
	outputStr := fmt.Sprintf("%v", output)
	expectedStr := fmt.Sprintf("%v", expected)

	if outputStr == expectedStr {
		return 1.0
	}

	// Calculate similarity (simple)
	similarity := qs.stringSimilarity(outputStr, expectedStr)
	return similarity
}

/* calculateConsistency calculates consistency score by comparing with historical outputs */
func (qs *QualityScorer) calculateConsistency(output interface{}, metadata map[string]interface{}) float64 {
	outputStr := fmt.Sprintf("%v", output)
	outputWords := qs.tokenize(outputStr)

	// Get model ID from metadata for historical lookup
	modelIDStr, _ := metadata["model_id"].(string)
	if modelIDStr == "" {
		return 0.8 // Default if no model context
	}

	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		return 0.8
	}

	// Query last 10 outputs for this model
	ctx := context.Background()
	rows, err := qs.pool.Query(ctx, `
		SELECT metrics->>'output' as prev_output
		FROM neuronip.model_quality_scores
		WHERE model_id = $1
		ORDER BY evaluated_at DESC
		LIMIT 10
	`, modelID)
	if err != nil {
		return 0.8
	}
	defer rows.Close()

	var consistencySum float64
	var count int
	for rows.Next() {
		var prevOutput *string
		if err := rows.Scan(&prevOutput); err != nil || prevOutput == nil {
			continue
		}
		prevWords := qs.tokenize(*prevOutput)
		overlap := qs.wordOverlap(outputWords, prevWords)
		consistencySum += overlap
		count++
	}

	if count == 0 {
		return 0.9 // First output, assume consistent
	}
	return consistencySum / float64(count)
}

/* calculateRelevance calculates relevance score based on input/context similarity */
func (qs *QualityScorer) calculateRelevance(output interface{}, metadata map[string]interface{}) float64 {
	outputStr := fmt.Sprintf("%v", output)
	outputWords := qs.tokenize(outputStr)

	// Check for input/prompt in metadata
	inputStr, _ := metadata["input"].(string)
	if inputStr == "" {
		inputStr, _ = metadata["prompt"].(string)
	}
	if inputStr == "" {
		inputStr, _ = metadata["query"].(string)
	}
	if inputStr == "" {
		return 0.85 // Default if no input context
	}

	inputWords := qs.tokenize(inputStr)
	if len(inputWords) == 0 {
		return 0.85
	}

	// Check keyword overlap between input and output
	overlap := qs.wordOverlap(inputWords, outputWords)

	// Score based on overlap ratio (input keywords present in output)
	return 0.5 + 0.5*overlap
}

/* wordOverlap calculates the overlap ratio of words1 found in words2 */
func (qs *QualityScorer) wordOverlap(words1, words2 []string) float64 {
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

/* calculateCompleteness calculates completeness score */
func (qs *QualityScorer) calculateCompleteness(output interface{}, metadata map[string]interface{}) float64 {
	// Check if output is complete
	outputStr := fmt.Sprintf("%v", output)
	if len(outputStr) < 10 {
		return 0.5
	}
	return 0.9
}

/* stringSimilarity calculates similarity between two strings */
func (qs *QualityScorer) stringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	// Simple word overlap
	words1 := qs.tokenize(s1)
	words2 := qs.tokenize(s2)

	overlap := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 {
				overlap++
				break
			}
		}
	}

	if len(words1) == 0 {
		return 0.0
	}

	return float64(overlap) / float64(len(words1))
}

/* tokenize tokenizes a string into words */
func (qs *QualityScorer) tokenize(s string) []string {
	// Simple tokenization
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

/* storeScore stores a quality score */
func (qs *QualityScorer) storeScore(ctx context.Context, score *QualityScore) error {
	componentsJSON, _ := json.Marshal(score.ScoreComponents)
	metricsJSON, _ := json.Marshal(score.Metrics)

	query := `
		INSERT INTO neuronip.model_quality_scores 
		(id, model_id, model_version, output_id, score, score_components, metrics, evaluated_at, evaluated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := qs.pool.Exec(ctx, query,
		score.ID, score.ModelID, score.ModelVersion, score.OutputID, score.Score,
		componentsJSON, metricsJSON, score.EvaluatedAt, score.EvaluatedBy,
	)
	return err
}

/* GetQualityScores retrieves quality scores for a model */
func (qs *QualityScorer) GetQualityScores(ctx context.Context, modelID uuid.UUID, limit int) ([]QualityScore, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, model_id, model_version, output_id, score, score_components, metrics, evaluated_at, evaluated_by
		FROM neuronip.model_quality_scores
		WHERE model_id = $1
		ORDER BY evaluated_at DESC
		LIMIT $2
	`

	rows, err := qs.pool.Query(ctx, query, modelID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get quality scores: %w", err)
	}
	defer rows.Close()

	var scores []QualityScore
	for rows.Next() {
		var score QualityScore
		var outputID *uuid.UUID
		var componentsJSON, metricsJSON json.RawMessage
		var evaluatedBy *string

		err := rows.Scan(
			&score.ID, &score.ModelID, &score.ModelVersion, &outputID, &score.Score,
			&componentsJSON, &metricsJSON, &score.EvaluatedAt, &evaluatedBy,
		)
		if err != nil {
			continue
		}

		score.OutputID = outputID
		score.EvaluatedBy = evaluatedBy
		if componentsJSON != nil {
			json.Unmarshal(componentsJSON, &score.ScoreComponents)
		}
		if metricsJSON != nil {
			json.Unmarshal(metricsJSON, &score.Metrics)
		}

		scores = append(scores, score)
	}

	return scores, nil
}

/* GetAverageQualityScore calculates average quality score for a model */
func (qs *QualityScorer) GetAverageQualityScore(ctx context.Context, modelID uuid.UUID) (float64, error) {
	query := `
		SELECT AVG(score) as avg_score
		FROM neuronip.model_quality_scores
		WHERE model_id = $1
	`

	var avgScore *float64
	err := qs.pool.QueryRow(ctx, query, modelID).Scan(&avgScore)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get average score: %w", err)
	}

	if avgScore == nil {
		return 0.0, nil
	}

	return *avgScore, nil
}
