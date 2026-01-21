package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* BackupService provides backup and restore functionality */
type BackupService struct {
	pool       *pgxpool.Pool
	backupDir  string
	config     BackupConfig
}

/* BackupConfig represents backup configuration */
type BackupConfig struct {
	BackupDir      string
	RetentionDays  int
	ScheduleCron   string
	Compress       bool
	IncludeFiles   bool
}

/* NewBackupService creates a new backup service */
func NewBackupService(pool *pgxpool.Pool, config BackupConfig) *BackupService {
	if config.BackupDir == "" {
		config.BackupDir = "/var/backups/neuronip"
	}
	if config.RetentionDays == 0 {
		config.RetentionDays = 30
	}

	return &BackupService{
		pool:      pool,
		backupDir: config.BackupDir,
		config:    config,
	}
}

/* Backup represents a backup record */
type Backup struct {
	ID           uuid.UUID              `json:"id"`
	Type         string                 `json:"type"` // "full", "incremental", "config"
	Status       string                 `json:"status"` // "running", "completed", "failed"
	FilePath     string                 `json:"file_path"`
	SizeBytes    int64                  `json:"size_bytes"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

/* CreateFullBackup creates a full database backup */
func (s *BackupService) CreateFullBackup(ctx context.Context) (*Backup, error) {
	backupID := uuid.New()
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(s.backupDir, fmt.Sprintf("full_%s_%s.sql", timestamp, backupID.String()))

	// Ensure backup directory exists
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Get database connection config
	config := s.pool.Config()
	connString := config.ConnString()

	// Create backup record
	backup := &Backup{
		ID:        backupID,
		Type:      "full",
		Status:    "running",
		FilePath:  backupFile,
		StartedAt: time.Now(),
	}

	// Save backup record to database
	if err := s.saveBackupRecord(ctx, backup); err != nil {
		return nil, fmt.Errorf("failed to save backup record: %w", err)
	}

	// Execute pg_dump
	cmd := exec.CommandContext(ctx, "pg_dump", connString, "-f", backupFile)
	if err := cmd.Run(); err != nil {
		backup.Status = "failed"
		backup.Error = err.Error()
		s.updateBackupRecord(ctx, backup)
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	// Get file size
	fileInfo, err := os.Stat(backupFile)
	if err == nil {
		backup.SizeBytes = fileInfo.Size()
	}

	// Compress if enabled
	if s.config.Compress {
		if err := s.compressBackup(backupFile); err != nil {
			backup.Error = fmt.Sprintf("compression failed: %v", err)
		} else {
			backup.FilePath += ".gz"
		}
	}

	completedAt := time.Now()
	backup.CompletedAt = &completedAt
	backup.Status = "completed"

	if err := s.updateBackupRecord(ctx, backup); err != nil {
		return nil, fmt.Errorf("failed to update backup record: %w", err)
	}

	return backup, nil
}

/* RestoreBackup restores a backup */
func (s *BackupService) RestoreBackup(ctx context.Context, backupID uuid.UUID) error {
	// Get backup record
	backup, err := s.getBackupRecord(ctx, backupID)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	if backup.Status != "completed" {
		return fmt.Errorf("backup is not completed")
	}

	// Check if file exists
	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backup.FilePath)
	}

	// Get database connection config
	config := s.pool.Config()
	connString := config.ConnString()

	// Decompress if needed
	restoreFile := backup.FilePath
	if filepath.Ext(backup.FilePath) == ".gz" {
		restoreFile, err = s.decompressBackup(backup.FilePath)
		if err != nil {
			return fmt.Errorf("failed to decompress backup: %w", err)
		}
		defer os.Remove(restoreFile)
	}

	// Execute pg_restore or psql depending on format
	var cmd *exec.Cmd
	if filepath.Ext(restoreFile) == ".sql" {
		cmd = exec.CommandContext(ctx, "psql", connString, "-f", restoreFile)
	} else {
		cmd = exec.CommandContext(ctx, "pg_restore", "-d", connString, restoreFile)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	return nil
}

/* ListBackups lists all backups */
func (s *BackupService) ListBackups(ctx context.Context, limit int) ([]Backup, error) {
	query := `
		SELECT id, type, status, file_path, size_bytes, started_at, completed_at, error, metadata
		FROM neuronip.backups
		ORDER BY started_at DESC
		LIMIT $1`

	if limit == 0 {
		limit = 100
	}

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}
	defer rows.Close()

	backups := []Backup{}
	for rows.Next() {
		var backup Backup
		var completedAt sql.NullTime
		var errorMsg sql.NullString

		err := rows.Scan(
			&backup.ID, &backup.Type, &backup.Status, &backup.FilePath,
			&backup.SizeBytes, &backup.StartedAt, &completedAt, &errorMsg, &backup.Metadata,
		)
		if err != nil {
			continue
		}

		if completedAt.Valid {
			backup.CompletedAt = &completedAt.Time
		}
		if errorMsg.Valid {
			backup.Error = errorMsg.String
		}

		backups = append(backups, backup)
	}

	return backups, nil
}

/* CleanupOldBackups removes backups older than retention period */
func (s *BackupService) CleanupOldBackups(ctx context.Context) error {
	cutoffDate := time.Now().AddDate(0, 0, -s.config.RetentionDays)

	query := `
		SELECT id, file_path FROM neuronip.backups
		WHERE completed_at < $1 AND status = 'completed'`

	rows, err := s.pool.Query(ctx, query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to query old backups: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var backupID uuid.UUID
		var filePath string

		if err := rows.Scan(&backupID, &filePath); err != nil {
			continue
		}

		// Delete file
		os.Remove(filePath)

		// Delete record
		deleteQuery := `DELETE FROM neuronip.backups WHERE id = $1`
		s.pool.Exec(ctx, deleteQuery, backupID)
	}

	return nil
}

/* saveBackupRecord saves a backup record to database */
func (s *BackupService) saveBackupRecord(ctx context.Context, backup *Backup) error {
	metadataJSON, _ := json.Marshal(backup.Metadata)

	query := `
		INSERT INTO neuronip.backups
		(id, type, status, file_path, size_bytes, started_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.pool.Exec(ctx, query,
		backup.ID, backup.Type, backup.Status, backup.FilePath,
		backup.SizeBytes, backup.StartedAt, metadataJSON,
	)
	return err
}

/* updateBackupRecord updates a backup record */
func (s *BackupService) updateBackupRecord(ctx context.Context, backup *Backup) error {
	metadataJSON, _ := json.Marshal(backup.Metadata)

	var completedAt sql.NullTime
	if backup.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *backup.CompletedAt, Valid: true}
	}

	query := `
		UPDATE neuronip.backups
		SET status = $1, size_bytes = $2, completed_at = $3, error = $4, metadata = $5
		WHERE id = $6`

	_, err := s.pool.Exec(ctx, query,
		backup.Status, backup.SizeBytes, completedAt, backup.Error,
		metadataJSON, backup.ID,
	)
	return err
}

/* getBackupRecord gets a backup record */
func (s *BackupService) getBackupRecord(ctx context.Context, backupID uuid.UUID) (*Backup, error) {
	var backup Backup
	var completedAt sql.NullTime
	var errorMsg sql.NullString
	var metadataJSON []byte

	query := `
		SELECT id, type, status, file_path, size_bytes, started_at, completed_at, error, metadata
		FROM neuronip.backups
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, backupID).Scan(
		&backup.ID, &backup.Type, &backup.Status, &backup.FilePath,
		&backup.SizeBytes, &backup.StartedAt, &completedAt, &errorMsg, &metadataJSON,
	)
	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		backup.CompletedAt = &completedAt.Time
	}
	if errorMsg.Valid {
		backup.Error = errorMsg.String
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &backup.Metadata)
	}

	return &backup, nil
}

/* compressBackup compresses a backup file */
func (s *BackupService) compressBackup(filePath string) error {
	cmd := exec.Command("gzip", filePath)
	return cmd.Run()
}

/* decompressBackup decompresses a backup file */
func (s *BackupService) decompressBackup(filePath string) (string, error) {
	outputPath := filePath[:len(filePath)-3] // Remove .gz
	cmd := exec.Command("gunzip", "-c", filePath)
	
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer outputFile.Close()

	cmd.Stdout = outputFile
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return outputPath, nil
}
