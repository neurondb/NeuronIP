package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestAPIKeyHandler_CreateAPIKey_InvalidJSON(t *testing.T) {
	handler := &APIKeyHandler{queries: nil}

	body := bytes.NewBufferString(`{broken json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateAPIKey(rec, req)

	// May return 400 (bad request) or 401 (unauthorized) depending on handler logic
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 400 or 401 for invalid JSON, got %d", rec.Code)
	}
}

func TestAPIKeyHandler_DeleteAPIKey_InvalidID(t *testing.T) {
	handler := &APIKeyHandler{queries: nil}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/api-keys/bad-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-id"})
	rec := httptest.NewRecorder()

	handler.DeleteAPIKey(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
