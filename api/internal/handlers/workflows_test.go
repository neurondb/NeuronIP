package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestWorkflowHandler_GetWorkflow_InvalidID(t *testing.T) {
	handler := &WorkflowHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workflows/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	rec := httptest.NewRecorder()

	handler.GetWorkflow(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestWorkflowHandler_CreateWorkflow_InvalidJSON(t *testing.T) {
	handler := &WorkflowHandler{service: nil}

	body := bytes.NewBufferString(`{not valid}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateWorkflow(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestWorkflowHandler_ExecuteWorkflow_InvalidID(t *testing.T) {
	handler := &WorkflowHandler{service: nil}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bad/execute", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	rec := httptest.NewRecorder()

	handler.ExecuteWorkflow(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestWorkflowHandler_DeleteWorkflow_InvalidID(t *testing.T) {
	handler := &WorkflowHandler{service: nil}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/workflows/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad"})
	rec := httptest.NewRecorder()

	handler.DeleteWorkflow(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
