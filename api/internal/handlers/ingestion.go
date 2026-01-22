package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
	"strconv"
)

/* IngestionHandler handles ingestion requests */
type IngestionHandler struct {
	service *ingestion.IngestionService
}

/* NewIngestionHandler creates a new ingestion handler */
func NewIngestionHandler(service *ingestion.IngestionService) *IngestionHandler {
	return &IngestionHandler{service: service}
}

/* CreateJob handles POST /api/v1/ingestion/jobs */
func (h *IngestionHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DataSourceID uuid.UUID              `json:"data_source_id"`
		JobType      string                 `json:"job_type"`
		Config       map[string]interface{} `json:"config"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	
	job, err := h.service.CreateIngestionJob(r.Context(), req.DataSourceID, req.JobType, req.Config)
	if err != nil {
		WriteError(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

/* GetJob handles GET /api/v1/ingestion/jobs/{id} */
func (h *IngestionHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid job ID"))
		return
	}
	
	job, err := h.service.GetIngestionJob(r.Context(), jobID)
	if err != nil {
		WriteError(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

/* ListJobs handles GET /api/v1/ingestion/jobs */
func (h *IngestionHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	var dataSourceID *uuid.UUID
	if dsIDStr := r.URL.Query().Get("data_source_id"); dsIDStr != "" {
		if id, err := uuid.Parse(dsIDStr); err == nil {
			dataSourceID = &id
		}
	}
	
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	jobs, err := h.service.ListIngestionJobs(r.Context(), dataSourceID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

/* ExecuteJob handles POST /api/v1/ingestion/jobs/{id}/execute */
func (h *IngestionHandler) ExecuteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid job ID"))
		return
	}
	
	// Execute job asynchronously (in production, would use a job queue)
	go func() {
		h.service.ExecuteSyncJob(r.Context(), jobID)
	}()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": jobID,
		"status": "accepted",
	})
}

/* GetIngestionStatus handles GET /api/v1/ingestion/data-sources/{id}/status */
func (h *IngestionHandler) GetIngestionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataSourceID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data source ID"))
		return
	}

	status, err := h.service.GetIngestionStatus(r.Context(), dataSourceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

/* GetIngestionFailures handles GET /api/v1/ingestion/failures */
func (h *IngestionHandler) GetIngestionFailures(w http.ResponseWriter, r *http.Request) {
	var dataSourceID *uuid.UUID
	if dsIDStr := r.URL.Query().Get("data_source_id"); dsIDStr != "" {
		if id, err := uuid.Parse(dsIDStr); err == nil {
			dataSourceID = &id
		}
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	failures, err := h.service.GetIngestionFailures(r.Context(), dataSourceID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(failures)
}

/* RetryIngestionJob handles POST /api/v1/ingestion/jobs/{id}/retry */
func (h *IngestionHandler) RetryIngestionJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid job ID"))
		return
	}

	err = h.service.RetryIngestionJob(r.Context(), jobID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": jobID,
		"status": "retrying",
	})
}
