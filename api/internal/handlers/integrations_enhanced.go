package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	crmint "github.com/neurondb/NeuronIP/api/internal/integrations/crm"
	itsmint "github.com/neurondb/NeuronIP/api/internal/integrations/itsm"
	biint "github.com/neurondb/NeuronIP/api/internal/integrations/bi"
)

/* EnhancedIntegrationHandler handles enhanced integration requests */
type EnhancedIntegrationHandler struct {
	crmService  *crmint.CRMAutomationService
	itsmService *itsmint.ITSMTriggerService
	biService   *biint.BIExporterService
}

/* NewEnhancedIntegrationHandler creates a new enhanced integration handler */
func NewEnhancedIntegrationHandler(crmService *crmint.CRMAutomationService, itsmService *itsmint.ITSMTriggerService, biService *biint.BIExporterService) *EnhancedIntegrationHandler {
	return &EnhancedIntegrationHandler{
		crmService:  crmService,
		itsmService: itsmService,
		biService:   biService,
	}
}

/* CreateCRMHook handles POST /api/v1/integrations/crm/hooks */
func (h *EnhancedIntegrationHandler) CreateCRMHook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CRMType       string                 `json:"crm_type"`
		EventType     string                 `json:"event_type"`
		TriggerConfig map[string]interface{} `json:"trigger_config"`
		ActionConfig  map[string]interface{} `json:"action_config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	hook, err := h.crmService.CreateHook(r.Context(), crmint.AutomationHook{
		CRMType:       req.CRMType,
		EventType:     req.EventType,
		TriggerConfig: req.TriggerConfig,
		ActionConfig:  req.ActionConfig,
		Enabled:       true,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(hook)
}

/* CreateITSMTrigger handles POST /api/v1/integrations/itsm/triggers */
func (h *EnhancedIntegrationHandler) CreateITSMTrigger(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ITSMType    string                 `json:"itsm_type"`
		TriggerType string                 `json:"trigger_type"`
		Config      map[string]interface{} `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	trigger, err := h.itsmService.CreateTrigger(r.Context(), itsmint.Trigger{
		ITSMType:    req.ITSMType,
		TriggerType: req.TriggerType,
		Config:      req.Config,
		Enabled:     true,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(trigger)
}

/* ExportToBI handles GET /api/v1/integrations/bi/export */
func (h *EnhancedIntegrationHandler) ExportToBI(w http.ResponseWriter, r *http.Request) {
	queryIDStr := r.URL.Query().Get("query_id")
	biType := r.URL.Query().Get("bi_type")
	format := r.URL.Query().Get("format")

	if queryIDStr == "" {
		WriteErrorResponse(w, errors.ValidationFailed("query_id is required", nil))
		return
	}

	queryID, err := uuid.Parse(queryIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid query ID"))
		return
	}

	if biType == "" {
		biType = "tableau"
	}
	if format == "" {
		format = "csv"
	}

	data, err := h.biService.ExportQuery(r.Context(), queryID, biType, format)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=export."+format)
	w.Write(data)
}
