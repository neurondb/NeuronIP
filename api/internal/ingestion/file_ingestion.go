package ingestion

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/ingestion/parsers"
)

/* FileIngestionService provides file ingestion functionality */
type FileIngestionService struct {
	pool *pgxpool.Pool
}

/* NewFileIngestionService creates a new file ingestion service */
func NewFileIngestionService(pool *pgxpool.Pool) *FileIngestionService {
	return &FileIngestionService{pool: pool}
}

/* FileIngestionJob represents a file ingestion job */
type FileIngestionJob struct {
	ID            uuid.UUID              `json:"id"`
	FilePath      string                 `json:"file_path"`
	FileName      string                 `json:"file_name"`
	FileType      string                 `json:"file_type"`
	ParserType    string                 `json:"parser_type"`
	Status        string                 `json:"status"`
	DataSourceID  *uuid.UUID             `json:"data_source_id,omitempty"`
	RowsProcessed int                    `json:"rows_processed"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy     *string                `json:"created_by,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

/* CreateFileIngestionJob creates a new file ingestion job */
func (s *FileIngestionService) CreateFileIngestionJob(ctx context.Context, file multipart.File, header *multipart.FileHeader, dataSourceID *uuid.UUID, createdBy *string) (*FileIngestionJob, error) {
	jobID := uuid.New()
	fileName := header.Filename
	
	// Determine file type from extension
	fileType := "csv"
	if len(fileName) > 4 {
		ext := fileName[len(fileName)-4:]
		switch ext {
		case ".pdf":
			fileType = "pdf"
		case ".xlsx", ".xls":
			fileType = "xlsx"
		case ".docx", ".doc":
			fileType = "docx"
		case ".pptx", ".ppt":
			fileType = "pptx"
		case ".json":
			fileType = "json"
		case ".csv":
			fileType = "csv"
		}
	}

	// Determine parser type
	parserType := "csv"
	switch fileType {
	case "pdf":
		parserType = "pdf"
	case "xlsx", "docx", "pptx":
		parserType = "office"
	case "json":
		parserType = "json"
	}

	now := time.Now()
	query := `
		INSERT INTO neuronip.file_ingestion_jobs 
		(id, file_path, file_name, file_type, parser_type, status, data_source_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7, $8, $9)
		RETURNING id, file_path, file_name, file_type, parser_type, status, data_source_id, rows_processed, error_message, metadata, created_by, created_at, updated_at, started_at, completed_at`

	var job FileIngestionJob
	var metadataJSON json.RawMessage
	var errMsg sql.NullString

	err := s.pool.QueryRow(ctx, query, jobID, fileName, fileName, fileType, parserType, dataSourceID, createdBy, now, now).Scan(
		&job.ID, &job.FilePath, &job.FileName, &job.FileType, &job.ParserType, &job.Status,
		&job.DataSourceID, &job.RowsProcessed, &errMsg, &metadataJSON,
		&job.CreatedBy, &job.CreatedAt, &job.UpdatedAt, &job.StartedAt, &job.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create file ingestion job: %w", err)
	}

	if errMsg.Valid {
		job.ErrorMessage = &errMsg.String
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &job.Metadata)
	}

	return &job, nil
}

/* ProcessFileIngestionJob processes a file ingestion job */
func (s *FileIngestionService) ProcessFileIngestionJob(ctx context.Context, jobID uuid.UUID, file io.Reader) error {
	job, err := s.GetFileIngestionJob(ctx, jobID)
	if err != nil {
		return err
	}

	// Update status to processing
	now := time.Now()
	s.updateFileJobStatus(ctx, jobID, "processing", nil, &now)

	// Parse file based on type
	var rowsProcessed int
	var parseErr error

	switch job.ParserType {
	case "pdf":
		pdfParser := parsers.NewPDFParser(false)
		content, parseErr := pdfParser.Parse(file)
		if parseErr == nil && content != nil {
			rowsProcessed = len(content.Pages)
			job.Metadata = map[string]interface{}{
				"page_count": len(content.Pages),
				"extracted_text_length": len(content.Text),
			}
		}
	case "office":
		officeParser := parsers.NewOfficeParser(false)
		content, parseErr := officeParser.Parse(file, job.FileType)
		if parseErr == nil && content != nil {
			rowsProcessed = len(content.Sections)
			job.Metadata = map[string]interface{}{
				"section_count": len(content.Sections),
				"extracted_text_length": len(content.Text),
			}
		}
	case "csv":
		csvParser := parsers.NewCSVParser(true, ',')
		records, schema, parseErr := csvParser.Parse(file)
		if parseErr == nil {
			rowsProcessed = len(records)
			job.Metadata = map[string]interface{}{
				"row_count": rowsProcessed,
				"column_count": len(schema.Columns),
			}
		}
	default:
		parseErr = fmt.Errorf("unsupported parser type: %s", job.ParserType)
	}

	if parseErr != nil {
		errMsg := parseErr.Error()
		s.updateFileJobStatus(ctx, jobID, "failed", &errMsg, nil)
		return fmt.Errorf("failed to parse file: %w", parseErr)
	}

	// Update job as completed
	completedAt := time.Now()
	s.updateFileJobStatus(ctx, jobID, "completed", nil, &completedAt)
	
	updateQuery := `
		UPDATE neuronip.file_ingestion_jobs 
		SET rows_processed = $1, metadata = $2, completed_at = $3, updated_at = $3
		WHERE id = $4`
	metadataJSON, _ := json.Marshal(job.Metadata)
	s.pool.Exec(ctx, updateQuery, rowsProcessed, metadataJSON, completedAt, jobID)

	return nil
}

/* GetFileIngestionJob retrieves a file ingestion job */
func (s *FileIngestionService) GetFileIngestionJob(ctx context.Context, jobID uuid.UUID) (*FileIngestionJob, error) {
	query := `
		SELECT id, file_path, file_name, file_type, parser_type, status, data_source_id, rows_processed, error_message, metadata, created_by, created_at, updated_at, started_at, completed_at
		FROM neuronip.file_ingestion_jobs
		WHERE id = $1`

	var job FileIngestionJob
	var metadataJSON json.RawMessage
	var errMsg sql.NullString

	err := s.pool.QueryRow(ctx, query, jobID).Scan(
		&job.ID, &job.FilePath, &job.FileName, &job.FileType, &job.ParserType, &job.Status,
		&job.DataSourceID, &job.RowsProcessed, &errMsg, &metadataJSON,
		&job.CreatedBy, &job.CreatedAt, &job.UpdatedAt, &job.StartedAt, &job.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get file ingestion job: %w", err)
	}

	if errMsg.Valid {
		job.ErrorMessage = &errMsg.String
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &job.Metadata)
	}

	return &job, nil
}

/* ListFileIngestionJobs lists file ingestion jobs */
func (s *FileIngestionService) ListFileIngestionJobs(ctx context.Context, dataSourceID *uuid.UUID, limit int) ([]FileIngestionJob, error) {
	query := `
		SELECT id, file_path, file_name, file_type, parser_type, status, data_source_id, rows_processed, error_message, metadata, created_by, created_at, updated_at, started_at, completed_at
		FROM neuronip.file_ingestion_jobs`

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
		return nil, fmt.Errorf("failed to list file ingestion jobs: %w", err)
	}
	defer rows.Close()

	var jobs []FileIngestionJob
	for rows.Next() {
		var job FileIngestionJob
		var metadataJSON json.RawMessage
		var errMsg sql.NullString

		err := rows.Scan(
			&job.ID, &job.FilePath, &job.FileName, &job.FileType, &job.ParserType, &job.Status,
			&job.DataSourceID, &job.RowsProcessed, &errMsg, &metadataJSON,
			&job.CreatedBy, &job.CreatedAt, &job.UpdatedAt, &job.StartedAt, &job.CompletedAt,
		)
		if err != nil {
			continue
		}

		if errMsg.Valid {
			job.ErrorMessage = &errMsg.String
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &job.Metadata)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

/* updateFileJobStatus updates file job status */
func (s *FileIngestionService) updateFileJobStatus(ctx context.Context, jobID uuid.UUID, status string, errorMsg *string, startedAt *time.Time) {
	if startedAt != nil {
		query := `
			UPDATE neuronip.file_ingestion_jobs 
			SET status = $1, started_at = $2, updated_at = NOW()
			WHERE id = $3`
		s.pool.Exec(ctx, query, status, startedAt, jobID)
	} else {
		query := `
			UPDATE neuronip.file_ingestion_jobs 
			SET status = $1, error_message = $2, updated_at = NOW()
			WHERE id = $3`
		s.pool.Exec(ctx, query, status, errorMsg, jobID)
	}
}
