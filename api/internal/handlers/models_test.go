package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestModelHandler_GetModel_InvalidID(t *testing.T) {
	handler := &ModelHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/models/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	rec := httptest.NewRecorder()

	handler.GetModel(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestModelHandler_RegisterModel_InvalidJSON(t *testing.T) {
	handler := &ModelHandler{service: nil}

	body := bytes.NewBufferString(`{broken json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.RegisterModel(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestModelHandler_InferModel_InvalidJSON(t *testing.T) {
	handler := &ModelHandler{service: nil}

	body := bytes.NewBufferString(`{bad}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/models/123/infer", body)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.InferModel(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
