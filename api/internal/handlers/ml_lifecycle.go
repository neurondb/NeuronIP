package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/ml"
)

/* MLLifecycleHandler handles ML lifecycle requests */
type MLLifecycleHandler struct {
	pipelineService  *ml.MLPipelineService
	experimentService *ml.ExperimentService
	servingService   *ml.ModelServingService
}

/* NewMLLifecycleHandler creates a new ML lifecycle handler */
func NewMLLifecycleHandler(
	pipelineService *ml.MLPipelineService,
	experimentService *ml.ExperimentService,
	servingService *ml.ModelServingService,
) *MLLifecycleHandler {
	return &MLLifecycleHandler{
		pipelineService:  pipelineService,
		experimentService: experimentService,
		servingService:   servingService,
	}
}

/* CreatePipeline handles POST /api/v1/ml/pipelines */
func (h *MLLifecycleHandler) CreatePipeline(w http.ResponseWriter, r *http.Request) {
	var pipeline ml.TrainingPipeline
	if err := json.NewDecoder(r.Body).Decode(&pipeline); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	createdPipeline, err := h.pipelineService.CreatePipeline(r.Context(), pipeline)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPipeline)
}

/* ExecutePipeline handles POST /api/v1/ml/pipelines/{id}/execute */
func (h *MLLifecycleHandler) ExecutePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid pipeline ID"))
		return
	}

	execution, err := h.pipelineService.ExecutePipeline(r.Context(), pipelineID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(execution)
}

/* CreateExperiment handles POST /api/v1/ml/experiments */
func (h *MLLifecycleHandler) CreateExperiment(w http.ResponseWriter, r *http.Request) {
	var experiment ml.Experiment
	if err := json.NewDecoder(r.Body).Decode(&experiment); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	createdExperiment, err := h.experimentService.CreateExperiment(r.Context(), experiment)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdExperiment)
}

/* LogMetric handles POST /api/v1/ml/experiments/{id}/metrics */
func (h *MLLifecycleHandler) LogMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid experiment ID"))
		return
	}

	var req struct {
		MetricName  string  `json:"metric_name"`
		MetricValue float64 `json:"metric_value"`
		Step        int     `json:"step"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.experimentService.LogMetric(r.Context(), experimentID, req.MetricName, req.MetricValue, req.Step); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "logged",
	})
}

/* DeployModel handles POST /api/v1/ml/models/{id}/deploy */
func (h *MLLifecycleHandler) DeployModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}

	var deployment ml.ModelDeployment
	if err := json.NewDecoder(r.Body).Decode(&deployment); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	deployment.ModelID = modelID
	createdDeployment, err := h.servingService.DeployModel(r.Context(), deployment)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdDeployment)
}

/* CreateABTest handles POST /api/v1/ml/ab-tests */
func (h *MLLifecycleHandler) CreateABTest(w http.ResponseWriter, r *http.Request) {
	var abTest ml.ABTest
	if err := json.NewDecoder(r.Body).Decode(&abTest); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	createdABTest, err := h.servingService.CreateABTest(r.Context(), abTest)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdABTest)
}
