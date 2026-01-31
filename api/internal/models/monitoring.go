package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ModelMonitoringService provides model monitoring functionality */
type ModelMonitoringService struct {
	pool *pgxpool.Pool
}

/* NewModelMonitoringService creates a new model monitoring service */
func NewModelMonitoringService(pool *pgxpool.Pool) *ModelMonitoringService {
	return &ModelMonitoringService{pool: pool}
}

/* RecordPrediction records a model prediction for monitoring */
func (mms *ModelMonitoringService) RecordPrediction(ctx context.Context, prediction ModelPrediction) error {
	prediction.ID = uuid.New()
	prediction.Timestamp = time.Now()

	featuresJSON, _ := json.Marshal(prediction.Features)
	metadataJSON, _ := json.Marshal(prediction.Metadata)

	query := `
		INSERT INTO neuronip.model_predictions 
		(id, model_id, features, prediction, actual_value, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := mms.pool.Exec(ctx, query,
		prediction.ID, prediction.ModelID, featuresJSON, prediction.Prediction,
		prediction.ActualValue, metadataJSON, prediction.Timestamp,
	)
	return err
}

/* DetectDataDrift detects data drift */
func (mms *ModelMonitoringService) DetectDataDrift(ctx context.Context, modelID uuid.UUID) (*DriftDetection, error) {
	// Get baseline distribution (from training data)
	baselineQuery := `
		SELECT AVG(features->>'feature1')::float, STDDEV(features->>'feature1')::float
		FROM neuronip.model_training_data
		WHERE model_id = $1`

	var baselineMean, baselineStddev float64
	err := mms.pool.QueryRow(ctx, baselineQuery, modelID).Scan(&baselineMean, &baselineStddev)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline: %w", err)
	}

	// Get current distribution (from recent predictions)
	currentQuery := `
		SELECT AVG(features->>'feature1')::float, STDDEV(features->>'feature1')::float
		FROM neuronip.model_predictions
		WHERE model_id = $1 AND timestamp > NOW() - INTERVAL '24 hours'`

	var currentMean, currentStddev float64
	err = mms.pool.QueryRow(ctx, currentQuery, modelID).Scan(&currentMean, &currentStddev)
	if err != nil {
		return nil, fmt.Errorf("failed to get current distribution: %w", err)
	}

	// Calculate drift score
	driftScore := calculateDriftScore(baselineMean, baselineStddev, currentMean, currentStddev)

	detection := &DriftDetection{
		ID:           uuid.New(),
		ModelID:      modelID,
		DriftScore:   driftScore,
		DetectedAt:   time.Now(),
		BaselineMean: baselineMean,
		CurrentMean:  currentMean,
	}

	// Store detection
	query := `
		INSERT INTO neuronip.model_drift_detections 
		(id, model_id, drift_score, baseline_mean, current_mean, detected_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = mms.pool.Exec(ctx, query,
		detection.ID, detection.ModelID, detection.DriftScore,
		detection.BaselineMean, detection.CurrentMean, detection.DetectedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to store drift detection: %w", err)
	}

	return detection, nil
}

/* calculateDriftScore calculates drift score */
func calculateDriftScore(baselineMean, baselineStddev, currentMean, currentStddev float64) float64 {
	// Simple drift calculation using statistical distance
	if baselineStddev == 0 {
		return abs(currentMean - baselineMean)
	}
	return abs((currentMean - baselineMean) / baselineStddev)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

/* ModelPrediction represents a model prediction */
type ModelPrediction struct {
	ID          uuid.UUID              `json:"id"`
	ModelID     uuid.UUID              `json:"model_id"`
	Features    map[string]interface{} `json:"features"`
	Prediction  interface{}            `json:"prediction"`
	ActualValue *interface{}           `json:"actual_value,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

/* DriftDetection represents a data drift detection */
type DriftDetection struct {
	ID           uuid.UUID `json:"id"`
	ModelID      uuid.UUID `json:"model_id"`
	DriftScore   float64   `json:"drift_score"`
	BaselineMean float64   `json:"baseline_mean"`
	CurrentMean  float64   `json:"current_mean"`
	DetectedAt   time.Time `json:"detected_at"`
}
