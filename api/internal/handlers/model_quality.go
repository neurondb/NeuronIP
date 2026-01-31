package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/models"
)

/* ModelQualityHandler handles model quality scoring requests */
type ModelQualityHandler struct {
	scorer *models.QualityScorer
}

/* NewModelQualityHandler creates a new model quality handler */
func NewModelQualityHandler(scorer *models.QualityScorer) *ModelQualityHandler {
	return &ModelQualityHandler{scorer: scorer}
}

/* ScoreOutputRequest is the request body for scoring model output */
type ScoreOutputRequest struct {
	ModelID        string                 `json:"model_id"`
	ModelVersion   string                 `json:"model_version"`
	Output         interface{}            `json:"output"`
	ExpectedOutput interface{}            `json:"expected_output,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

/* ScoreOutput handles POST /api/v1/models/quality/score */
func (h *ModelQualityHandler) ScoreOutput(w http.ResponseWriter, r *http.Request) {
	var req ScoreOutputRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	modelID, err := uuid.Parse(req.ModelID)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model_id"))
		return
	}
	if req.ModelVersion == "" {
		req.ModelVersion = "1.0"
	}
	score, err := h.scorer.ScoreOutput(r.Context(), modelID, req.ModelVersion, req.Output, req.ExpectedOutput, req.Metadata)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(score)
}

/* GetScores handles GET /api/v1/models/{id}/quality/scores */
func (h *ModelQualityHandler) GetScores(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}
	scores, err := h.scorer.GetQualityScores(r.Context(), modelID, 100)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

/* GetAverageScore handles GET /api/v1/models/{id}/quality/average */
func (h *ModelQualityHandler) GetAverageScore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}
	avg, err := h.scorer.GetAverageQualityScore(r.Context(), modelID)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"model_id": modelID, "average_score": avg})
}
