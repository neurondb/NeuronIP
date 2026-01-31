package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/catalog"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* GlossaryLinkHandler handles glossary linking requests */
type GlossaryLinkHandler struct {
	linker *catalog.GlossaryLinker
}

/* NewGlossaryLinkHandler creates a new glossary link handler */
func NewGlossaryLinkHandler(linker *catalog.GlossaryLinker) *GlossaryLinkHandler {
	return &GlossaryLinkHandler{linker: linker}
}

/* LinkRequest is the request body for creating a glossary link */
type LinkRequest struct {
	GlossaryID   string  `json:"glossary_id"`
	LinkType     string  `json:"link_type"` // "column", "metric", "document", "table"
	Relationship string  `json:"relationship"`
	Confidence   float64 `json:"confidence"`

	// For column
	SchemaName string `json:"schema_name,omitempty"`
	TableName  string `json:"table_name,omitempty"`
	ColumnName string `json:"column_name,omitempty"`

	// For metric/document (entity_id)
	EntityID string `json:"entity_id,omitempty"`
}

/* LinkGlossary handles POST /api/v1/catalog/glossary/link */
func (h *GlossaryLinkHandler) LinkGlossary(w http.ResponseWriter, r *http.Request) {
	var req LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	glossaryID, err := uuid.Parse(req.GlossaryID)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid glossary_id"))
		return
	}
	if req.Relationship == "" {
		req.Relationship = "defines"
	}
	if req.Confidence <= 0 {
		req.Confidence = 0.8
	}

	ctx := r.Context()
	switch req.LinkType {
	case "column":
		if req.SchemaName == "" || req.TableName == "" || req.ColumnName == "" {
			WriteErrorResponse(w, errors.ValidationFailed("schema_name, table_name, column_name required for column link", nil))
			return
		}
		err = h.linker.LinkGlossaryToColumn(ctx, glossaryID, req.SchemaName, req.TableName, req.ColumnName, req.Relationship, req.Confidence)
	case "table":
		if req.SchemaName == "" || req.TableName == "" {
			WriteErrorResponse(w, errors.ValidationFailed("schema_name, table_name required for table link", nil))
			return
		}
		err = h.linker.LinkGlossaryToTable(ctx, glossaryID, req.SchemaName, req.TableName, req.Relationship, req.Confidence)
	case "metric", "document":
		if req.EntityID == "" {
			WriteErrorResponse(w, errors.ValidationFailed("entity_id required for metric/document link", nil))
			return
		}
		entityID, parseErr := uuid.Parse(req.EntityID)
		if parseErr != nil {
			WriteErrorResponse(w, errors.BadRequest("Invalid entity_id"))
			return
		}
		if req.LinkType == "metric" {
			err = h.linker.LinkGlossaryToMetric(ctx, glossaryID, entityID, req.Relationship, req.Confidence)
		} else {
			err = h.linker.LinkGlossaryToDocument(ctx, glossaryID, entityID, req.Relationship, req.Confidence)
		}
	default:
		WriteErrorResponse(w, errors.ValidationFailed("link_type must be column, table, metric, or document", nil))
		return
	}
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

/* GetGlossaryLinks handles GET /api/v1/catalog/glossary/{id}/links */
func (h *GlossaryLinkHandler) GetGlossaryLinks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	glossaryID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid glossary ID"))
		return
	}
	entityType := r.URL.Query().Get("entity_type")
	var entityTypePtr *string
	if entityType != "" {
		entityTypePtr = &entityType
	}
	links, err := h.linker.GetGlossaryLinks(r.Context(), glossaryID, entityTypePtr)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

/* GetEntityGlossaryLinks handles GET /api/v1/catalog/glossary/entity/{entity_type}/{entity_id}/links */
func (h *GlossaryLinkHandler) GetEntityGlossaryLinks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityType := vars["entity_type"]
	entityID, err := uuid.Parse(vars["entity_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid entity_id"))
		return
	}
	links, err := h.linker.GetEntityLinks(r.Context(), entityType, entityID)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}
