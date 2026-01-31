package blocks

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/handlers"
)

/* Handler handles block HTTP requests */
type Handler struct {
	service *Service
}

/* NewHandler creates a new block handler */
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

/* GetBlocks handles GET /api/v1/blocks */
func (h *Handler) GetBlocks(w http.ResponseWriter, r *http.Request) {
	pageIDStr := r.URL.Query().Get("page_id")
	if pageIDStr == "" {
		handlers.WriteErrorResponse(w, errors.BadRequest("page_id is required"))
		return
	}

	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid page_id"))
		return
	}

	blocks, err := h.service.GetBlocks(r.Context(), pageID)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"blocks": blocks,
	})
}

/* CreateBlock handles POST /api/v1/blocks */
func (h *Handler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	var req CreateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	var userID *uuid.UUID
	if uidStr, ok := auth.GetUserIDFromContext(r.Context()); ok {
		if uid, err := uuid.Parse(uidStr); err == nil {
			userID = &uid
		}
	}

	block, err := h.service.CreateBlock(r.Context(), req, userID)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"block": block,
	})
}

/* UpdateBlock handles PATCH /api/v1/blocks/{id} */
func (h *Handler) UpdateBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockID, err := uuid.Parse(vars["id"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid block ID"))
		return
	}

	var req UpdateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	block, err := h.service.UpdateBlock(r.Context(), blockID, req)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"block": block,
	})
}

/* DeleteBlock handles DELETE /api/v1/blocks/{id} */
func (h *Handler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockID, err := uuid.Parse(vars["id"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid block ID"))
		return
	}

	if err := h.service.DeleteBlock(r.Context(), blockID); err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* ReorderBlocks handles POST /api/v1/blocks/reorder */
func (h *Handler) ReorderBlocks(w http.ResponseWriter, r *http.Request) {
	var req ReorderBlocksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.ReorderBlocks(r.Context(), req); err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
