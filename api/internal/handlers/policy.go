package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/policy"
)

/* PolicyHandler handles policy requests */
type PolicyHandler struct {
	engine *policy.PolicyEngine
}

/* NewPolicyHandler creates a new policy handler */
func NewPolicyHandler(engine *policy.PolicyEngine) *PolicyHandler {
	return &PolicyHandler{engine: engine}
}

/* CreatePolicy handles POST /api/v1/policies */
func (h *PolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var p policy.Policy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	createdPolicy, err := h.engine.CreatePolicy(r.Context(), p)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPolicy)
}

/* GetPolicy handles GET /api/v1/policies/{id} */
func (h *PolicyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid policy ID"))
		return
	}

	p, err := h.engine.GetPolicy(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

/* EvaluatePolicy handles POST /api/v1/policies/{id}/evaluate */
func (h *PolicyHandler) EvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid policy ID"))
		return
	}

	var req policy.PolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	result, err := h.engine.EvaluatePolicy(r.Context(), id, req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
