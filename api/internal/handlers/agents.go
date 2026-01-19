package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/agents"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* AgentsHandler handles agent requests */
type AgentsHandler struct {
	service *agents.AgentsService
}

/* NewAgentsHandler creates a new agents handler */
func NewAgentsHandler(service *agents.AgentsService) *AgentsHandler {
	return &AgentsHandler{service: service}
}

/* ListAgents handles GET /api/v1/agents */
func (h *AgentsHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	agentList, err := h.service.ListAgents(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agentList)
}

/* GetAgent handles GET /api/v1/agents/{id} */
func (h *AgentsHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid agent ID"))
		return
	}

	agent, err := h.service.GetAgent(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

/* CreateAgent handles POST /api/v1/agents */
func (h *AgentsHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req agents.Agent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name is required", nil))
		return
	}

	if req.Status == "" {
		req.Status = "draft"
	}

	agent, err := h.service.CreateAgent(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agent)
}

/* UpdateAgent handles PUT /api/v1/agents/{id} */
func (h *AgentsHandler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid agent ID"))
		return
	}

	var req agents.Agent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	agent, err := h.service.UpdateAgent(r.Context(), id, req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

/* DeleteAgent handles DELETE /api/v1/agents/{id} */
func (h *AgentsHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid agent ID"))
		return
	}

	if err := h.service.DeleteAgent(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* GetPerformance handles GET /api/v1/agents/{id}/performance */
func (h *AgentsHandler) GetPerformance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid agent ID"))
		return
	}

	perf, err := h.service.GetPerformance(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(perf)
}

/* DeployAgent handles POST /api/v1/agents/{id}/deploy */
func (h *AgentsHandler) DeployAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid agent ID"))
		return
	}

	if err := h.service.DeployAgent(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "deployed",
		"id":     id,
	})
}
