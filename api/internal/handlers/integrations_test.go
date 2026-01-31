package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestIntegrationHandler_GetIntegration_InvalidID(t *testing.T) {
	handler := &IntegrationHandler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	rec := httptest.NewRecorder()

	handler.GetIntegration(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestIntegrationHandler_CreateIntegration_InvalidJSON(t *testing.T) {
	handler := &IntegrationHandler{}

	body := bytes.NewBufferString(`{not valid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateIntegration(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestIntegrationHandler_DeleteIntegration_InvalidID(t *testing.T) {
	handler := &IntegrationHandler{}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/integrations/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	rec := httptest.NewRecorder()

	handler.DeleteIntegration(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
