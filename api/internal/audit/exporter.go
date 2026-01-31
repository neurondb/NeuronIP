package audit

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ExporterService provides audit export functionality */
type ExporterService struct {
	pool *pgxpool.Pool
}

/* NewExporterService creates a new exporter service */
func NewExporterService(pool *pgxpool.Pool) *ExporterService {
	return &ExporterService{pool: pool}
}

/* ExportRequest represents an audit export request */
type ExportRequest struct {
	ResourceType  *string
	ResourceID    *uuid.UUID
	UserID        *string
	StartDate     *time.Time
	EndDate       *time.Time
	Format        string // "csv", "json"
	IncludeFields []string
}

/* ExportAuditLogs exports audit logs */
func (es *ExporterService) ExportAuditLogs(ctx context.Context, req ExportRequest, writer io.Writer) error {
	// Build query
	query, args := es.buildExportQuery(req)

	rows, err := es.pool.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	switch req.Format {
	case "csv":
		return es.exportCSV(rows, writer, req.IncludeFields)
	case "json":
		return es.exportJSON(rows, writer, req.IncludeFields)
	default:
		return fmt.Errorf("unsupported format: %s", req.Format)
	}
}

/* buildExportQuery builds the export query */
func (es *ExporterService) buildExportQuery(req ExportRequest) (string, []interface{}) {
	query := `
		SELECT id, resource_type, resource_id, user_id, action, details, ip_address, user_agent, created_at
		FROM neuronip.audit_logs
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if req.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, *req.ResourceType)
		argIndex++
	}

	if req.ResourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, *req.ResourceID)
		argIndex++
	}

	if req.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *req.UserID)
		argIndex++
	}

	if req.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *req.StartDate)
		argIndex++
	}

	if req.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *req.EndDate)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	return query, args
}

/* exportCSV exports audit logs as CSV */
func (es *ExporterService) exportCSV(rows interface{}, writer io.Writer, includeFields []string) error {
	r, ok := rows.(pgx.Rows)
	if !ok {
		return fmt.Errorf("invalid rows type for CSV export")
	}
	defer r.Close()
	cw := csv.NewWriter(writer)
	fields := r.FieldDescriptions()
	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = string(f.Name)
	}
	if err := cw.Write(headers); err != nil {
		return err
	}
	for r.Next() {
		vals, err := r.Values()
		if err != nil {
			return err
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			row[i] = fmt.Sprintf("%v", v)
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

/* exportJSON exports audit logs as JSON */
func (es *ExporterService) exportJSON(rows interface{}, writer io.Writer, includeFields []string) error {
	r, ok := rows.(pgx.Rows)
	if !ok {
		return fmt.Errorf("invalid rows type for JSON export")
	}
	defer r.Close()
	fields := r.FieldDescriptions()
	var out []map[string]interface{}
	for r.Next() {
		vals, err := r.Values()
		if err != nil {
			return err
		}
		row := make(map[string]interface{})
		for i, f := range fields {
			row[string(f.Name)] = vals[i]
		}
		out = append(out, row)
	}
	return json.NewEncoder(writer).Encode(out)
}
