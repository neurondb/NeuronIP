package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/models"
)

/* ModelHandler handles model management requests */
type ModelHandler struct {
	service *models.Service
}

/* NewModelHandler creates a new model handler */
func NewModelHandler(service *models.Service) *ModelHandler {
	return &ModelHandler{service: service}
}

/* RegisterModel handles model registration requests */
func (h *ModelHandler) RegisterModel(w http.ResponseWriter, r *http.Request) {
	var model models.Model
	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if model.Name == "" || model.ModelType == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name and model_type are required", nil))
		return
	}

	registeredModel, err := h.service.RegisterModel(r.Context(), model)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(registeredModel)
}

/* GetModel handles model retrieval requests */
func (h *ModelHandler) GetModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}

	model, err := h.service.GetModel(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model)
}

/* InferModel handles model inference requests */
func (h *ModelHandler) InferModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	result, err := h.service.InferModel(r.Context(), modelID, req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": result,
	})
}
