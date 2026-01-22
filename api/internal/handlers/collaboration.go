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
}

/* NewCollaborationHandler creates a new collaboration handler */
func NewCollaborationHandler(pool *pgxpool.Pool) *CollaborationHandler {
	return &CollaborationHandler{
		service: collaboration.NewCollaborationService(pool),
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dashboard)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboards)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(saved)
}
