package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestComplianceHandler_GetPolicy_InvalidID(t *testing.T) {
	handler := &ComplianceHandler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/compliance/policies/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-uuid"})
	rec := httptest.NewRecorder()

	handler.GetPolicy(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestComplianceHandler_CreatePolicy_InvalidJSON(t *testing.T) {
	handler := &ComplianceHandler{}

	body := bytes.NewBufferString(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/compliance/policies", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreatePolicy(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestComplianceHandler_CheckCompliance_InvalidJSON(t *testing.T) {
	handler := &ComplianceHandler{}

	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/compliance/check", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CheckCompliance(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
