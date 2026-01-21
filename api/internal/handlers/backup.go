package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/backup"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* BackupHandler handles backup requests */
type BackupHandler struct {
	service *backup.BackupService
}

/* NewBackupHandler creates a new backup handler */
func NewBackupHandler(service *backup.BackupService) *BackupHandler {
	return &BackupHandler{service: service}
}

/* CreateBackup handles POST /api/v1/backups */
func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	backup, err := h.service.CreateFullBackup(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(backup)
}

/* RestoreBackup handles POST /api/v1/backups/{id}/restore */
func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid backup ID"))
		return
	}

	if err := h.service.RestoreBackup(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "restored",
		"backup_id": id,
	})
}

/* ListBackups handles GET /api/v1/backups */
func (h *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	backups, err := h.service.ListBackups(r.Context(), limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backups": backups,
		"count":   len(backups),
	})
}
