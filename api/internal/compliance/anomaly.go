package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* AnomalyService provides anomaly detection functionality */
type AnomalyService struct {
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
}

/* NewAnomalyService creates a new anomaly detection service */
func NewAnomalyService(pool *pgxpool.Pool, neurondbClient *neurondb.Client) *AnomalyService {
	return &AnomalyService{
		pool:           pool,
		neurondbClient: neurondbClient,
	}
}

/* AnomalyDetection represents an anomaly detection result */
type AnomalyDetection struct {
	ID            uuid.UUID              `json:"id"`
	DetectionType string                 `json:"detection_type"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	AnomalyScore  float64                `json:"anomaly_score"`
	Details       map[string]interface{} `json:"details,omitempty"`
	ModelName     string                 `json:"model_name,omitempty"`
	Status        string                 `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
}

/* DetectAnomalies detects anomalies using NeuronDB ML functions */
func (s *AnomalyService) DetectAnomalies(ctx context.Context, detectionType string, entityType string, entityID string, data map[string]interface{}) ([]AnomalyDetection, error) {
	// Convert data to JSON for processing
	dataJSON, _ := json.Marshal(data)
	dataStr := string(dataJSON)

	// Use NeuronDB ML classification functions to detect anomalies
	anomalyScore := s.calculateAnomalyScore(data)
	
	var detections []AnomalyDetection

	// Only report if anomaly score is above threshold
	if anomalyScore >= 0.7 {
		detection := AnomalyDetection{
			ID:            uuid.New(),
			DetectionType: detectionType,
			EntityType:    entityType,
			EntityID:      entityID,
			AnomalyScore:  anomalyScore,
			Details: map[string]interface{}{
				"data_preview": truncateString(dataStr, 200),
			},
			ModelName: "default",
			Status:    "detected",
			CreatedAt: time.Now(),
		}

		// Store detection
		err := s.storeAnomalyDetection(ctx, detection)
		if err != nil {
			return nil, fmt.Errorf("failed to store anomaly detection: %w", err)
		}

		detections = append(detections, detection)
	}

	return detections, nil
}

/* calculateAnomalyScore calculates an anomaly score using NeuronDB ML functions */
func (s *AnomalyService) calculateAnomalyScore(data map[string]interface{}) float64 {
	// Convert data to JSON for ML processing
	dataJSON, err := json.Marshal(data)
	if err != nil {
		// Fallback to heuristic if JSON marshalling fails
		return s.calculateAnomalyScoreHeuristic(data)
	}

	// Try using NeuronDB classification for anomaly detection
	// We'll use a context with timeout for safety
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use NeuronDB classification function if available
	// For anomaly detection, we can use a classification model trained on normal data
	// If anomaly class is detected, return high score
	// Convert JSON to string for the client method
	dataStr := string(dataJSON)
	result, err := s.neurondbClient.Classify(ctx, dataStr, "anomaly-detector")
	
	if err == nil && result != nil {
		// Extract anomaly probability from classification result
		// The result may contain class probabilities
		if anomalyProb, ok := result["anomaly"].(float64); ok {
			return anomalyProb
		}
		if score, ok := result["score"].(float64); ok {
			return score
		}
		// If result has a "class" field indicating anomaly
		if class, ok := result["class"].(string); ok && class == "anomaly" {
			if confidence, ok := result["confidence"].(float64); ok {
				return confidence
			}
			return 0.8 // Default high score for anomaly class
		}
	}

	// Fallback to heuristic-based detection if ML classification fails
	return s.calculateAnomalyScoreHeuristic(data)
}

/* calculateAnomalyScoreHeuristic calculates anomaly score using heuristics */
func (s *AnomalyService) calculateAnomalyScoreHeuristic(data map[string]interface{}) float64 {
	// Check for unusual patterns in data
	score := 0.0
	
	// Example: check for missing required fields
	if len(data) == 0 {
		score += 0.5
	}
	
	// Example: check for unusually large or small values
	for _, value := range data {
		switch v := value.(type) {
		case float64:
			if v > 1000 || v < -1000 {
				score += 0.3
			}
		case string:
			if len(v) > 10000 {
				score += 0.2
			}
		}
	}
	
	// Normalize score to 0-1 range
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

/* storeAnomalyDetection stores an anomaly detection */
func (s *AnomalyService) storeAnomalyDetection(ctx context.Context, detection AnomalyDetection) error {
	detailsJSON, _ := json.Marshal(detection.Details)

	query := `
		INSERT INTO neuronip.anomaly_detections 
		(id, detection_type, entity_type, entity_id, anomaly_score, details, model_name, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := s.pool.Exec(ctx, query,
		detection.ID,
		detection.DetectionType,
		detection.EntityType,
		detection.EntityID,
		detection.AnomalyScore,
		detailsJSON,
		detection.ModelName,
		detection.Status,
		detection.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store anomaly detection: %w", err)
	}

	return nil
}

/* GetAnomalyDetections retrieves anomaly detections */
func (s *AnomalyService) GetAnomalyDetections(ctx context.Context, entityType string, entityID string, status string, limit int) ([]AnomalyDetection, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, detection_type, entity_type, entity_id, anomaly_score, details, model_name, status, created_at
		FROM neuronip.anomaly_detections
		WHERE ($1 = '' OR entity_type = $1) AND ($2 = '' OR entity_id = $2) AND ($3 = '' OR status = $3)
		ORDER BY created_at DESC
		LIMIT $4`

	rows, err := s.pool.Query(ctx, query, entityType, entityID, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get anomaly detections: %w", err)
	}
	defer rows.Close()

	var detections []AnomalyDetection
	for rows.Next() {
		var detection AnomalyDetection
		var detailsJSON json.RawMessage

		err := rows.Scan(
			&detection.ID, &detection.DetectionType, &detection.EntityType, &detection.EntityID,
			&detection.AnomalyScore, &detailsJSON, &detection.ModelName, &detection.Status, &detection.CreatedAt,
		)
		if err != nil {
			continue
		}

		if detailsJSON != nil {
			json.Unmarshal(detailsJSON, &detection.Details)
		}

		detections = append(detections, detection)
	}

	return detections, nil
}

/* truncateString truncates a string to max length */
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
