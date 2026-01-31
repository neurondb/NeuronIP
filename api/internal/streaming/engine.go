package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* StreamingEngine provides stream processing functionality */
type StreamingEngine struct {
	pool   *pgxpool.Pool
	runner *Runner
}

/* NewStreamingEngine creates a new streaming engine with in-process pipeline runner */
func NewStreamingEngine(pool *pgxpool.Pool) *StreamingEngine {
	return &StreamingEngine{pool: pool, runner: NewRunner(pool)}
}

/* CreatePipeline creates a new streaming pipeline */
func (se *StreamingEngine) CreatePipeline(ctx context.Context, pipeline Pipeline) (*Pipeline, error) {
	pipeline.ID = uuid.New()
	pipeline.CreatedAt = time.Now()
	pipeline.UpdatedAt = time.Now()
	pipeline.Status = "created"

	sourcesJSON, _ := json.Marshal(pipeline.Sources)
	transformationsJSON, _ := json.Marshal(pipeline.Transformations)
	sinksJSON, _ := json.Marshal(pipeline.Sinks)
	configJSON, _ := json.Marshal(pipeline.Config)

	query := `
		INSERT INTO neuronip.streaming_pipelines 
		(id, pipeline_name, description, sources, transformations, sinks, config, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, pipeline_name, description, sources, transformations, sinks, config, status, created_at, updated_at`

	var sourcesJSONRaw, transformationsJSONRaw, sinksJSONRaw, configJSONRaw json.RawMessage
	err := se.pool.QueryRow(ctx, query,
		pipeline.ID, pipeline.Name, pipeline.Description, sourcesJSON, transformationsJSON,
		sinksJSON, configJSON, pipeline.Status, pipeline.CreatedAt, pipeline.UpdatedAt,
	).Scan(
		&pipeline.ID, &pipeline.Name, &pipeline.Description, &sourcesJSONRaw, &transformationsJSONRaw,
		&sinksJSONRaw, &configJSONRaw, &pipeline.Status, &pipeline.CreatedAt, &pipeline.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	if sourcesJSONRaw != nil {
		json.Unmarshal(sourcesJSONRaw, &pipeline.Sources)
	}
	if transformationsJSONRaw != nil {
		json.Unmarshal(transformationsJSONRaw, &pipeline.Transformations)
	}
	if sinksJSONRaw != nil {
		json.Unmarshal(sinksJSONRaw, &pipeline.Sinks)
	}
	if configJSONRaw != nil {
		json.Unmarshal(configJSONRaw, &pipeline.Config)
	}

	return &pipeline, nil
}

/* StartPipeline starts a streaming pipeline (runs worker that polls postgres_events and writes to sink) */
func (se *StreamingEngine) StartPipeline(ctx context.Context, pipelineID uuid.UUID) error {
	_, err := se.pool.Exec(ctx, `UPDATE neuronip.streaming_pipelines SET status = 'running', updated_at = NOW() WHERE id = $1`, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to start pipeline: %w", err)
	}
	if se.runner != nil {
		return se.runner.Start(ctx, se, pipelineID)
	}
	return nil
}

/* StopPipeline stops a streaming pipeline and its worker */
func (se *StreamingEngine) StopPipeline(ctx context.Context, pipelineID uuid.UUID) error {
	if se.runner != nil {
		se.runner.Stop(pipelineID)
	}
	_, err := se.pool.Exec(ctx, `UPDATE neuronip.streaming_pipelines SET status = 'stopped', updated_at = NOW() WHERE id = $1`, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to stop pipeline: %w", err)
	}
	return nil
}

/* Pipeline represents a streaming pipeline */
type Pipeline struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Sources         []Source               `json:"sources"`
	Transformations []Transformation       `json:"transformations"`
	Sinks           []Sink                 `json:"sinks"`
	Config          map[string]interface{} `json:"config"`
	Status          string                 `json:"status"` // "created", "running", "stopped", "failed"
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

/* Source represents a data source */
type Source struct {
	Type          string                 `json:"type"` // "kafka", "kinesis", "pulsar", "webhook"
	Config        map[string]interface{} `json:"config"`
	Topic         string                 `json:"topic,omitempty"`
	ConsumerGroup string                 `json:"consumer_group,omitempty"`
}

/* Transformation represents a stream transformation */
type Transformation struct {
	Type       string                 `json:"type"` // "filter", "map", "aggregate", "window", "join"
	Config     map[string]interface{} `json:"config"`
	WindowType string                 `json:"window_type,omitempty"` // "tumbling", "sliding", "session"
	WindowSize string                 `json:"window_size,omitempty"` // e.g., "5m", "1h"
}

/* Sink represents a data sink */
type Sink struct {
	Type   string                 `json:"type"` // "kafka", "database", "warehouse", "api"
	Config map[string]interface{} `json:"config"`
	Target string                 `json:"target"`
}
