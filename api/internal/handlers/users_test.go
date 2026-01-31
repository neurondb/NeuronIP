package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestUserHandler_GetUserProfile_InvalidID(t *testing.T) {
	handler := &UserHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/bad/profile", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	rec := httptest.NewRecorder()

	handler.GetUserProfile(rec, req)

	// May return 400 (bad request) or 401 (unauthorized) depending on auth check
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 400 or 401, got %d", rec.Code)
	}
}

func TestUserHandler_UpdateUserProfile_InvalidID(t *testing.T) {
	handler := &UserHandler{service: nil}

	body := bytes.NewBufferString(`{"name": "test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/bad/profile", body)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateUserProfile(rec, req)

	// May return 400 (bad request) or 401 (unauthorized) depending on auth check
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 400 or 401, got %d", rec.Code)
	}
}

func TestUserHandler_ChangePassword_InvalidJSON(t *testing.T) {
	handler := &UserHandler{service: nil}

	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/password", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ChangePassword(rec, req)

	// May return 400 (bad request) or 401 (unauthorized) depending on auth check
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 400 or 401, got %d", rec.Code)
	}
}
