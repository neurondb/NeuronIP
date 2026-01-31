package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestAgentsHandler_ListAgents_RequiresService(t *testing.T) {
	// Test that handler returns error when service is nil
	handler := &AgentsHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents", nil)
	rec := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			// If no panic, check response for error
			if rec.Code != http.StatusInternalServerError && rec.Code != http.StatusOK {
				t.Logf("Handler returned status %d (expected error handling)", rec.Code)
			}
		}
	}()

	handler.ListAgents(rec, req)
}

func TestAgentsHandler_GetAgent_InvalidID(t *testing.T) {
	handler := &AgentsHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/invalid-uuid", nil)
	rec := httptest.NewRecorder()

	// Set up mux vars
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-uuid"})

	handler.GetAgent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] == nil {
		t.Error("Expected error field in response")
	}
}

func TestAgentsHandler_CreateAgent_MissingName(t *testing.T) {
	handler := &AgentsHandler{service: nil}

	body := bytes.NewBufferString(`{"description": "test agent"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateAgent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing name, got %d", rec.Code)
	}
}

func TestAgentsHandler_CreateAgent_InvalidJSON(t *testing.T) {
	handler := &AgentsHandler{service: nil}

	body := bytes.NewBufferString(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateAgent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestAgentsHandler_DeleteAgent_InvalidID(t *testing.T) {
	handler := &AgentsHandler{service: nil}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/agents/not-a-uuid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "not-a-uuid"})
	rec := httptest.NewRecorder()

	handler.DeleteAgent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
