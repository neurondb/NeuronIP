package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/support"
)

/* SupportHandler handles support ticket requests */
type SupportHandler struct {
	service *support.Service
}

/* NewSupportHandler creates a new support handler */
func NewSupportHandler(service *support.Service) *SupportHandler {
	return &SupportHandler{service: service}
}

/* CreateTicketRequest represents ticket creation request */
type CreateTicketRequest struct {
	CustomerID    string                 `json:"customer_id"`
	CustomerEmail *string                `json:"customer_email,omitempty"`
	Subject       string                 `json:"subject"`
	Priority      string                 `json:"priority"`
	Message       string                 `json:"message"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* CreateTicket handles ticket creation requests */
func (h *SupportHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.CustomerID == "" || req.Subject == "" {
		WriteErrorResponse(w, errors.ValidationFailed("customer_id and subject are required", nil))
		return
	}

	ticketReq := support.TicketRequest{
		CustomerID:    req.CustomerID,
		CustomerEmail: req.CustomerEmail,
		Subject:       req.Subject,
		Priority:      req.Priority,
		Message:       req.Message,
		Metadata:      req.Metadata,
	}

	ticket, err := h.service.CreateTicket(r.Context(), ticketReq)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

/* GetTicket handles ticket retrieval requests */
func (h *SupportHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid ticket ID"))
		return
	}

	ticket, conversations, err := h.service.GetTicket(r.Context(), ticketID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ticket":       ticket,
		"conversations": conversations,
	})
}

/* ListTickets handles ticket listing requests */
func (h *SupportHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	customerID := r.URL.Query().Get("customer_id")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	tickets, err := h.service.ListTickets(r.Context(), status, customerID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tickets": tickets,
		"count":   len(tickets),
	})
}

/* AddConversationRequest represents conversation addition request */
type AddConversationRequest struct {
	MessageText string                 `json:"message_text"`
	SenderType  string                 `json:"sender_type"`
	SenderID    *string                `json:"sender_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

/* AddConversation handles conversation addition requests */
func (h *SupportHandler) AddConversation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid ticket ID"))
		return
	}

	var req AddConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.MessageText == "" || req.SenderType == "" {
		WriteErrorResponse(w, errors.ValidationFailed("message_text and sender_type are required", nil))
		return
	}

	err = h.service.AddConversation(r.Context(), ticketID, req.MessageText, req.SenderType, req.SenderID, req.Metadata)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Conversation added successfully",
		"ticket_id": ticketID,
	})
}

/* GetConversations handles conversation retrieval requests */
func (h *SupportHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid ticket ID"))
		return
	}

	conversations, err := h.service.GetConversations(r.Context(), ticketID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conversations": conversations,
		"count":         len(conversations),
	})
}

/* GetSimilarCases handles similar cases retrieval requests */
func (h *SupportHandler) GetSimilarCases(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid ticket ID"))
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	similarCases, err := h.service.GetSimilarCases(r.Context(), ticketID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"similar_cases": similarCases,
		"count":         len(similarCases),
	})
}
