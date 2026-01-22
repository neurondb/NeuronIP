package ingestion

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/ingestion/cdc"
	"github.com/neurondb/NeuronIP/api/internal/ingestion/etl"
	"github.com/neurondb/NeuronIP/api/internal/mcp"
)

/* Service provides data ingestion functionality */
type IngestionService struct {
	pool          *pgxpool.Pool
	connectorPool *ConnectorPool
	cdcManager    *cdc.CDCManager
	etlEngine     *etl.ETLEngine
	mcpClient     *mcp.Client
}

/* RegisterConnectorFactory registers a connector factory */
func (s *IngestionService) RegisterConnectorFactory(connectorType string, factory ConnectorFactory) {
	s.connectorPool.factories[connectorType] = factory
}

/* NewService creates a new ingestion service */
func NewService(pool *pgxpool.Pool, mcpClient *mcp.Client) *IngestionService {
	return &IngestionService{
		pool:          pool,
		connectorPool: NewConnectorPool(10),
		cdcManager:    cdc.NewCDCManager(pool),
		etlEngine:     etl.NewETLEngine(),
		mcpClient:     mcpClient,
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
	
	// Handle incremental sync with watermark
	if syncOptions.Mode == SyncModeIncremental {
		if watermark, ok := job.Config["watermark_column"].(string); ok && watermark != "" {
			lastWatermark, _ := s.getLastWatermark(ctx, job.DataSourceID, watermark)
			if lastWatermark != "" {
				if t, err := time.Parse(time.RFC3339, lastWatermark); err == nil {
					syncOptions.Since = &t
				} else {
					// Try as string if not timestamp
					// Will be handled by connector
					syncOptions.Transformations = map[string]interface{}{
						"watermark_column": watermark,
						"watermark_value":  lastWatermark,
					}
				}
			}
		} else if since, ok := job.Config["since"].(string); ok {
			if t, err := time.Parse(time.RFC3339, since); err == nil {
				syncOptions.Since = &t
			}
		}
	} else if since, ok := job.Config["since"].(string); ok {
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
		// Handle retry logic
		if retryErr := s.handleJobFailure(ctx, jobID, err); retryErr != nil {
			return retryErr
		}
		// If retry succeeded, return early
		return nil
	}
	
	// Update watermark for incremental sync
	if syncOptions.Mode == SyncModeIncremental {
		if watermark, ok := job.Config["watermark_column"].(string); ok && watermark != "" {
			// Check checkpoint from result for watermark value
			if result.Checkpoint != nil {
				if lastValue, ok := result.Checkpoint["last_value"].(string); ok {
					s.updateWatermark(ctx, job.DataSourceID, watermark, lastValue)
				}
			}
		}
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

/* handleJobFailure handles job failures with retry logic */
func (s *IngestionService) handleJobFailure(ctx context.Context, jobID uuid.UUID, err error) error {
	job, jobErr := s.GetIngestionJob(ctx, jobID)
	if jobErr != nil {
		return fmt.Errorf("failed to get job for retry: %w", jobErr)
	}

	maxRetries := 3
	if max, ok := job.Config["max_retries"].(float64); ok {
		maxRetries = int(max)
	}

	currentRetryCount := 0
	if count, ok := job.Config["retry_count"].(float64); ok {
		currentRetryCount = int(count)
	}

	if currentRetryCount >= maxRetries {
		errMsg := err.Error()
		s.updateJobStatus(ctx, jobID, "failed", &errMsg)
		return fmt.Errorf("max retries exceeded: %w", err)
	}

	// Increment retry count
	job.Config["retry_count"] = currentRetryCount + 1
	configJSON, _ := json.Marshal(job.Config)
	
	backoffMs := 1000
	if backoff, ok := job.Config["retry_backoff_ms"].(float64); ok {
		backoffMs = int(backoff)
	}
	backoffDuration := time.Duration(backoffMs) * time.Millisecond * time.Duration(currentRetryCount+1)

	// Update job with retry info
	query := `
		UPDATE neuronip.ingestion_jobs 
		SET retry_count = $1, config = $2, last_error = $3, last_error_at = NOW(), updated_at = NOW()
		WHERE id = $4`
	errMsg := err.Error()
	s.pool.Exec(ctx, query, currentRetryCount+1, configJSON, errMsg, jobID)

	// Schedule retry (in production, use a job queue)
	time.Sleep(backoffDuration)
	return s.ExecuteSyncJob(ctx, jobID)
}

/* updateWatermark updates the watermark value for incremental sync */
func (s *IngestionService) updateWatermark(ctx context.Context, dataSourceID uuid.UUID, watermarkColumn string, watermarkValue string) error {
	query := `
		UPDATE neuronip.ingestion_jobs 
		SET watermark_column = $1, last_watermark_value = $2, updated_at = NOW()
		WHERE data_source_id = $3 AND incremental_sync_enabled = true`
	
	_, err := s.pool.Exec(ctx, query, watermarkColumn, watermarkValue, dataSourceID)
	return err
}

/* getLastWatermark retrieves the last watermark value for a data source */
func (s *IngestionService) getLastWatermark(ctx context.Context, dataSourceID uuid.UUID, watermarkColumn string) (string, error) {
	query := `
		SELECT last_watermark_value 
		FROM neuronip.ingestion_jobs
		WHERE data_source_id = $1 AND watermark_column = $2 AND last_watermark_value IS NOT NULL
		ORDER BY completed_at DESC
		LIMIT 1`
	
	var watermark string
	err := s.pool.QueryRow(ctx, query, dataSourceID, watermarkColumn).Scan(&watermark)
	if err != nil {
		return "", nil // No watermark found, start from beginning
	}
	return watermark, nil
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

/* LoadDatasetFromSource loads a dataset using MCP LoadDataset tool */
func (s *IngestionService) LoadDatasetFromSource(ctx context.Context, sourceType string, sourcePath string, targetTable string, options map[string]interface{}) (map[string]interface{}, error) {
	if s.mcpClient == nil {
		return nil, fmt.Errorf("MCP client not configured")
	}

	// Use MCP LoadDataset tool
		// Use MCP LoadDataset tool
		result, err := s.mcpClient.LoadDataset(ctx, sourceType, sourcePath, options)
		if err != nil {
			return nil, fmt.Errorf("failed to load dataset via MCP: %w", err)
		}
	if err != nil {
		return nil, fmt.Errorf("failed to load dataset: %w", err)
	}

	// If target table specified, could insert data into table
	// This would depend on the dataset format and target schema
	if targetTable != "" {
		// Store dataset metadata in ingestion jobs table
		jobConfig := map[string]interface{}{
			"source_type":  sourceType,
			"source_path":  sourcePath,
			"target_table": targetTable,
			"dataset_result": result,
		}
		// Create ingestion job record for tracking
		_, err := s.CreateIngestionJob(ctx, uuid.New(), "dataset_load", jobConfig)
		if err != nil {
			// Log error but continue
		}
	}

	return result, nil
}

/* IngestionStatus represents detailed ingestion status */
type IngestionStatus struct {
	DataSourceID      uuid.UUID              `json:"data_source_id"`
	TotalJobs         int                    `json:"total_jobs"`
	RunningJobs       int                    `json:"running_jobs"`
	CompletedJobs     int                    `json:"completed_jobs"`
	FailedJobs        int                    `json:"failed_jobs"`
	PendingJobs       int                    `json:"pending_jobs"`
	TotalRowsProcessed int64                 `json:"total_rows_processed"`
	LastSyncAt        *time.Time             `json:"last_sync_at,omitempty"`
	LastSyncStatus    string                 `json:"last_sync_status"`
	CDCEnabled        bool                   `json:"cdc_enabled"`
	CDCLag            *time.Duration         `json:"cdc_lag,omitempty"`
	RecentJobs        []IngestionJob         `json:"recent_jobs,omitempty"`
}

/* GetIngestionStatus retrieves detailed ingestion status for a data source */
func (s *IngestionService) GetIngestionStatus(ctx context.Context, dataSourceID uuid.UUID) (*IngestionStatus, error) {
	// Get job statistics
	statsQuery := `
		SELECT 
			COUNT(*) as total_jobs,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running_jobs,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_jobs,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_jobs,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_jobs,
			SUM(rows_processed) as total_rows_processed,
			MAX(completed_at) as last_sync_at
		FROM neuronip.ingestion_jobs
		WHERE data_source_id = $1`

	var status IngestionStatus
	status.DataSourceID = dataSourceID
	var lastSyncAt sql.NullTime

	err := s.pool.QueryRow(ctx, statsQuery, dataSourceID).Scan(
		&status.TotalJobs, &status.RunningJobs, &status.CompletedJobs,
		&status.FailedJobs, &status.PendingJobs, &status.TotalRowsProcessed, &lastSyncAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ingestion status: %w", err)
	}

	if lastSyncAt.Valid {
		status.LastSyncAt = &lastSyncAt.Time
	}

	// Get last job status
	lastJobQuery := `
		SELECT status
		FROM neuronip.ingestion_jobs
		WHERE data_source_id = $1
		ORDER BY created_at DESC
		LIMIT 1`
	
	var lastStatus sql.NullString
	s.pool.QueryRow(ctx, lastJobQuery, dataSourceID).Scan(&lastStatus)
	if lastStatus.Valid {
		status.LastSyncStatus = lastStatus.String
	} else {
		status.LastSyncStatus = "never_synced"
	}

	// Check CDC status (table may not exist, so handle gracefully)
	cdcQuery := `
		SELECT COUNT(*) > 0 as cdc_enabled
		FROM neuronip.cdc_replications
		WHERE data_source_id = $1 AND enabled = true`
	
	var cdcEnabled bool
	err = s.pool.QueryRow(ctx, cdcQuery, dataSourceID).Scan(&cdcEnabled)
	if err != nil {
		// CDC table may not exist, default to false
		status.CDCEnabled = false
	} else {
		status.CDCEnabled = cdcEnabled
	}

	// Get CDC lag if enabled
	if status.CDCEnabled {
		lagQuery := `
			SELECT EXTRACT(EPOCH FROM (NOW() - MAX(last_replicated_at))) * 1000 as lag_ms
			FROM neuronip.cdc_replications
			WHERE data_source_id = $1 AND enabled = true`
		
		var lagMs sql.NullInt64
		if err := s.pool.QueryRow(ctx, lagQuery, dataSourceID).Scan(&lagMs); err == nil && lagMs.Valid {
			lagDuration := time.Duration(lagMs.Int64) * time.Millisecond
			status.CDCLag = &lagDuration
		}
	}

	// Get recent jobs
	recentJobs, _ := s.ListIngestionJobs(ctx, &dataSourceID, 10)
	status.RecentJobs = recentJobs

	return &status, nil
}

/* IngestionFailure represents a failed ingestion job */
type IngestionFailure struct {
	JobID            uuid.UUID              `json:"job_id"`
	DataSourceID     uuid.UUID              `json:"data_source_id"`
	JobType          string                 `json:"job_type"`
	ErrorMessage     string                 `json:"error_message"`
	FailedAt         time.Time              `json:"failed_at"`
	RetryCount       int                    `json:"retry_count"`
	Config           map[string]interface{} `json:"config,omitempty"`
}

/* GetIngestionFailures retrieves failed ingestion jobs (DLQ) */
func (s *IngestionService) GetIngestionFailures(ctx context.Context, dataSourceID *uuid.UUID, limit int) ([]IngestionFailure, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, data_source_id, job_type, error_message, updated_at, retry_count, config
		FROM neuronip.ingestion_jobs
		WHERE status = 'failed'`

	args := []interface{}{}
	argIndex := 1

	if dataSourceID != nil {
		query += fmt.Sprintf(" AND data_source_id = $%d", argIndex)
		args = append(args, *dataSourceID)
		argIndex++
	}

	query += " ORDER BY updated_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get ingestion failures: %w", err)
	}
	defer rows.Close()

	var failures []IngestionFailure
	for rows.Next() {
		var failure IngestionFailure
		var configJSON json.RawMessage
		var errorMsg sql.NullString

		err := rows.Scan(
			&failure.JobID, &failure.DataSourceID, &failure.JobType,
			&errorMsg, &failure.FailedAt, &failure.RetryCount, &configJSON,
		)
		if err != nil {
			continue
		}

		if errorMsg.Valid {
			failure.ErrorMessage = errorMsg.String
		}
		if configJSON != nil {
			json.Unmarshal(configJSON, &failure.Config)
		}

		failures = append(failures, failure)
	}

	return failures, nil
}

/* RetryIngestionJob retries a failed ingestion job */
func (s *IngestionService) RetryIngestionJob(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.GetIngestionJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job.Status != "failed" {
		return fmt.Errorf("job is not in failed state")
	}

	// Reset job status and retry
	query := `
		UPDATE neuronip.ingestion_jobs 
		SET status = 'pending', error_message = NULL, updated_at = NOW()
		WHERE id = $1`
	
	_, err = s.pool.Exec(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to reset job status: %w", err)
	}

	// Execute the job again
	return s.ExecuteSyncJob(ctx, jobID)
}
