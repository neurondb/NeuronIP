package bi

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* BIExporterService provides BI export functionality */
type BIExporterService struct {
	pool *pgxpool.Pool
}

/* NewBIExporterService creates a new BI exporter service */
func NewBIExporterService(pool *pgxpool.Pool) *BIExporterService {
	return &BIExporterService{pool: pool}
}

/* ExportConfig represents a BI export configuration */
type ExportConfig struct {
	ID           uuid.UUID              `json:"id"`
	BIType       string                 `json:"bi_type"` // "tableau", "powerbi", "looker"
	QueryID      uuid.UUID              `json:"query_id"`
	ExportFormat string                 `json:"export_format"` // "csv", "json", "xlsx"
	Config       map[string]interface{} `json:"config"`
	Enabled      bool                   `json:"enabled"`
	CreatedAt    time.Time              `json:"created_at"`
}

/* ExportQuery exports a query result to BI format (json or csv). Fetches cached result from query_results; returns empty document if not found. */
func (bies *BIExporterService) ExportQuery(ctx context.Context, queryID uuid.UUID, biType, format string) ([]byte, error) {
	_ = biType

	var resultData json.RawMessage
	err := bies.pool.QueryRow(ctx, `
		SELECT result_data FROM neuronip.query_results WHERE query_id = $1 ORDER BY created_at DESC LIMIT 1
	`, queryID).Scan(&resultData)
	if err != nil || len(resultData) == 0 {
		// No cached result: return format-appropriate empty document
		switch format {
		case "json", "application/json":
			return []byte("[]"), nil
		case "csv", "text/csv":
			return []byte("data\n"), nil
		default:
			return []byte("[]"), nil
		}
	}

	var rows []map[string]interface{}
	if err := json.Unmarshal(resultData, &rows); err != nil {
		return nil, fmt.Errorf("failed to parse result_data: %w", err)
	}

	switch format {
	case "csv", "text/csv":
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		if len(rows) > 0 {
			headers := make([]string, 0, len(rows[0]))
			for k := range rows[0] {
				headers = append(headers, k)
			}
			sort.Strings(headers)
			_ = w.Write(headers)
			for _, row := range rows {
				record := make([]string, len(headers))
				for i, h := range headers {
					if v, ok := row[h]; ok && v != nil {
						record[i] = fmt.Sprintf("%v", v)
					}
				}
				_ = w.Write(record)
			}
		}
		w.Flush()
		return buf.Bytes(), nil
	case "json", "application/json":
		return []byte(resultData), nil
	default:
		return []byte(resultData), nil
	}
}

/* CreateExportConfig creates an export configuration */
func (bies *BIExporterService) CreateExportConfig(ctx context.Context, config ExportConfig) (*ExportConfig, error) {
	config.ID = uuid.New()
	config.CreatedAt = time.Now()

	configJSON, _ := json.Marshal(config.Config)

	query := `
		INSERT INTO neuronip.bi_export_configs 
		(id, bi_type, query_id, export_format, config, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := bies.pool.Exec(ctx, query,
		config.ID, config.BIType, config.QueryID, config.ExportFormat, configJSON, config.Enabled, config.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create export config: %w", err)
	}

	return &config, nil
}
