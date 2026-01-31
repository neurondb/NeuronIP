package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/streaming"
)

/* StreamingHandler handles streaming pipeline requests */
type StreamingHandler struct {
	service *streaming.StreamingEngine
}

/* NewStreamingHandler creates a new streaming handler */
func NewStreamingHandler(service *streaming.StreamingEngine) *StreamingHandler {
	return &StreamingHandler{service: service}
}

/* CreatePipeline handles POST /api/v1/streaming/pipelines */
func (h *StreamingHandler) CreatePipeline(w http.ResponseWriter, r *http.Request) {
	var pipeline streaming.Pipeline
	if err := json.NewDecoder(r.Body).Decode(&pipeline); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	createdPipeline, err := h.service.CreatePipeline(r.Context(), pipeline)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPipeline)
}

/* StartPipeline handles POST /api/v1/streaming/pipelines/{id}/start */
func (h *StreamingHandler) StartPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid pipeline ID"))
		return
	}

	if err := h.service.StartPipeline(r.Context(), pipelineID); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pipeline_id": pipelineID,
		"status":      "started",
	})
}

/* StopPipeline handles POST /api/v1/streaming/pipelines/{id}/stop */
func (h *StreamingHandler) StopPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid pipeline ID"))
		return
	}

	if err := h.service.StopPipeline(r.Context(), pipelineID); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pipeline_id": pipelineID,
		"status":      "stopped",
	})
}
