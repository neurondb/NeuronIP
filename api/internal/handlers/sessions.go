package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/users"
)

/* SessionHandler handles session management requests */
type SessionHandler struct {
	service *users.SessionService
}

/* NewSessionHandler creates a new session handler */
func NewSessionHandler(service *users.SessionService) *SessionHandler {
	return &SessionHandler{service: service}
}

/* ListSessions handles listing user sessions */
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid user ID"))
		return
	}

	sessions, err := h.service.GetUserSessions(r.Context(), userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

/* RevokeSession handles revoking a session */
func (h *SessionHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid session ID"))
		return
	}

	if err := h.service.RevokeSession(r.Context(), sessionID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* RevokeAllSessions handles revoking all sessions */
func (h *SessionHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid user ID"))
		return
	}

	if err := h.service.RevokeAllSessions(r.Context(), userID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
