package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/semantic"
)

/* PipelineHandler handles pipeline management requests */
type PipelineHandler struct {
	service *semantic.PipelineService
}

/* NewPipelineHandler creates a new pipeline handler */
func NewPipelineHandler(service *semantic.PipelineService) *PipelineHandler {
	return &PipelineHandler{service: service}
}

/* CreatePipeline handles creating a pipeline */
func (h *PipelineHandler) CreatePipeline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string                 `json:"name"`
		ChunkingConfig  semantic.ChunkingConfig `json:"chunking_config"`
		EmbeddingModel  string                 `json:"embedding_model"`
		EmbeddingConfig map[string]interface{} `json:"embedding_config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	pipeline, err := h.service.CreatePipeline(
		r.Context(),
		req.Name,
		req.ChunkingConfig,
		req.EmbeddingModel,
		req.EmbeddingConfig,
	)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pipeline)
}

/* GetPipeline handles retrieving a pipeline */
func (h *PipelineHandler) GetPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid pipeline ID"))
		return
	}

	version := r.URL.Query().Get("version")
	if version == "" {
		version = "latest"
	}

	pipeline, err := h.service.GetPipeline(r.Context(), pipelineID, version)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pipeline)
}

/* ListPipelineVersions handles listing pipeline versions */
func (h *PipelineHandler) ListPipelineVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid pipeline ID"))
		return
	}

	versions, err := h.service.ListPipelineVersions(r.Context(), pipelineID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

/* ReplayPipeline handles replaying documents with a new pipeline */
func (h *PipelineHandler) ReplayPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid pipeline ID"))
		return
	}

	var req struct {
		DocumentIDs []uuid.UUID `json:"document_ids"`
		Version     string      `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.ReplayPipeline(r.Context(), req.DocumentIDs, pipelineID, req.Version); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

/* ActivatePipeline handles activating a pipeline version */
func (h *PipelineHandler) ActivatePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid pipeline ID"))
		return
	}

	version := r.URL.Query().Get("version")
	if version == "" {
		WriteErrorResponse(w, errors.BadRequest("version is required"))
		return
	}

	if err := h.service.ActivatePipeline(r.Context(), pipelineID, version); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
