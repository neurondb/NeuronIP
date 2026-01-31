package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ExperimentService provides experiment tracking functionality */
type ExperimentService struct {
	pool *pgxpool.Pool
}

/* NewExperimentService creates a new experiment service */
func NewExperimentService(pool *pgxpool.Pool) *ExperimentService {
	return &ExperimentService{pool: pool}
}

/* CreateExperiment creates a new experiment */
func (es *ExperimentService) CreateExperiment(ctx context.Context, experiment Experiment) (*Experiment, error) {
	experiment.ID = uuid.New()
	experiment.CreatedAt = time.Now()
	experiment.UpdatedAt = time.Now()

	paramsJSON, _ := json.Marshal(experiment.Parameters)
	metricsJSON, _ := json.Marshal(experiment.Metrics)
	tagsJSON, _ := json.Marshal(experiment.Tags)

	query := `
		INSERT INTO neuronip.ml_experiments 
		(id, experiment_name, description, parameters, metrics, tags, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, experiment_name, description, parameters, metrics, tags, status, created_at, updated_at`

	var paramsJSONRaw, metricsJSONRaw, tagsJSONRaw json.RawMessage
	err := es.pool.QueryRow(ctx, query,
		experiment.ID, experiment.Name, experiment.Description, paramsJSON, metricsJSON,
		tagsJSON, experiment.Status, experiment.CreatedAt, experiment.UpdatedAt,
	).Scan(
		&experiment.ID, &experiment.Name, &experiment.Description, &paramsJSONRaw,
		&metricsJSONRaw, &tagsJSONRaw, &experiment.Status, &experiment.CreatedAt, &experiment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create experiment: %w", err)
	}

	if paramsJSONRaw != nil {
		json.Unmarshal(paramsJSONRaw, &experiment.Parameters)
	}
	if metricsJSONRaw != nil {
		json.Unmarshal(metricsJSONRaw, &experiment.Metrics)
	}
	if tagsJSONRaw != nil {
		json.Unmarshal(tagsJSONRaw, &experiment.Tags)
	}

	return &experiment, nil
}

/* LogMetric logs a metric for an experiment */
func (es *ExperimentService) LogMetric(ctx context.Context, experimentID uuid.UUID, metricName string, metricValue float64, step int) error {
	query := `
		INSERT INTO neuronip.ml_experiment_metrics 
		(experiment_id, metric_name, metric_value, step, timestamp)
		VALUES ($1, $2, $3, $4, NOW())`

	_, err := es.pool.Exec(ctx, query, experimentID, metricName, metricValue, step)
	return err
}

/* CompareExperiments compares multiple experiments */
func (es *ExperimentService) CompareExperiments(ctx context.Context, experimentIDs []uuid.UUID) ([]ExperimentComparison, error) {
	// Query experiments and compare metrics
	comparisons := make([]ExperimentComparison, 0, len(experimentIDs))

	for _, expID := range experimentIDs {
		exp, err := es.GetExperiment(ctx, expID)
		if err != nil {
			continue
		}

		comparison := ExperimentComparison{
			ExperimentID: expID,
			ExperimentName: exp.Name,
			Metrics: exp.Metrics,
			Parameters: exp.Parameters,
		}
		comparisons = append(comparisons, comparison)
	}

	return comparisons, nil
}

/* GetExperiment retrieves an experiment */
func (es *ExperimentService) GetExperiment(ctx context.Context, experimentID uuid.UUID) (*Experiment, error) {
	query := `
		SELECT id, experiment_name, description, parameters, metrics, tags, status, created_at, updated_at
		FROM neuronip.ml_experiments
		WHERE id = $1`

	var experiment Experiment
	var paramsJSONRaw, metricsJSONRaw, tagsJSONRaw json.RawMessage

	err := es.pool.QueryRow(ctx, query, experimentID).Scan(
		&experiment.ID, &experiment.Name, &experiment.Description, &paramsJSONRaw,
		&metricsJSONRaw, &tagsJSONRaw, &experiment.Status, &experiment.CreatedAt, &experiment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}

	if paramsJSONRaw != nil {
		json.Unmarshal(paramsJSONRaw, &experiment.Parameters)
	}
	if metricsJSONRaw != nil {
		json.Unmarshal(metricsJSONRaw, &experiment.Metrics)
	}
	if tagsJSONRaw != nil {
		json.Unmarshal(tagsJSONRaw, &experiment.Tags)
	}

	return &experiment, nil
}

/* Experiment represents an ML experiment */
type Experiment struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Metrics     map[string]interface{} `json:"metrics"`
	Tags        map[string]string      `json:"tags"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* ExperimentComparison represents experiment comparison */
type ExperimentComparison struct {
	ExperimentID   uuid.UUID              `json:"experiment_id"`
	ExperimentName string                 `json:"experiment_name"`
	Metrics        map[string]interface{} `json:"metrics"`
	Parameters     map[string]interface{} `json:"parameters"`
}
