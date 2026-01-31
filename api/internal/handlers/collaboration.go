package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/collaboration"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* CollaborationHandler handles collaboration requests */
type CollaborationHandler struct {
	service *collaboration.CollaborationService
	pool    *pgxpool.Pool
}

/* NewCollaborationHandler creates a new collaboration handler */
func NewCollaborationHandler(pool *pgxpool.Pool) *CollaborationHandler {
	return &CollaborationHandler{
		service: collaboration.NewCollaborationService(pool),
		pool:    pool,
	}
}

/* CreateSharedDashboard handles POST /api/v1/collaboration/dashboards */
func (h *CollaborationHandler) CreateSharedDashboard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name           string                 `json:"name"`
		Description    *string                `json:"description,omitempty"`
		DashboardConfig map[string]interface{} `json:"dashboard_config"`
		WorkspaceID    *uuid.UUID             `json:"workspace_id,omitempty"`
		IsPublic       bool                   `json:"is_public"`
		SharedWith     []string               `json:"shared_with,omitempty"`
		Tags           []string               `json:"tags,omitempty"`
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

	dashboard, err := h.service.CreateSharedDashboard(
		r.Context(), req.Name, req.Description, req.DashboardConfig,
		userID, req.WorkspaceID, req.IsPublic, req.SharedWith, req.Tags,
	)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, dashboard)
}

/* GetSharedDashboards handles GET /api/v1/collaboration/dashboards */
func (h *CollaborationHandler) GetSharedDashboards(w http.ResponseWriter, r *http.Request) {
	var workspaceID *uuid.UUID
	if wsIDStr := r.URL.Query().Get("workspace_id"); wsIDStr != "" {
		if id, err := uuid.Parse(wsIDStr); err == nil {
			workspaceID = &id
		}
	}

	userID := r.Header.Get("X-User-ID")
	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	dashboards, err := h.service.GetSharedDashboards(r.Context(), workspaceID, userIDPtr, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, dashboards)
}

/* AddDashboardComment handles POST /api/v1/collaboration/dashboards/{id}/comments */
func (h *CollaborationHandler) AddDashboardComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid dashboard ID"))
		return
	}

	var req struct {
		CommentText    string    `json:"comment_text"`
		ParentCommentID *uuid.UUID `json:"parent_comment_id,omitempty"`
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

	comment, err := h.service.AddDashboardComment(r.Context(), dashboardID, userID, req.CommentText, req.ParentCommentID)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, comment)
}

/* GetDashboardComments handles GET /api/v1/collaboration/dashboards/{id}/comments */
func (h *CollaborationHandler) GetDashboardComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid dashboard ID"))
		return
	}

	comments, err := h.service.GetDashboardComments(r.Context(), dashboardID)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, comments)
}

/* CreateAnswerCard handles POST /api/v1/collaboration/answer-cards */
func (h *CollaborationHandler) CreateAnswerCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string                 `json:"title"`
		QueryText   string                 `json:"query_text"`
		QueryResult map[string]interface{} `json:"query_result"`
		Explanation *string                `json:"explanation,omitempty"`
		WorkspaceID *uuid.UUID             `json:"workspace_id,omitempty"`
		IsPublic    bool                   `json:"is_public"`
		SharedWith  []string               `json:"shared_with,omitempty"`
		Tags        []string               `json:"tags,omitempty"`
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

	card, err := h.service.CreateAnswerCard(
		r.Context(), req.Title, req.QueryText, req.QueryResult, req.Explanation,
		userID, req.WorkspaceID, req.IsPublic, req.SharedWith, req.Tags,
	)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, card)
}

/* CreateAnnotation handles POST /api/v1/collaboration/annotations */
func (h *CollaborationHandler) CreateAnnotation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ResourceType   string    `json:"resource_type"`
		ResourceID     string    `json:"resource_id"`
		TargetType     string    `json:"target_type"`
		TargetPath     string    `json:"target_path"`
		AnnotationText string    `json:"annotation_text"`
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

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource ID"))
		return
	}

	annotationService := collaboration.NewAnnotationService(h.pool)
	annotation, err := annotationService.CreateAnnotation(r.Context(), collaboration.Annotation{
		ResourceType:   req.ResourceType,
		ResourceID:     resourceID,
		TargetType:     req.TargetType,
		TargetPath:     req.TargetPath,
		AnnotationText: req.AnnotationText,
		AuthorID:       userID,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, annotation)
}

/* GetAnnotations handles GET /api/v1/collaboration/annotations */
func (h *CollaborationHandler) GetAnnotations(w http.ResponseWriter, r *http.Request) {
	resourceType := r.URL.Query().Get("resource_type")
	resourceIDStr := r.URL.Query().Get("resource_id")
	
	if resourceType == "" || resourceIDStr == "" {
		WriteErrorResponse(w, errors.ValidationFailed("resource_type and resource_id are required", nil))
		return
	}

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource ID"))
		return
	}

	annotationService := collaboration.NewAnnotationService(h.pool)
	annotations, err := annotationService.GetAnnotations(r.Context(), resourceType, resourceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(annotations)
}

/* CreateThread handles POST /api/v1/collaboration/threads */
func (h *CollaborationHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ResourceType string                 `json:"resource_type"`
		ResourceID   string                 `json:"resource_id"`
		Title        string                 `json:"title"`
		InitialPost  map[string]interface{} `json:"initial_post"`
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

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource ID"))
		return
	}

	if req.InitialPost == nil {
		WriteErrorResponse(w, errors.ValidationFailed("initial_post is required", nil))
		return
	}
	contentVal, ok := req.InitialPost["content"]
	if !ok || contentVal == nil {
		WriteErrorResponse(w, errors.ValidationFailed("initial_post.content is required", nil))
		return
	}
	content, ok := contentVal.(string)
	if !ok {
		WriteErrorResponse(w, errors.ValidationFailed("initial_post.content must be a string", nil))
		return
	}

	threadService := collaboration.NewThreadService(h.pool)
	thread, err := threadService.CreateThread(r.Context(), collaboration.Thread{
		ResourceType: req.ResourceType,
		ResourceID:   resourceID,
		Title:        req.Title,
		InitialPost: collaboration.ThreadPost{
			AuthorID: userID,
			Content:  content,
		},
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, thread)
}

/* GetThreads handles GET /api/v1/collaboration/threads */
func (h *CollaborationHandler) GetThreads(w http.ResponseWriter, r *http.Request) {
	resourceType := r.URL.Query().Get("resource_type")
	resourceIDStr := r.URL.Query().Get("resource_id")
	
	if resourceType == "" || resourceIDStr == "" {
		WriteErrorResponse(w, errors.ValidationFailed("resource_type and resource_id are required", nil))
		return
	}

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource ID"))
		return
	}

	threadService := collaboration.NewThreadService(h.pool)
	threads, err := threadService.GetThreadsForResource(r.Context(), resourceType, resourceID, 50)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, threads)
}

/* GetThread handles GET /api/v1/collaboration/threads/{id} */
func (h *CollaborationHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid thread ID"))
		return
	}

	threadService := collaboration.NewThreadService(h.pool)
	thread, err := threadService.GetThread(r.Context(), threadID)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, thread)
}

/* AddPost handles POST /api/v1/collaboration/threads/{id}/posts */
func (h *CollaborationHandler) AddPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid thread ID"))
		return
	}

	var req struct {
		Content string `json:"content"`
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

	threadService := collaboration.NewThreadService(h.pool)
	err = threadService.AddPost(r.Context(), collaboration.ThreadPost{
		ThreadID: threadID,
		AuthorID: userID,
		Content:  req.Content,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

/* RecordDecision handles POST /api/v1/collaboration/decisions */
func (h *CollaborationHandler) RecordDecision(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ResourceType string                 `json:"resource_type"`
		ResourceID   string                 `json:"resource_id"`
		DecisionType string                 `json:"decision_type"`
		Decision     string                 `json:"decision"`
		Reasoning    string                 `json:"reasoning"`
		Context      map[string]interface{} `json:"context"`
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

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource ID"))
		return
	}

	decisionService := collaboration.NewDecisionHistoryService(h.pool)
	err = decisionService.RecordDecision(r.Context(), collaboration.Decision{
		ResourceType: req.ResourceType,
		ResourceID:   resourceID,
		DecisionType: req.DecisionType,
		Decision:     req.Decision,
		Reasoning:    req.Reasoning,
		MadeBy:       userID,
		Context:      req.Context,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

/* GetDecisionHistory handles GET /api/v1/collaboration/decisions */
func (h *CollaborationHandler) GetDecisionHistory(w http.ResponseWriter, r *http.Request) {
	resourceType := r.URL.Query().Get("resource_type")
	resourceIDStr := r.URL.Query().Get("resource_id")
	
	if resourceType == "" || resourceIDStr == "" {
		WriteErrorResponse(w, errors.ValidationFailed("resource_type and resource_id are required", nil))
		return
	}

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource ID"))
		return
	}

	decisionService := collaboration.NewDecisionHistoryService(h.pool)
	decisions, err := decisionService.GetDecisionHistory(r.Context(), resourceType, resourceID, 100)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, decisions)
}

/* SaveQuestion handles POST /api/v1/collaboration/saved-questions */
func (h *CollaborationHandler) SaveQuestion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QuestionText string     `json:"question_text"`
		AnswerText   *string    `json:"answer_text,omitempty"`
		Explanation  *string    `json:"explanation,omitempty"`
		QueryUsed    *string    `json:"query_used,omitempty"`
		WorkspaceID   *uuid.UUID `json:"workspace_id,omitempty"`
		IsShared      bool      `json:"is_shared"`
		SharedWith    []string  `json:"shared_with,omitempty"`
		Tags          []string  `json:"tags,omitempty"`
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

	saved, err := h.service.SaveQuestion(
		r.Context(), req.QuestionText, req.AnswerText, req.Explanation, req.QueryUsed,
		userID, req.WorkspaceID, req.IsShared, req.SharedWith, req.Tags,
	)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, saved)
}
