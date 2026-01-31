package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/warehouse"
)

/* DataProductsHandler handles data product (share) requests */
type DataProductsHandler struct {
	service *warehouse.DataProductService
}

/* NewDataProductsHandler creates a new data products handler */
func NewDataProductsHandler(service *warehouse.DataProductService) *DataProductsHandler {
	return &DataProductsHandler{service: service}
}

/* CreateDataProduct handles POST /api/v1/data-products */
func (h *DataProductsHandler) CreateDataProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string      `json:"name"`
		Description       string      `json:"description"`
		OwnerID           string      `json:"owner_id"`
		WorkspaceID       *uuid.UUID  `json:"workspace_id,omitempty"`
		Version           string      `json:"version"`
		SchemaIDs         []uuid.UUID `json:"schema_ids,omitempty"`
		MetricIDs         []uuid.UUID `json:"metric_ids,omitempty"`
		DatasetIDs        []uuid.UUID `json:"dataset_ids,omitempty"`
		SLAFreshnessHours *int        `json:"sla_freshness_hours,omitempty"`
		Visibility        string      `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.Name == "" || req.OwnerID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name and owner_id are required", nil))
		return
	}
	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}
	dp, err := h.service.Create(r.Context(), req.Name, req.Description, req.OwnerID, req.WorkspaceID, req.Version, req.SchemaIDs, req.MetricIDs, req.DatasetIDs, req.SLAFreshnessHours, req.Visibility)
	if err != nil {
		WriteError(w, err)
		return
	}
	_ = desc
	WriteJSON(w, http.StatusCreated, dp)
}

/* GetDataProduct handles GET /api/v1/data-products/{id} */
func (h *DataProductsHandler) GetDataProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data product ID"))
		return
	}
	dp, err := h.service.Get(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, dp)
}

/* ListDataProducts handles GET /api/v1/data-products */
func (h *DataProductsHandler) ListDataProducts(w http.ResponseWriter, r *http.Request) {
	var workspaceID *uuid.UUID
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if id, err := uuid.Parse(wsStr); err == nil {
			workspaceID = &id
		}
	}
	ownerID := r.URL.Query().Get("owner_id")
	list, err := h.service.List(r.Context(), workspaceID, ownerID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

/* ShareDataProduct handles POST /api/v1/data-products/{id}/share */
func (h *DataProductsHandler) ShareDataProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data product ID"))
		return
	}
	var req struct {
		ConsumerWorkspaceID *uuid.UUID `json:"consumer_workspace_id,omitempty"`
		ConsumerUserID      *string    `json:"consumer_user_id,omitempty"`
		Permissions         string     `json:"permissions"`
		GrantedBy           string     `json:"granted_by"`
		ExpiresAt           *time.Time `json:"expires_at,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	var gBy string
	if req.GrantedBy != "" {
		gBy = req.GrantedBy
	}
	c, err := h.service.Share(r.Context(), id, req.ConsumerWorkspaceID, req.ConsumerUserID, req.Permissions, gBy, req.ExpiresAt)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, c)
}

/* RevokeDataProduct handles POST /api/v1/data-products/{id}/revoke */
func (h *DataProductsHandler) RevokeDataProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data product ID"))
		return
	}
	var req struct {
		ConsumerWorkspaceID *uuid.UUID `json:"consumer_workspace_id,omitempty"`
		ConsumerUserID      *string    `json:"consumer_user_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	err = h.service.Revoke(r.Context(), id, req.ConsumerWorkspaceID, req.ConsumerUserID)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

/* ListConsumers handles GET /api/v1/data-products/{id}/consumers */
func (h *DataProductsHandler) ListConsumers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid data product ID"))
		return
	}
	list, err := h.service.ListConsumers(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}
