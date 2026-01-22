package handlers

import (
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/errors"
	bibot "github.com/neurondb/NeuronIP/api/internal/integrations/bi"
)

/* BIExportHandler handles BI export requests */
type BIExportHandler struct {
	biExportService *bibot.BIExportService
}

/* NewBIExportHandler creates a new BI export handler */
func NewBIExportHandler(biExportService *bibot.BIExportService) *BIExportHandler {
	return &BIExportHandler{biExportService: biExportService}
}

/* ExportQuery handles GET /api/v1/integrations/bi/export */
func (h *BIExportHandler) ExportQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	format := r.URL.Query().Get("format")
	
	if query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("query parameter is required", nil))
		return
	}
	
	if format == "" {
		format = "csv" // Default to CSV
	}
	
	// Validate format
	validFormats := map[string]bool{
		"csv":     true,
		"excel":   true,
		"tableau": true,
		"powerbi": true,
		"looker":  true,
	}
	if !validFormats[format] {
		WriteErrorResponse(w, errors.ValidationFailed("Invalid format. Supported: csv, excel, tableau, powerbi, looker", nil))
		return
	}
	
	data, contentType, err := h.biExportService.ExportQuery(r.Context(), query, bibot.ExportFormat(format))
	if err != nil {
		WriteError(w, err)
		return
	}
	
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=export."+format)
	w.Write(data)
}
