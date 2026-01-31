package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/onboarding"
)

/* OnboardingHandler handles enterprise onboarding requests */
type OnboardingHandler struct {
	service *onboarding.OnboardingService
}

/* NewOnboardingHandler creates a new onboarding handler */
func NewOnboardingHandler(service *onboarding.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{service: service}
}

/* StartOnboarding handles POST /api/v1/onboarding/start */
func (h *OnboardingHandler) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkspaceID uuid.UUID `json:"workspace_id"`
		WorkflowID  uuid.UUID `json:"workflow_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	progress, err := h.service.StartOnboarding(r.Context(), req.WorkspaceID, req.WorkflowID, userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(progress)
}

/* CompleteStep handles POST /api/v1/onboarding/progress/{id}/complete-step */
func (h *OnboardingHandler) CompleteStep(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	progressID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid progress ID"))
		return
	}

	var req struct {
		StepID string `json:"step_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.CompleteStep(r.Context(), progressID, req.StepID); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "completed",
	})
}

/* GetProgress handles GET /api/v1/onboarding/progress/{id} */
func (h *OnboardingHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	progressID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid progress ID"))
		return
	}

	progress, err := h.service.GetProgress(r.Context(), progressID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}
