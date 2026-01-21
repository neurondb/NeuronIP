package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/profiling"
)

/* OutlierHandler handles outlier detection requests */
type OutlierHandler struct {
	outlierService *profiling.OutlierService
}

/* NewOutlierHandler creates a new outlier handler */
func NewOutlierHandler(outlierService *profiling.OutlierService) *OutlierHandler {
	return &OutlierHandler{outlierService: outlierService}
}

/* DetectOutliers handles POST /api/v1/profiling/outliers/detect */
func (h *OutlierHandler) DetectOutliers(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConnectorID uuid.UUID `json:"connector_id"`
		SchemaName  string    `json:"schema_name"`
		TableName   string    `json:"table_name"`
		ColumnName  string    `json:"column_name"`
		DetectionType string  `json:"detection_type"` // "statistical", "temporal", "pattern", "all"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.DetectionType == "" {
		req.DetectionType = "statistical"
	}

	outliers, err := h.outlierService.DetectOutliers(r.Context(),
		req.ConnectorID, req.SchemaName, req.TableName, req.ColumnName, req.DetectionType)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"outliers": outliers,
		"count":    len(outliers),
	})
}

/* GetOutlierHistory handles GET /api/v1/profiling/outliers/history */
func (h *OutlierHandler) GetOutlierHistory(w http.ResponseWriter, r *http.Request) {
	connectorIDStr := r.URL.Query().Get("connector_id")
	if connectorIDStr == "" {
		WriteErrorResponse(w, errors.BadRequest("connector_id is required"))
		return
	}

	connectorID, err := uuid.Parse(connectorIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid connector ID"))
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	history, err := h.outlierService.GetOutlierHistory(r.Context(), connectorID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
