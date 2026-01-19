package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/users"
)

/* UserHandler handles user management requests */
type UserHandler struct {
	service *users.Service
}

/* NewUserHandler creates a new user handler */
func NewUserHandler(service *users.Service) *UserHandler {
	return &UserHandler{service: service}
}

/* GetCurrentUser handles getting current user */
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// User ID would be extracted from context in middleware
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

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

/* UpdateCurrentUser handles updating current user */
func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Name      *string `json:"name"`
		AvatarURL *string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.UpdateUser(r.Context(), userID, req.Name, req.AvatarURL, nil); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* GetUserProfile handles getting user profile */
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
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

	profile, err := h.service.GetUserProfile(r.Context(), userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

/* UpdateUserProfile handles updating user profile */
func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
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

	var profile struct {
		Bio      *string                `json:"bio"`
		Company  *string                `json:"company"`
		JobTitle *string                `json:"job_title"`
		Location *string                `json:"location"`
		Website  *string                `json:"website"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	userProfile := &db.UserProfile{
		UserID:   userID,
		Bio:      profile.Bio,
		Company:  profile.Company,
		JobTitle: profile.JobTitle,
		Location: profile.Location,
		Website:  profile.Website,
		Metadata: profile.Metadata,
	}

	if err := h.service.UpdateUserProfile(r.Context(), userID, userProfile); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* GetUserPreferences handles getting user preferences */
func (h *UserHandler) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
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

	preferences, err := h.service.GetUserPreferences(r.Context(), userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"preferences": preferences,
	})
}

/* UpdateUserPreferences handles updating user preferences */
func (h *UserHandler) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Preferences map[string]interface{} `json:"preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.UpdateUserPreferences(r.Context(), userID, req.Preferences); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* ChangePassword handles password change */
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
