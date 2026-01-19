package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/users"
)

/* NotificationHandler handles notification requests */
type NotificationHandler struct {
	service *users.NotificationService
}

/* NewNotificationHandler creates a new notification handler */
func NewNotificationHandler(service *users.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

/* ListNotifications handles listing notifications */
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
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

	limitStr := r.URL.Query().Get("limit")
	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	notifications, err := h.service.GetUserNotifications(r.Context(), userID, limit, unreadOnly)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

/* MarkNotificationRead handles marking a notification as read */
func (h *NotificationHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid notification ID"))
		return
	}

	if err := h.service.MarkNotificationRead(r.Context(), notificationID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* MarkAllNotificationsRead handles marking all notifications as read */
func (h *NotificationHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.MarkAllNotificationsRead(r.Context(), userID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
