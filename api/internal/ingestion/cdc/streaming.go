package cdc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* StreamingPipelineEngine manages streaming data pipelines */
type StreamingPipelineEngine struct {
	pool *pgxpool.Pool
}

/* NewStreamingPipelineEngine creates a new streaming pipeline engine */
func NewStreamingPipelineEngine(pool *pgxpool.Pool) *StreamingPipelineEngine {
	return &StreamingPipelineEngine{pool: pool}
}

/* CreatePipeline creates a new streaming pipeline */
func (e *StreamingPipelineEngine) CreatePipeline(ctx context.Context, pipeline StreamingPipeline) (*StreamingPipeline, error) {
	pipeline.ID = uuid.New()
	sourceJSON, _ := json.Marshal(pipeline.SourceConfig)
	destJSON, _ := json.Marshal(pipeline.DestinationConfig)
	transformJSON, _ := json.Marshal(pipeline.TransformationConfig)

	query := `
		INSERT INTO neuronip.streaming_pipelines 
		(id, pipeline_name, source_type, source_config, destination_type, destination_config, 
		 transformation_config, status, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, pipeline_name, source_type, source_config, destination_type, destination_config,
		          transformation_config, status, enabled, created_at, updated_at`

	var sourceJSONRaw, destJSONRaw, transformJSONRaw json.RawMessage
	err := e.pool.QueryRow(ctx, query,
		pipeline.ID, pipeline.PipelineName, pipeline.SourceType, sourceJSON,
		pipeline.DestinationType, destJSON, transformJSON, pipeline.Status, pipeline.Enabled,
	).Scan(
		&pipeline.ID, &pipeline.PipelineName, &pipeline.SourceType, &sourceJSONRaw,
		&pipeline.DestinationType, &destJSONRaw, &transformJSONRaw, &pipeline.Status,
		&pipeline.Enabled, &pipeline.CreatedAt, &pipeline.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	if sourceJSONRaw != nil {
		json.Unmarshal(sourceJSONRaw, &pipeline.SourceConfig)
	}
	if destJSONRaw != nil {
		json.Unmarshal(destJSONRaw, &pipeline.DestinationConfig)
	}
	if transformJSONRaw != nil {
		json.Unmarshal(transformJSONRaw, &pipeline.TransformationConfig)
	}

	return &pipeline, nil
}

/* StartPipeline starts a streaming pipeline */
func (e *StreamingPipelineEngine) StartPipeline(ctx context.Context, pipelineID uuid.UUID) error {
	query := `
		UPDATE neuronip.streaming_pipelines 
		SET status = 'active', enabled = true, updated_at = NOW()
		WHERE id = $1`
	
	result, err := e.pool.Exec(ctx, query, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to start pipeline: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pipeline not found")
	}

	return nil
}

/* StopPipeline stops a streaming pipeline */
func (e *StreamingPipelineEngine) StopPipeline(ctx context.Context, pipelineID uuid.UUID) error {
	query := `
		UPDATE neuronip.streaming_pipelines 
		SET status = 'inactive', enabled = false, updated_at = NOW()
		WHERE id = $1`
	
	result, err := e.pool.Exec(ctx, query, pipelineID)
	if err != nil {
		return fmt.Errorf("failed to stop pipeline: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pipeline not found")
	}

	return nil
}

/* SaveCheckpoint saves a stream processing checkpoint */
func (e *StreamingPipelineEngine) SaveCheckpoint(ctx context.Context, pipelineID uuid.UUID, partitionID string, offset int64, watermark *time.Time) error {
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"updated_at": time.Now(),
	})

	query := `
		INSERT INTO neuronip.stream_checkpoints 
		(id, pipeline_id, partition_id, offset, watermark, metadata, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())
		ON CONFLICT (pipeline_id, partition_id) DO UPDATE SET
			offset = EXCLUDED.offset,
			watermark = EXCLUDED.watermark,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()`

	_, err := e.pool.Exec(ctx, query, pipelineID, partitionID, offset, watermark, metadataJSON)
	return err
}

/* GetCheckpoint retrieves a stream processing checkpoint */
func (e *StreamingPipelineEngine) GetCheckpoint(ctx context.Context, pipelineID uuid.UUID, partitionID string) (*StreamCheckpoint, error) {
	query := `
		SELECT id, pipeline_id, partition_id, offset, watermark, metadata, updated_at
		FROM neuronip.stream_checkpoints
		WHERE pipeline_id = $1 AND partition_id = $2`

	var checkpoint StreamCheckpoint
	var metadataJSON json.RawMessage
	var watermark *time.Time

	err := e.pool.QueryRow(ctx, query, pipelineID, partitionID).Scan(
		&checkpoint.ID, &checkpoint.PipelineID, &checkpoint.PartitionID,
		&checkpoint.Offset, &watermark, &metadataJSON, &checkpoint.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint: %w", err)
	}

	checkpoint.Watermark = watermark
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &checkpoint.Metadata)
	}

	return &checkpoint, nil
}

/* RecordStreamMetric records a streaming pipeline metric */
func (e *StreamingPipelineEngine) RecordStreamMetric(ctx context.Context, pipelineID uuid.UUID, metricName string, metricValue float64, unit *string, partitionID *string) error {
	metricID := uuid.New()
	query := `
		INSERT INTO neuronip.stream_metrics 
		(id, pipeline_id, metric_name, metric_value, metric_unit, partition_id, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	_, err := e.pool.Exec(ctx, query, metricID, pipelineID, metricName, metricValue, unit, partitionID)
	return err
}

/* StreamingPipeline represents a streaming data pipeline */
type StreamingPipeline struct {
	ID                  uuid.UUID              `json:"id"`
	PipelineName        string                 `json:"pipeline_name"`
	SourceType          string                 `json:"source_type"`
	SourceConfig         map[string]interface{} `json:"source_config"`
	DestinationType     string                 `json:"destination_type"`
	DestinationConfig    map[string]interface{} `json:"destination_config"`
	TransformationConfig map[string]interface{} `json:"transformation_config,omitempty"`
	Status              string                 `json:"status"`
	Enabled             bool                   `json:"enabled"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

/* StreamCheckpoint represents a stream processing checkpoint */
type StreamCheckpoint struct {
	ID          uuid.UUID              `json:"id"`
	PipelineID  uuid.UUID              `json:"pipeline_id"`
	PartitionID string                 `json:"partition_id"`
	Offset      int64                  `json:"offset"`
	Watermark   *time.Time             `json:"watermark,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
