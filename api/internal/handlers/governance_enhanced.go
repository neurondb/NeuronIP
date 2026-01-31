package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/governance"
	"github.com/neurondb/NeuronIP/api/internal/audit"
)

/* GovernanceEnhancedHandler handles governance requests */
type GovernanceEnhancedHandler struct {
	uiRLSService      *governance.UIRLSService
	residencyService  *compliance.ResidencyService
	auditExporter     *audit.ExporterService
}

/* NewGovernanceEnhancedHandler creates a new governance enhanced handler */
func NewGovernanceEnhancedHandler(uiRLSService *governance.UIRLSService, residencyService *compliance.ResidencyService, auditExporter *audit.ExporterService) *GovernanceEnhancedHandler {
	return &GovernanceEnhancedHandler{
		uiRLSService:     uiRLSService,
		residencyService: residencyService,
		auditExporter:    auditExporter,
	}
}

/* CreateRLSPolicy handles POST /api/v1/governance/rls/policies */
func (h *GovernanceEnhancedHandler) CreateRLSPolicy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TableName    string `json:"table_name"`
		SchemaName   string `json:"schema_name"`
		PolicyName   string `json:"policy_name"`
		PolicyType   string `json:"policy_type"`
		Condition    string `json:"condition"`
		Description  string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		WriteErrorResponse(w, errors.Unauthorized("User ID required"))
		return
	}

	policy, err := h.uiRLSService.CreateRLSPolicy(r.Context(), governance.RLSPolicy{
		TableName:  req.TableName,
		SchemaName:  req.SchemaName,
		PolicyName:  req.PolicyName,
		PolicyType:  req.PolicyType,
		Condition:   req.Condition,
		Description: req.Description,
		Enabled:     true,
		CreatedBy:   userID,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(policy)
}

/* GetRLSPolicies handles GET /api/v1/governance/rls/policies */
func (h *GovernanceEnhancedHandler) GetRLSPolicies(w http.ResponseWriter, r *http.Request) {
	schemaName := r.URL.Query().Get("schema_name")
	tableName := r.URL.Query().Get("table_name")

	if schemaName == "" || tableName == "" {
		WriteErrorResponse(w, errors.ValidationFailed("schema_name and table_name are required", nil))
		return
	}

	policies, err := h.uiRLSService.GetRLSPolicies(r.Context(), schemaName, tableName)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policies)
}

/* CreateResidencyRule handles POST /api/v1/governance/residency/rules */
func (h *GovernanceEnhancedHandler) CreateResidencyRule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TableName        string `json:"table_name"`
		SchemaName       string `json:"schema_name"`
		RequiredRegion   string `json:"required_region"`
		EnforcementLevel string `json:"enforcement_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	rule, err := h.residencyService.CreateRule(r.Context(), compliance.DataResidencyRule{
		TableName:        req.TableName,
		SchemaName:       req.SchemaName,
		RequiredRegion:   req.RequiredRegion,
		EnforcementLevel: req.EnforcementLevel,
		Enabled:          true,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

/* ExportAudit handles GET /api/v1/governance/audit/export */
func (h *GovernanceEnhancedHandler) ExportAudit(w http.ResponseWriter, r *http.Request) {
	var req audit.ExportRequest
	req.Format = r.URL.Query().Get("format")
	if req.Format == "" {
		req.Format = "csv"
	}

	err := h.auditExporter.ExportAuditLogs(r.Context(), req, w)
	if err != nil {
		WriteError(w, err)
		return
	}
}
