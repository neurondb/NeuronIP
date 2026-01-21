package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/comments"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* CommentsHandler handles comments requests */
type CommentsHandler struct {
	service *comments.Service
}

/* NewCommentsHandler creates a new comments handler */
func NewCommentsHandler(service *comments.Service) *CommentsHandler {
	return &CommentsHandler{service: service}
}

/* CreateComment creates a new comment */
func (h *CommentsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var comment comments.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateComment(r.Context(), comment)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetComment retrieves a comment */
func (h *CommentsHandler) GetComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid comment ID"))
		return
	}

	comment, err := h.service.GetComment(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

/* ListComments lists comments for a resource */
func (h *CommentsHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := vars["resource_type"]
	resourceID, err := uuid.Parse(vars["resource_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource ID"))
		return
	}

	includeResolved := r.URL.Query().Get("include_resolved") == "true"
	commentList, err := h.service.ListComments(r.Context(), resourceType, resourceID, includeResolved)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commentList)
}

/* ResolveComment resolves a comment */
func (h *CommentsHandler) ResolveComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid comment ID"))
		return
	}

	// Get user ID from context (would be set by auth middleware)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	if err := h.service.ResolveComment(r.Context(), id, userID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* DeleteComment deletes a comment */
func (h *CommentsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid comment ID"))
		return
	}

	if err := h.service.DeleteComment(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
