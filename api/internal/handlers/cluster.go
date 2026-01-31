package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/cluster"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* ClusterHandler handles cluster management requests */
type ClusterHandler struct {
	service *cluster.Service
}

/* NewClusterHandler creates a new cluster handler */
func NewClusterHandler(service *cluster.Service) *ClusterHandler {
	return &ClusterHandler{service: service}
}

/* GetClusterHealth handles GET /api/v1/cluster/health */
func (h *ClusterHandler) GetClusterHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.service.GetClusterHealth(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

/* GetActiveNodes handles GET /api/v1/cluster/nodes */
func (h *ClusterHandler) GetActiveNodes(w http.ResponseWriter, r *http.Request) {
	nodeType := r.URL.Query().Get("type")

	nodes, err := h.service.GetActiveNodes(r.Context(), nodeType)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

/* EnqueueTask handles POST /api/v1/cluster/tasks */
func (h *ClusterHandler) EnqueueTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskType string                 `json:"task_type"`
		Priority int                    `json:"priority"`
		Payload  map[string]interface{} `json:"payload"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	task := cluster.Task{
		TaskType: req.TaskType,
		Priority: req.Priority,
		Payload:  req.Payload,
		Metadata: req.Metadata,
	}

	enqueuedTask, err := h.service.EnqueueTask(r.Context(), task)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(enqueuedTask)
}

/* CreateAutoScalingPolicy handles POST /api/v1/cluster/autoscaling/policies */
func (h *ClusterHandler) CreateAutoScalingPolicy(w http.ResponseWriter, r *http.Request) {
	var policy cluster.AutoScalingPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	createdPolicy, err := h.service.CreateAutoScalingPolicy(r.Context(), policy)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPolicy)
}

/* GetLeader handles GET /api/v1/cluster/leader */
func (h *ClusterHandler) GetLeader(w http.ResponseWriter, r *http.Request) {
	leader := h.service.GetLeader()
	var leaderID string
	if leader != nil {
		leaderID = *leader
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leader_id": leaderID,
		"is_leader": h.service.IsLeader(),
	})
}
