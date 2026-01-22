package semantic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* PipelineService provides versioned pipeline management */
type PipelineService struct {
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
}

/* NewPipelineService creates a new pipeline service */
func NewPipelineService(pool *pgxpool.Pool) *PipelineService {
	return &PipelineService{pool: pool}
}

/* NewPipelineServiceWithClient creates a new pipeline service with NeuronDB client */
func NewPipelineServiceWithClient(pool *pgxpool.Pool, neurondbClient *neurondb.Client) *PipelineService {
	return &PipelineService{
		pool:           pool,
		neurondbClient: neurondbClient,
	}
}

/* Pipeline represents a versioned chunking/embedding pipeline */
type Pipeline struct {
	ID                uuid.UUID              `json:"id"`
	Version           string                 `json:"version"`
	Name              string                 `json:"name"`
	Description       *string                `json:"description,omitempty"`
	ChunkingConfig    ChunkingConfig         `json:"chunking_config"`
	EmbeddingModel    string                 `json:"embedding_model"`
	EmbeddingConfig   map[string]interface{} `json:"embedding_config"`
	IsActive          bool                   `json:"is_active"`
	PerformanceMetrics map[string]interface{} `json:"performance_metrics,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	CreatedBy         *string                `json:"created_by,omitempty"`
}

/* CreatePipeline creates a new pipeline version */
func (s *PipelineService) CreatePipeline(ctx context.Context, name string, chunkingConfig ChunkingConfig, embeddingModel string, embeddingConfig map[string]interface{}) (*Pipeline, error) {
	pipelineID := uuid.New()
	version := fmt.Sprintf("v%d", time.Now().Unix())

	chunkingJSON, _ := json.Marshal(chunkingConfig)
	embeddingJSON, _ := json.Marshal(embeddingConfig)

	query := `
		INSERT INTO neuronip.pipelines (
			id, version, name, chunking_config, embedding_model,
			embedding_config, is_active, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, false, NOW())
		RETURNING created_at
	`
	var createdAt time.Time
	err := s.pool.QueryRow(ctx, query,
		pipelineID, version, name, chunkingJSON, embeddingModel, embeddingJSON,
	).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	return &Pipeline{
		ID:              pipelineID,
		Version:         version,
		Name:            name,
		ChunkingConfig:  chunkingConfig,
		EmbeddingModel:  embeddingModel,
		EmbeddingConfig: embeddingConfig,
		IsActive:        false,
		CreatedAt:       createdAt,
	}, nil
}

/* GetPipeline retrieves a pipeline by ID and version */
func (s *PipelineService) GetPipeline(ctx context.Context, pipelineID uuid.UUID, version string) (*Pipeline, error) {
	query := `
		SELECT id, version, name, description, chunking_config, embedding_model,
		       embedding_config, is_active, performance_metrics, created_at, created_by
		FROM neuronip.pipelines
		WHERE id = $1 AND version = $2
	`
	var pipeline Pipeline
	var chunkingJSON, embeddingJSON, perfJSON []byte
	var description *string
	var createdBy *string

	err := s.pool.QueryRow(ctx, query, pipelineID, version).Scan(
		&pipeline.ID, &pipeline.Version, &pipeline.Name, &description,
		&chunkingJSON, &pipeline.EmbeddingModel, &embeddingJSON,
		&pipeline.IsActive, &perfJSON, &pipeline.CreatedAt, &createdBy,
	)
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	json.Unmarshal(chunkingJSON, &pipeline.ChunkingConfig)
	json.Unmarshal(embeddingJSON, &pipeline.EmbeddingConfig)
	if perfJSON != nil {
		json.Unmarshal(perfJSON, &pipeline.PerformanceMetrics)
	}
	pipeline.Description = description
	pipeline.CreatedBy = createdBy

	return &pipeline, nil
}

/* ListPipelineVersions lists all versions of a pipeline */
func (s *PipelineService) ListPipelineVersions(ctx context.Context, pipelineID uuid.UUID) ([]Pipeline, error) {
	query := `
		SELECT id, version, name, description, chunking_config, embedding_model,
		       embedding_config, is_active, performance_metrics, created_at, created_by
		FROM neuronip.pipelines
		WHERE id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.pool.Query(ctx, query, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pipeline versions: %w", err)
	}
	defer rows.Close()

	var pipelines []Pipeline
	for rows.Next() {
		var pipeline Pipeline
		var chunkingJSON, embeddingJSON, perfJSON []byte
		var description, createdBy *string

		err := rows.Scan(
			&pipeline.ID, &pipeline.Version, &pipeline.Name, &description,
			&chunkingJSON, &pipeline.EmbeddingModel, &embeddingJSON,
			&pipeline.IsActive, &perfJSON, &pipeline.CreatedAt, &createdBy,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(chunkingJSON, &pipeline.ChunkingConfig)
		json.Unmarshal(embeddingJSON, &pipeline.EmbeddingConfig)
		if perfJSON != nil {
			json.Unmarshal(perfJSON, &pipeline.PerformanceMetrics)
		}
		pipeline.Description = description
		pipeline.CreatedBy = createdBy

		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

/* ReplayPipeline reprocesses documents with a new pipeline version */
func (s *PipelineService) ReplayPipeline(ctx context.Context, documentIDs []uuid.UUID, newPipelineID uuid.UUID, newVersion string) error {
	// Get new pipeline config
	newPipeline, err := s.GetPipeline(ctx, newPipelineID, newVersion)
	if err != nil {
		return fmt.Errorf("failed to get new pipeline: %w", err)
	}

	// For each document, reprocess with new pipeline
	for _, docID := range documentIDs {
		// Get document
		docQuery := `SELECT content FROM neuronip.knowledge_documents WHERE id = $1`
		var content string
		err := s.pool.QueryRow(ctx, docQuery, docID).Scan(&content)
		if err != nil {
			continue
		}

		// Delete old embeddings
		deleteQuery := `DELETE FROM neuronip.knowledge_embeddings WHERE document_id = $1`
		s.pool.Exec(ctx, deleteQuery, docID)

		// Rechunk with new config
		chunks := ChunkText(content, newPipeline.ChunkingConfig)

		// Regenerate embeddings with new model
		modelName := newPipeline.EmbeddingModel
		if modelName == "" {
			modelName = "sentence-transformers/all-MiniLM-L6-v2" // Default model
		}
		
		// Generate embeddings for each chunk
		for chunkIdx, chunk := range chunks {
			var embedding string
			var err error
			
			// Use NeuronDB client if available, otherwise use SQL function
			if s.neurondbClient != nil {
				embedding, err = s.neurondbClient.GenerateEmbedding(ctx, chunk, modelName)
				if err != nil {
					// Log error but continue with other chunks
					continue
				}
			} else {
				// Fallback: use PostgreSQL embedding function if available
				// This requires pgvector extension with embedding functions
				embedQuery := `SELECT neurondb_embed($1, $2)`
				err = s.pool.QueryRow(ctx, embedQuery, chunk, modelName).Scan(&embedding)
				if err != nil {
					// If embedding function not available, skip this chunk
					continue
				}
			}
			
			// Store embedding
			embeddingID := uuid.New()
			insertEmbeddingQuery := `
				INSERT INTO neuronip.knowledge_embeddings 
				(id, document_id, chunk_index, chunk_text, embedding, embedding_model, created_at)
				VALUES ($1, $2, $3, $4, $5::vector, $6, NOW())
			`
			_, err = s.pool.Exec(ctx, insertEmbeddingQuery,
				embeddingID, docID, chunkIdx, chunk, embedding, modelName)
			if err != nil {
				// Log error but continue
				continue
			}
		}
	}

	// Record replay
	replayQuery := `
		INSERT INTO neuronip.pipeline_replays (
			document_ids, old_pipeline_id, new_pipeline_id, new_version, replayed_at
		) VALUES ($1, NULL, $2, $3, NOW())
	`
	docIDsJSON, _ := json.Marshal(documentIDs)
	_, err = s.pool.Exec(ctx, replayQuery, docIDsJSON, newPipelineID, newVersion)
	return err
}

/* ActivatePipeline activates a pipeline version */
func (s *PipelineService) ActivatePipeline(ctx context.Context, pipelineID uuid.UUID, version string) error {
	// Deactivate all other versions
	deactivateQuery := `UPDATE neuronip.pipelines SET is_active = false WHERE id = $1`
	s.pool.Exec(ctx, deactivateQuery, pipelineID)

	// Activate this version
	activateQuery := `UPDATE neuronip.pipelines SET is_active = true WHERE id = $1 AND version = $2`
	_, err := s.pool.Exec(ctx, activateQuery, pipelineID, version)
	return err
}

/* RecordPipelineMetrics records performance metrics for a pipeline */
func (s *PipelineService) RecordPipelineMetrics(ctx context.Context, pipelineID uuid.UUID, version string, metrics map[string]interface{}) error {
	metricsJSON, _ := json.Marshal(metrics)
	query := `UPDATE neuronip.pipelines SET performance_metrics = $1 WHERE id = $2 AND version = $3`
	_, err := s.pool.Exec(ctx, query, metricsJSON, pipelineID, version)
	return err
}
