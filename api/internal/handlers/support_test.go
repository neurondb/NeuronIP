package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestSupportHandler_GetTicket_InvalidID(t *testing.T) {
	handler := &SupportHandler{service: nil}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/support/tickets/bad", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	rec := httptest.NewRecorder()

	handler.GetTicket(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestSupportHandler_CreateTicket_InvalidJSON(t *testing.T) {
	handler := &SupportHandler{service: nil}

	body := bytes.NewBufferString(`{broken}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/support/tickets", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateTicket(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestSupportHandler_AddConversation_InvalidJSON(t *testing.T) {
	handler := &SupportHandler{service: nil}

	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/support/tickets/123/conversations", body)
	req = mux.SetURLVars(req, map[string]string{"id": "bad-uuid"})
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.AddConversation(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
