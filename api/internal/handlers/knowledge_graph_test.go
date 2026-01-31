package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestKnowledgeGraphHandler_GetEntity_InvalidID(t *testing.T) {
	handler := &KnowledgeGraphHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge-graph/entities/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	rec := httptest.NewRecorder()

	handler.GetEntity(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestKnowledgeGraphHandler_ExtractEntities_InvalidJSON(t *testing.T) {
	handler := &KnowledgeGraphHandler{service: nil}

	body := bytes.NewBufferString(`{broken}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge-graph/extract", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ExtractEntities(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestKnowledgeGraphHandler_LinkEntities_InvalidJSON(t *testing.T) {
	handler := &KnowledgeGraphHandler{service: nil}

	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge-graph/link", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LinkEntities(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestKnowledgeGraphHandler_CreateGlossaryTerm_InvalidJSON(t *testing.T) {
	handler := &KnowledgeGraphHandler{service: nil}

	body := bytes.NewBufferString(`{bad`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge-graph/glossary", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateGlossaryTerm(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestKnowledgeGraphHandler_GetGlossaryTerm_InvalidID(t *testing.T) {
	handler := &KnowledgeGraphHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge-graph/glossary/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	rec := httptest.NewRecorder()

	handler.GetGlossaryTerm(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
