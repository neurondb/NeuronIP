package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/knowledgegraph"
)

/* KnowledgeGraphHandler handles knowledge graph requests */
type KnowledgeGraphHandler struct {
	service *knowledgegraph.Service
}

/* NewKnowledgeGraphHandler creates a new knowledge graph handler */
func NewKnowledgeGraphHandler(service *knowledgegraph.Service) *KnowledgeGraphHandler {
	return &KnowledgeGraphHandler{service: service}
}

/* ExtractEntitiesRequest represents entity extraction request */
type ExtractEntitiesRequest struct {
	DocumentID    string   `json:"document_id"`
	Text          string   `json:"text"`
	EntityTypes   []string `json:"entity_types,omitempty"`
	MinConfidence float64  `json:"min_confidence,omitempty"`
}

/* ExtractEntities handles entity extraction requests */
func (h *KnowledgeGraphHandler) ExtractEntities(w http.ResponseWriter, r *http.Request) {
	var req ExtractEntitiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Text == "" {
		WriteErrorResponse(w, errors.ValidationFailed("Text is required", nil))
		return
	}

	docID, err := uuid.Parse(req.DocumentID)
	if err != nil {
		WriteErrorResponse(w, errors.ValidationFailed("Invalid document_id", nil))
		return
	}

	extractReq := knowledgegraph.ExtractEntitiesRequest{
		DocumentID:    docID,
		Text:          req.Text,
		EntityTypes:   req.EntityTypes,
		MinConfidence: req.MinConfidence,
	}

	entities, err := h.service.ExtractEntities(r.Context(), extractReq)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}

/* LinkEntitiesRequest represents entity linking request */
type LinkEntitiesRequest struct {
	SourceEntityID     string   `json:"source_entity_id"`
	TargetEntityID     string   `json:"target_entity_id"`
	RelationshipType   string   `json:"relationship_type"`
	Description        *string  `json:"description,omitempty"`
	RelationshipStrength float64 `json:"relationship_strength,omitempty"`
}

/* LinkEntities handles entity linking requests */
func (h *KnowledgeGraphHandler) LinkEntities(w http.ResponseWriter, r *http.Request) {
	var req LinkEntitiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	sourceID, err := uuid.Parse(req.SourceEntityID)
	if err != nil {
		WriteErrorResponse(w, errors.ValidationFailed("Invalid source_entity_id", nil))
		return
	}

	targetID, err := uuid.Parse(req.TargetEntityID)
	if err != nil {
		WriteErrorResponse(w, errors.ValidationFailed("Invalid target_entity_id", nil))
		return
	}

	if req.RelationshipType == "" {
		WriteErrorResponse(w, errors.ValidationFailed("relationship_type is required", nil))
		return
	}

	link, err := h.service.LinkEntities(r.Context(), sourceID, targetID, req.RelationshipType, req.Description, req.RelationshipStrength)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

/* GetEntity handles entity retrieval requests */
func (h *KnowledgeGraphHandler) GetEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.ValidationFailed("Invalid entity ID", nil))
		return
	}

	entity, err := h.service.GetEntity(r.Context(), entityID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

/* GetEntityLinks handles entity links retrieval requests */
func (h *KnowledgeGraphHandler) GetEntityLinks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.ValidationFailed("Invalid entity ID", nil))
		return
	}

	direction := r.URL.Query().Get("direction") // "incoming", "outgoing", or empty for both

	links, err := h.service.GetEntityLinks(r.Context(), entityID, direction)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

/* SearchEntitiesRequest represents entity search request */
type SearchEntitiesRequest struct {
	Query       string    `json:"query"`
	EntityTypeID *string  `json:"entity_type_id,omitempty"`
	Limit       int       `json:"limit,omitempty"`
}

/* SearchEntities handles entity search requests */
func (h *KnowledgeGraphHandler) SearchEntities(w http.ResponseWriter, r *http.Request) {
	var req SearchEntitiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("Query is required", nil))
		return
	}

	var entityTypeID *uuid.UUID
	if req.EntityTypeID != nil {
		id, err := uuid.Parse(*req.EntityTypeID)
		if err == nil {
			entityTypeID = &id
		}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	entities, err := h.service.SearchEntities(r.Context(), req.Query, entityTypeID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}

/* TraverseGraphRequest represents graph traversal request */
type TraverseGraphRequest struct {
	StartEntityID uuid.UUID `json:"start_entity_id"`
	MaxDepth      int        `json:"max_depth,omitempty"`
	RelationshipTypes []string `json:"relationship_types,omitempty"`
	Direction     string     `json:"direction,omitempty"` // "outgoing", "incoming", "both"
}

/* TraverseGraph handles graph traversal requests */
func (h *KnowledgeGraphHandler) TraverseGraph(w http.ResponseWriter, r *http.Request) {
	var req TraverseGraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.MaxDepth <= 0 {
		req.MaxDepth = 3
	}

	if req.Direction == "" {
		req.Direction = "both"
	}

	result, err := h.service.TraverseGraph(r.Context(), req.StartEntityID, req.MaxDepth, req.RelationshipTypes, req.Direction)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* CreateGlossaryTermRequest represents glossary term creation request */
type CreateGlossaryTermRequest struct {
	Term            string    `json:"term"`
	Definition      string    `json:"definition"`
	Category        *string   `json:"category,omitempty"`
	RelatedEntityID *string  `json:"related_entity_id,omitempty"`
	Synonyms        []string `json:"synonyms,omitempty"`
}

/* CreateGlossaryTerm handles glossary term creation */
func (h *KnowledgeGraphHandler) CreateGlossaryTerm(w http.ResponseWriter, r *http.Request) {
	var req CreateGlossaryTermRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Term == "" || req.Definition == "" {
		WriteErrorResponse(w, errors.ValidationFailed("Term and definition are required", nil))
		return
	}

	var relatedEntityID *uuid.UUID
	if req.RelatedEntityID != nil {
		id, err := uuid.Parse(*req.RelatedEntityID)
		if err == nil {
			relatedEntityID = &id
		}
	}

	term, err := h.service.CreateGlossaryTerm(r.Context(), req.Term, req.Definition, req.Category, relatedEntityID, req.Synonyms)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(term)
}

/* GetGlossaryTerm handles glossary term retrieval */
func (h *KnowledgeGraphHandler) GetGlossaryTerm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	termID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.ValidationFailed("Invalid term ID", nil))
		return
	}

	term, err := h.service.GetGlossaryTerm(r.Context(), termID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(term)
}

/* SearchGlossaryRequest represents glossary search request */
type SearchGlossaryRequest struct {
	Query    string  `json:"query"`
	Category *string `json:"category,omitempty"`
	Limit    int     `json:"limit,omitempty"`
}

/* SearchGlossary handles glossary search requests */
func (h *KnowledgeGraphHandler) SearchGlossary(w http.ResponseWriter, r *http.Request) {
	var req SearchGlossaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("Query is required", nil))
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	terms, err := h.service.SearchGlossary(r.Context(), req.Query, req.Category, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(terms)
}

/* CreateEntityTypeRequest represents entity type creation request */
type CreateEntityTypeRequest struct {
	TypeName     string    `json:"type_name"`
	Description  *string   `json:"description,omitempty"`
	ParentTypeID   *string `json:"parent_type_id,omitempty"`
}

/* CreateEntityType handles entity type creation */
func (h *KnowledgeGraphHandler) CreateEntityType(w http.ResponseWriter, r *http.Request) {
	var req CreateEntityTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.TypeName == "" {
		WriteErrorResponse(w, errors.ValidationFailed("type_name is required", nil))
		return
	}

	var parentTypeID *uuid.UUID
	if req.ParentTypeID != nil {
		id, err := uuid.Parse(*req.ParentTypeID)
		if err == nil {
			parentTypeID = &id
		}
	}

	entityType, err := h.service.CreateEntityType(r.Context(), req.TypeName, req.Description, parentTypeID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entityType)
}

/* ExecuteGraphQueryRequest represents graph query request */
type ExecuteGraphQueryRequest struct {
	Query string `json:"query"`
}

/* ExecuteGraphQuery handles graph query execution */
func (h *KnowledgeGraphHandler) ExecuteGraphQuery(w http.ResponseWriter, r *http.Request) {
	var req ExecuteGraphQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("Query is required", nil))
		return
	}

	result, err := h.service.ExecuteGraphQuery(r.Context(), req.Query)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Convert to frontend-friendly format
	response := map[string]interface{}{
		"nodes": result.Nodes,
		"edges": result.Edges,
	}

	// Transform edges to have source/target as strings
	edges := make([]map[string]interface{}, 0, len(result.Edges))
	for _, edge := range result.Edges {
		edges = append(edges, map[string]interface{}{
			"id":                edge.ID,
			"source":            edge.SourceEntityID.String(),
			"target":            edge.TargetEntityID.String(),
			"relationship_type": edge.RelationshipType,
			"strength":          edge.RelationshipStrength,
		})
	}
	response["edges"] = edges

	// Transform nodes
	nodes := make([]map[string]interface{}, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		nodeMap := map[string]interface{}{
			"id":          node.ID.String(),
			"entity_name": node.EntityName,
		}
		if node.EntityTypeID != nil {
			nodeMap["entity_type_id"] = node.EntityTypeID.String()
		}
		if node.Metadata != nil {
			nodeMap["metadata"] = node.Metadata
		}
		nodes = append(nodes, nodeMap)
	}
	response["nodes"] = nodes

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
