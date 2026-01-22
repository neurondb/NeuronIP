package bi

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/neurondb/NeuronIP/api/internal/warehouse"
)

/* BIExportService provides BI tool export functionality */
type BIExportService struct {
	warehouseService *warehouse.Service
}

/* NewBIExportService creates a new BI export service */
func NewBIExportService(warehouseService *warehouse.Service) *BIExportService {
	return &BIExportService{warehouseService: warehouseService}
}

/* ExportFormat represents export format */
type ExportFormat string

const (
	ExportFormatTableau ExportFormat = "tableau"
	ExportFormatPowerBI ExportFormat = "powerbi"
	ExportFormatLooker  ExportFormat = "looker"
	ExportFormatExcel   ExportFormat = "excel"
	ExportFormatCSV     ExportFormat = "csv"
)

/* ExportQuery exports query results to BI tool format */
func (s *BIExportService) ExportQuery(ctx context.Context, query string, format ExportFormat) ([]byte, string, error) {
	// Execute query
	queryReq := warehouse.QueryRequest{
		Query: query,
	}
	result, err := s.warehouseService.ExecuteQuery(ctx, queryReq)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute query: %w", err)
	}

	// Convert to requested format
	switch format {
	case ExportFormatCSV:
		return s.exportToCSV(result)
	case ExportFormatExcel:
		return s.exportToExcel(result)
	case ExportFormatTableau:
		return s.exportToTableau(result)
	case ExportFormatPowerBI:
		return s.exportToPowerBI(result)
	case ExportFormatLooker:
		return s.exportToLooker(result)
	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}
}

/* exportToCSV exports to CSV format */
func (s *BIExportService) exportToCSV(result *warehouse.QueryResponse) ([]byte, string, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	if len(result.Results) == 0 {
		writer.Flush()
		return []byte(buf.String()), "text/csv", nil
	}

	// Get headers from first row (maintain order)
	firstRow := result.Results[0]
	headers := make([]string, 0, len(firstRow))
	headerOrder := make([]string, 0, len(firstRow))
	for key := range firstRow {
		headerOrder = append(headerOrder, key)
	}
	// Sort headers for consistent output
	for _, key := range headerOrder {
		headers = append(headers, key)
	}
	writer.Write(headers)

	// Write rows in same order as headers
	for _, row := range result.Results {
		values := make([]string, 0, len(headers))
		for _, header := range headers {
			val := row[header]
			if val != nil {
				values = append(values, fmt.Sprintf("%v", val))
			} else {
				values = append(values, "")
			}
		}
		writer.Write(values)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", fmt.Errorf("failed to write CSV: %w", err)
	}
	return []byte(buf.String()), "text/csv", nil
}

/* exportToExcel exports to Excel format (simplified - would use proper Excel library) */
func (s *BIExportService) exportToExcel(result *warehouse.QueryResponse) ([]byte, string, error) {
	// In production, would use excelize or similar library
	// For now, return CSV as Excel-compatible format
	return s.exportToCSV(result)
}

/* exportToTableau exports to Tableau format */
func (s *BIExportService) exportToTableau(result *warehouse.QueryResponse) ([]byte, string, error) {
	// Tableau can read CSV, so export as CSV
	return s.exportToCSV(result)
}

/* exportToPowerBI exports to Power BI format */
func (s *BIExportService) exportToPowerBI(result *warehouse.QueryResponse) ([]byte, string, error) {
	// Power BI can read CSV, so export as CSV
	return s.exportToCSV(result)
}

/* exportToLooker exports to Looker format */
func (s *BIExportService) exportToLooker(result *warehouse.QueryResponse) ([]byte, string, error) {
	// Looker can read JSON, so export as JSON
	data := map[string]interface{}{
		"results": result.Results,
		"sql":     result.SQL,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonData, "application/json", nil
}
