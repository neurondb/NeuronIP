package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/governance"
)

/* ModelGovernanceHandler handles model and prompt governance requests */
type ModelGovernanceHandler struct {
	modelRegistryService  *governance.ModelRegistryService
	promptTemplateService *governance.PromptTemplateService
	pool                  *pgxpool.Pool
}

/* NewModelGovernanceHandler creates a new model governance handler */
func NewModelGovernanceHandler(pool *pgxpool.Pool) *ModelGovernanceHandler {
	return &ModelGovernanceHandler{
		modelRegistryService:  governance.NewModelRegistryService(pool),
		promptTemplateService: governance.NewPromptTemplateService(pool),
		pool:                  pool,
	}
}

/* ListModels handles GET /api/v1/models */
func (h *ModelGovernanceHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	status := r.URL.Query().Get("status")

	var providerPtr, statusPtr *string
	if provider != "" {
		providerPtr = &provider
	}
	if status != "" {
		statusPtr = &status
	}

	models, err := h.modelRegistryService.ListModels(r.Context(), providerPtr, statusPtr)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

/* GetModel handles GET /api/v1/models/{id} */
func (h *ModelGovernanceHandler) GetModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}

	model, err := h.modelRegistryService.GetModel(r.Context(), modelID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model)
}

/* GetModelVersions handles GET /api/v1/models/{id}/versions */
func (h *ModelGovernanceHandler) GetModelVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}

	model, err := h.modelRegistryService.GetModel(r.Context(), modelID)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Get all versions of this model
	models, err := h.modelRegistryService.ListModels(r.Context(), nil, nil)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Filter by model name
	var versions []governance.Model
	for _, m := range models {
		if m.ModelName == model.ModelName {
			versions = append(versions, m)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

/* ApproveModel handles POST /api/v1/models/{id}/approve */
func (h *ModelGovernanceHandler) ApproveModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}

	var req struct {
		ApproverID string `json:"approver_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	err = h.modelRegistryService.ApproveModel(r.Context(), modelID, req.ApproverID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* RollbackModel handles POST /api/v1/models/{name}/rollback */
func (h *ModelGovernanceHandler) RollbackModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelName := vars["name"]

	var req struct {
		TargetVersion string `json:"target_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.TargetVersion == "" {
		WriteErrorResponse(w, errors.ValidationFailed("target_version is required", nil))
		return
	}

	err := h.modelRegistryService.RollbackModel(r.Context(), modelName, req.TargetVersion)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* ListPrompts handles GET /api/v1/prompts */
func (h *ModelGovernanceHandler) ListPrompts(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	prompts, err := h.promptTemplateService.ListPrompts(r.Context(), limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompts)
}
		if parentID.Valid {
			pid, _ := uuid.Parse(parentID.String)
			prompt.ParentTemplateID = &pid
		}
		if variablesJSON != nil {
			json.Unmarshal(variablesJSON, &prompt.Variables)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &prompt.Metadata)
		}
		
		prompts = append(prompts, prompt)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompts)
}

/* GetPrompt handles GET /api/v1/prompts/{id} */
func (h *ModelGovernanceHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	promptID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid prompt ID"))
		return
	}

	prompt, err := h.promptTemplateService.GetPromptTemplate(r.Context(), promptID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}

/* GetPromptVersions handles GET /api/v1/prompts/{name}/versions */
func (h *ModelGovernanceHandler) GetPromptVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	promptName := vars["name"]

	versions, err := h.promptTemplateService.GetPromptVersions(r.Context(), promptName)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

/* ApprovePrompt handles POST /api/v1/prompts/{id}/approve */
func (h *ModelGovernanceHandler) ApprovePrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	promptID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid prompt ID"))
		return
	}

	var req struct {
		ApproverID string `json:"approver_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	err = h.promptTemplateService.ApprovePrompt(r.Context(), promptID, req.ApproverID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* RollbackPrompt handles POST /api/v1/prompts/{name}/rollback */
func (h *ModelGovernanceHandler) RollbackPrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	promptName := vars["name"]

	var req struct {
		TargetVersion string `json:"target_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.TargetVersion == "" {
		WriteErrorResponse(w, errors.ValidationFailed("target_version is required", nil))
		return
	}

	err := h.promptTemplateService.RollbackPrompt(r.Context(), promptName, req.TargetVersion)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
