package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestAlertsHandler_ResolveAlert_InvalidID(t *testing.T) {
	handler := &AlertsHandler{service: nil}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/invalid/resolve", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-uuid"})
	rec := httptest.NewRecorder()

	handler.ResolveAlert(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAlertsHandler_CreateAlertRule_InvalidJSON(t *testing.T) {
	handler := &AlertsHandler{service: nil}

	body := bytes.NewBufferString(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/rules", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateAlertRule(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestAlertsHandler_UpdateAlertRule_InvalidID(t *testing.T) {
	handler := &AlertsHandler{service: nil}

	body := bytes.NewBufferString(`{"name": "test"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/alerts/rules/not-uuid", body)
	req = mux.SetURLVars(req, map[string]string{"id": "not-uuid"})
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateAlertRule(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestAlertsHandler_DeleteAlertRule_InvalidID(t *testing.T) {
	handler := &AlertsHandler{service: nil}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/rules/not-uuid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "not-uuid"})
	rec := httptest.NewRecorder()

	handler.DeleteAlertRule(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
