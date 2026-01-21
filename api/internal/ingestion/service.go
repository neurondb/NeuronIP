package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/ingestion/cdc"
	"github.com/neurondb/NeuronIP/api/internal/ingestion/etl"
)

/* Service provides data ingestion functionality */
type IngestionService struct {
	pool         *pgxpool.Pool
	connectorPool *ConnectorPool
	cdcManager   *cdc.CDCManager
	etlEngine    *etl.ETLEngine
}

/* RegisterConnectorFactory registers a connector factory */
func (s *IngestionService) RegisterConnectorFactory(connectorType string, factory ConnectorFactory) {
	s.connectorPool.factories[connectorType] = factory
}

/* NewService creates a new ingestion service */
func NewService(pool *pgxpool.Pool) *IngestionService {
	return &IngestionService{
		pool:          pool,
		connectorPool: NewConnectorPool(10),
		cdcManager:    cdc.NewCDCManager(pool),
		etlEngine:     etl.NewETLEngine(),
	}
}

/* CreateIngestionJob creates a new ingestion job */
func (s *IngestionService) CreateIngestionJob(ctx context.Context, dataSourceID uuid.UUID, jobType string, config map[string]interface{}) (*IngestionJob, error) {
	jobID := uuid.New()
	configJSON, _ := json.Marshal(config)
	now := time.Now()
	
	query := `
		INSERT INTO neuronip.ingestion_jobs 
		(id, data_source_id, job_type, status, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, data_source_id, job_type, status, config, progress, error_message, 
		          rows_processed, started_at, completed_at, created_at, updated_at`
	
	var job IngestionJob
	var configJSONRaw, progressJSON json.RawMessage
	
	err := s.pool.QueryRow(ctx, query, jobID, dataSourceID, jobType, "pending", configJSON, now, now).Scan(
		&job.ID, &job.DataSourceID, &job.JobType, &job.Status, &configJSONRaw, &progressJSON,
		&job.ErrorMessage, &job.RowsProcessed, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ingestion job: %w", err)
	}
	
	json.Unmarshal(configJSONRaw, &job.Config)
	if progressJSON != nil {
		json.Unmarshal(progressJSON, &job.Progress)
	}
	
	return &job, nil
}

/* ExecuteSyncJob executes a sync job */
func (s *IngestionService) ExecuteSyncJob(ctx context.Context, jobID uuid.UUID) error {
	// Get job
	job, err := s.GetIngestionJob(ctx, jobID)
	if err != nil {
		return err
	}
	
	// Get data source
	var dataSourceID uuid.UUID
	var connectorType string
	var connectorConfigJSON json.RawMessage
	
	query := `SELECT id, connector_type, connector_config FROM neuronip.data_sources WHERE id = $1`
	err = s.pool.QueryRow(ctx, query, job.DataSourceID).Scan(&dataSourceID, &connectorType, &connectorConfigJSON)
	if err != nil {
		return fmt.Errorf("failed to get data source: %w", err)
	}
	
	var connectorConfig map[string]interface{}
	json.Unmarshal(connectorConfigJSON, &connectorConfig)
	
	// Get connector
	connector, err := s.getConnector(connectorType, connectorConfig)
	if err != nil {
		return fmt.Errorf("failed to get connector: %w", err)
	}
	
	// Update job status
	s.updateJobStatus(ctx, jobID, "running", nil)
	
	// Execute sync
	syncOptions := SyncOptions{
		Mode: SyncModeFull,
	}
	if mode, ok := job.Config["mode"].(string); ok {
		syncOptions.Mode = SyncMode(mode)
	}
	if since, ok := job.Config["since"].(string); ok {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			syncOptions.Since = &t
		}
	}
	if tables, ok := job.Config["tables"].([]interface{}); ok {
		syncOptions.Tables = make([]string, 0, len(tables))
		for _, t := range tables {
			if tStr, ok := t.(string); ok {
				syncOptions.Tables = append(syncOptions.Tables, tStr)
			}
		}
	}
	
	result, err := connector.Sync(ctx, syncOptions)
	if err != nil {
		errMsg := err.Error()
		s.updateJobStatus(ctx, jobID, "failed", &errMsg)
		return err
	}
	
	// Update job with results
	progress := map[string]interface{}{
		"rows_synced":   result.RowsSynced,
		"tables_synced": result.TablesSynced,
		"duration":      result.Duration.String(),
	}
	
	s.updateJobComplete(ctx, jobID, result.RowsSynced, progress)
	
	return nil
}

/* getConnector gets or creates a connector */
func (s *IngestionService) getConnector(connectorType string, config map[string]interface{}) (Connector, error) {
	// Create connector factory based on type
	// This avoids import cycle by not importing connectors package here
	// Connectors will be registered in main.go or a separate init function
	factory := func(ct string) Connector {
		// Return nil - connectors should be registered via RegisterConnectorFactory
		return nil
	}
	
	// Check if factory is registered
	if f, exists := s.connectorPool.factories[connectorType]; exists {
		factory = f
	} else {
		return nil, fmt.Errorf("connector type not registered: %s", connectorType)
	}
	
	conn, err := s.connectorPool.GetConnector(context.Background(), connectorType, factory, config)
	if err != nil {
		return nil, err
	}
	
	return conn, nil
}

/* updateJobStatus updates job status */
func (s *IngestionService) updateJobStatus(ctx context.Context, jobID uuid.UUID, status string, errorMsg *string) {
	query := `
		UPDATE neuronip.ingestion_jobs 
		SET status = $1, error_message = $2, updated_at = NOW()
		WHERE id = $3`
	
	if status == "running" {
		query = `
			UPDATE neuronip.ingestion_jobs 
			SET status = $1, started_at = NOW(), updated_at = NOW()
			WHERE id = $3`
		s.pool.Exec(ctx, query, status, nil, jobID)
	} else {
		s.pool.Exec(ctx, query, status, errorMsg, jobID)
	}
}

/* updateJobComplete updates job as completed */
func (s *IngestionService) updateJobComplete(ctx context.Context, jobID uuid.UUID, rowsProcessed int64, progress map[string]interface{}) {
	progressJSON, _ := json.Marshal(progress)
	
	query := `
		UPDATE neuronip.ingestion_jobs 
		SET status = $1, rows_processed = $2, progress = $3, completed_at = NOW(), updated_at = NOW()
		WHERE id = $4`
	
	s.pool.Exec(ctx, query, "completed", rowsProcessed, progressJSON, jobID)
}

/* GetIngestionJob retrieves an ingestion job */
func (s *IngestionService) GetIngestionJob(ctx context.Context, jobID uuid.UUID) (*IngestionJob, error) {
	query := `
		SELECT id, data_source_id, job_type, status, config, progress, error_message,
		       rows_processed, started_at, completed_at, created_at, updated_at
		FROM neuronip.ingestion_jobs
		WHERE id = $1`
	
	var job IngestionJob
	var configJSON, progressJSON json.RawMessage
	
	err := s.pool.QueryRow(ctx, query, jobID).Scan(
		&job.ID, &job.DataSourceID, &job.JobType, &job.Status, &configJSON, &progressJSON,
		&job.ErrorMessage, &job.RowsProcessed, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ingestion job: %w", err)
	}
	
	json.Unmarshal(configJSON, &job.Config)
	if progressJSON != nil {
		json.Unmarshal(progressJSON, &job.Progress)
	}
	
	return &job, nil
}

/* ListIngestionJobs lists ingestion jobs */
func (s *IngestionService) ListIngestionJobs(ctx context.Context, dataSourceID *uuid.UUID, limit int) ([]IngestionJob, error) {
	query := `
		SELECT id, data_source_id, job_type, status, config, progress, error_message,
		       rows_processed, started_at, completed_at, created_at, updated_at
		FROM neuronip.ingestion_jobs`
	
	args := []interface{}{}
	argIndex := 1
	
	if dataSourceID != nil {
		query += fmt.Sprintf(" WHERE data_source_id = $%d", argIndex)
		args = append(args, *dataSourceID)
		argIndex++
	}
	
	query += " ORDER BY created_at DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}
	
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ingestion jobs: %w", err)
	}
	defer rows.Close()
	
	jobs := make([]IngestionJob, 0)
	for rows.Next() {
		var job IngestionJob
		var configJSON, progressJSON json.RawMessage
		
		err := rows.Scan(
			&job.ID, &job.DataSourceID, &job.JobType, &job.Status, &configJSON, &progressJSON,
			&job.ErrorMessage, &job.RowsProcessed, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		json.Unmarshal(configJSON, &job.Config)
		if progressJSON != nil {
			json.Unmarshal(progressJSON, &job.Progress)
		}
		
		jobs = append(jobs, job)
	}
	
	return jobs, nil
}

/* IngestionJob represents an ingestion job */
type IngestionJob struct {
	ID            uuid.UUID              `json:"id"`
	DataSourceID  uuid.UUID              `json:"data_source_id"`
	JobType       string                 `json:"job_type"`
	Status        string                 `json:"status"`
	Config        map[string]interface{} `json:"config"`
	Progress      map[string]interface{} `json:"progress,omitempty"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
	RowsProcessed int64                  `json:"rows_processed"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}
