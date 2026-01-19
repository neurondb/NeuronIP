package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/datasources"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* DataSourceHandler handles data source requests */
type DataSourceHandler struct {
	service *datasources.DataSourceService
}

/* NewDataSourceHandler creates a new data source handler */
func NewDataSourceHandler(service *datasources.DataSourceService) *DataSourceHandler {
	return &DataSourceHandler{service: service}
}

/* ListDataSources handles GET /api/v1/data-sources */
func (h *DataSourceHandler) ListDataSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.service.ListDataSources(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sources)
}

/* GetDataSource handles GET /api/v1/data-sources/{id} */
func (h *DataSourceHandler) GetDataSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data source ID"))
		return
	}

	source, err := h.service.GetDataSource(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(source)
}

/* CreateDataSource handles POST /api/v1/data-sources */
func (h *DataSourceHandler) CreateDataSource(w http.ResponseWriter, r *http.Request) {
	var req datasources.DataSource
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name is required", nil))
		return
	}

	source, err := h.service.CreateDataSource(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(source)
}

/* UpdateDataSource handles PUT /api/v1/data-sources/{id} */
func (h *DataSourceHandler) UpdateDataSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data source ID"))
		return
	}

	var req datasources.DataSource
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	source, err := h.service.UpdateDataSource(r.Context(), id, req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(source)
}

/* DeleteDataSource handles DELETE /api/v1/data-sources/{id} */
func (h *DataSourceHandler) DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data source ID"))
		return
	}

	if err := h.service.DeleteDataSource(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* TriggerSync handles POST /api/v1/data-sources/{id}/sync */
func (h *DataSourceHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data source ID"))
		return
	}

	if err := h.service.TriggerSync(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "sync_triggered",
		"id":     id,
	})
}

/* GetSyncStatus handles GET /api/v1/data-sources/{id}/status */
func (h *DataSourceHandler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data source ID"))
		return
	}

	status, err := h.service.GetSyncStatus(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
