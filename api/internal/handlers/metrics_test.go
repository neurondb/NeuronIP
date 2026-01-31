package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestMetricsHandler_GetMetric_InvalidID(t *testing.T) {
	handler := &MetricsHandler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	rec := httptest.NewRecorder()

	handler.GetMetric(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestMetricsHandler_CreateMetric_InvalidJSON(t *testing.T) {
	handler := &MetricsHandler{}

	body := bytes.NewBufferString(`{not valid json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/metrics", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateMetric(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestMetricsHandler_DeleteMetric_InvalidID(t *testing.T) {
	handler := &MetricsHandler{}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/metrics/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	rec := httptest.NewRecorder()

	handler.DeleteMetric(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
