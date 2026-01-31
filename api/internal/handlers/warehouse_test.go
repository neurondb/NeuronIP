package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestWarehouseHandler_GetQuery_InvalidID(t *testing.T) {
	handler := &WarehouseHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouse/queries/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	rec := httptest.NewRecorder()

	handler.GetQuery(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestWarehouseHandler_Query_InvalidJSON(t *testing.T) {
	handler := &WarehouseHandler{service: nil}

	body := bytes.NewBufferString(`{broken}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/warehouse/query", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestWarehouseHandler_Query_MissingQuery(t *testing.T) {
	handler := &WarehouseHandler{service: nil}

	body := bytes.NewBufferString(`{"query": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/warehouse/query", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Query(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing query, got %d", rec.Code)
	}
}

func TestWarehouseHandler_CreateSchema_InvalidJSON(t *testing.T) {
	handler := &WarehouseHandler{service: nil}

	body := bytes.NewBufferString(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/warehouse/schemas", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateSchema(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestWarehouseHandler_GetSchema_InvalidID(t *testing.T) {
	handler := &WarehouseHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/warehouse/schemas/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	rec := httptest.NewRecorder()

	handler.GetSchema(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
