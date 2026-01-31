package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestDataSourceHandler_GetDataSource_InvalidID(t *testing.T) {
	handler := &DataSourceHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data-sources/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-uuid"})
	rec := httptest.NewRecorder()

	handler.GetDataSource(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestDataSourceHandler_CreateDataSource_InvalidJSON(t *testing.T) {
	handler := &DataSourceHandler{service: nil}

	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/data-sources", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateDataSource(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestDataSourceHandler_DeleteDataSource_InvalidID(t *testing.T) {
	handler := &DataSourceHandler{service: nil}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/data-sources/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	rec := httptest.NewRecorder()

	handler.DeleteDataSource(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestDataSourceHandler_TriggerSync_InvalidID(t *testing.T) {
	handler := &DataSourceHandler{service: nil}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/data-sources/bad/sync", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	rec := httptest.NewRecorder()

	handler.TriggerSync(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
