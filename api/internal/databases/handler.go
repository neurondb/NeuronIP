package databases

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/handlers"
)

/* Handler handles database HTTP requests */
type Handler struct {
	service *Service
}

/* NewHandler creates a new database handler */
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

/* GetDatabase handles GET /api/v1/databases/{id} */
func (h *Handler) GetDatabase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID, err := uuid.Parse(vars["id"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid database ID"))
		return
	}

	db, columns, rows, err := h.service.GetDatabase(r.Context(), databaseID)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"database": db,
		"columns":  columns,
		"rows":     rows,
	})
}

/* CreateDatabase handles POST /api/v1/databases */
func (h *Handler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req CreateDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		handlers.WriteErrorResponse(w, errors.ValidationFailed("name is required", nil))
		return
	}

	var userID *uuid.UUID
	if uidStr, ok := auth.GetUserIDFromContext(r.Context()); ok {
		if uid, err := uuid.Parse(uidStr); err == nil {
			userID = &uid
		}
	}

	db, columns, err := h.service.CreateDatabase(r.Context(), req, userID)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"database": db,
		"columns":  columns,
	})
}

/* UpdateRow handles PATCH /api/v1/databases/{id}/rows/{rowId} */
func (h *Handler) UpdateRow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID, err := uuid.Parse(vars["id"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid database ID"))
		return
	}

	rowID, err := uuid.Parse(vars["rowId"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid row ID"))
		return
	}

	var req UpdateRowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	row, err := h.service.UpdateRow(r.Context(), databaseID, rowID, req)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"row": row,
	})
}

/* CreateRow handles POST /api/v1/databases/{id}/rows */
func (h *Handler) CreateRow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID, err := uuid.Parse(vars["id"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid database ID"))
		return
	}

	var req UpdateRowRequest
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

	row, err := h.service.CreateRow(r.Context(), databaseID, req, userID)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"row": row,
	})
}

/* DeleteRow handles DELETE /api/v1/databases/{id}/rows/{rowId} */
func (h *Handler) DeleteRow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID, err := uuid.Parse(vars["id"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid database ID"))
		return
	}

	rowID, err := uuid.Parse(vars["rowId"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid row ID"))
		return
	}

	if err := h.service.DeleteRow(r.Context(), databaseID, rowID); err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* UpdateViewPreferences handles PATCH /api/v1/databases/{id}/view-preferences */
func (h *Handler) UpdateViewPreferences(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID, err := uuid.Parse(vars["id"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid database ID"))
		return
	}

	var userID uuid.UUID
	if uidStr, ok := auth.GetUserIDFromContext(r.Context()); ok {
		if uid, err := uuid.Parse(uidStr); err == nil {
			userID = uid
		} else {
			handlers.WriteErrorResponse(w, errors.Unauthorized("Invalid user ID"))
			return
		}
	} else {
		handlers.WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	var req ViewPreferences
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.UpdateViewPreferences(r.Context(), databaseID, userID, req); err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* GetViewPreferences handles GET /api/v1/databases/{id}/view-preferences */
func (h *Handler) GetViewPreferences(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID, err := uuid.Parse(vars["id"])
	if err != nil {
		handlers.WriteErrorResponse(w, errors.BadRequest("Invalid database ID"))
		return
	}

	var userID uuid.UUID
	if uidStr, ok := auth.GetUserIDFromContext(r.Context()); ok {
		if uid, err := uuid.Parse(uidStr); err == nil {
			userID = uid
		} else {
			handlers.WriteErrorResponse(w, errors.Unauthorized("Invalid user ID"))
			return
		}
	} else {
		handlers.WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	prefs, err := h.service.GetViewPreferences(r.Context(), databaseID, userID)
	if err != nil {
		handlers.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"preferences": prefs,
	})
}
