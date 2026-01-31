package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestNotificationHandler_MarkNotificationRead_InvalidID(t *testing.T) {
	handler := &NotificationHandler{service: nil}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/bad/read", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	rec := httptest.NewRecorder()

	handler.MarkNotificationRead(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
