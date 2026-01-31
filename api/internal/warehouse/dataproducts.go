package warehouse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* DataProduct represents a published data product (share) */
type DataProduct struct {
	ID                uuid.UUID   `json:"id"`
	Name              string      `json:"name"`
	Description       *string     `json:"description,omitempty"`
	OwnerID           string      `json:"owner_id"`
	WorkspaceID       *uuid.UUID  `json:"workspace_id,omitempty"`
	Version           string      `json:"version"`
	SchemaIDs         []uuid.UUID `json:"schema_ids,omitempty"`
	MetricIDs         []uuid.UUID `json:"metric_ids,omitempty"`
	DatasetIDs        []uuid.UUID `json:"dataset_ids,omitempty"`
	SLAFreshnessHours *int        `json:"sla_freshness_hours,omitempty"`
	Visibility        string      `json:"visibility"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

/* DataProductConsumer represents a consumer grant for a data product */
type DataProductConsumer struct {
	ID                  uuid.UUID  `json:"id"`
	DataProductID       uuid.UUID  `json:"data_product_id"`
	ConsumerWorkspaceID *uuid.UUID `json:"consumer_workspace_id,omitempty"`
	ConsumerUserID      *string    `json:"consumer_user_id,omitempty"`
	Permissions         string     `json:"permissions"`
	GrantedAt           time.Time  `json:"granted_at"`
	GrantedBy           *string    `json:"granted_by,omitempty"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
}

/* DataProductService provides data product (share) management */
type DataProductService struct {
	pool *pgxpool.Pool
}

/* NewDataProductService creates a new data product service */
func NewDataProductService(pool *pgxpool.Pool) *DataProductService {
	return &DataProductService{pool: pool}
}

/* Create creates a data product */
func (s *DataProductService) Create(ctx context.Context, name, description, ownerID string, workspaceID *uuid.UUID, version string, schemaIDs, metricIDs, datasetIDs []uuid.UUID, slaFreshnessHours *int, visibility string) (*DataProduct, error) {
	if visibility == "" {
		visibility = "private"
	}
	if version == "" {
		version = "1.0.0"
	}
	id := uuid.New()
	now := time.Now()
	schemaJSON, _ := json.Marshal(schemaIDs)
	metricJSON, _ := json.Marshal(metricIDs)
	datasetJSON, _ := json.Marshal(datasetIDs)

	var descVal interface{}
	if description != "" {
		descVal = description
	}
	query := `
		INSERT INTO neuronip.data_products (id, name, description, owner_id, workspace_id, version, schema_ids, metric_ids, dataset_ids, sla_freshness_hours, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, name, description, owner_id, workspace_id, version, schema_ids, metric_ids, dataset_ids, sla_freshness_hours, visibility, created_at, updated_at
	`
	var dp DataProduct
	var descScanned *string
	var wsID *uuid.UUID
	var schemaRaw, metricRaw, datasetRaw []byte
	err := s.pool.QueryRow(ctx, query, id, name, descVal, ownerID, workspaceID, version, schemaJSON, metricJSON, datasetJSON, slaFreshnessHours, visibility, now, now).Scan(
		&dp.ID, &dp.Name, &descScanned, &dp.OwnerID, &wsID, &dp.Version, &schemaRaw, &metricRaw, &datasetRaw, &dp.SLAFreshnessHours, &dp.Visibility, &dp.CreatedAt, &dp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create data product: %w", err)
	}
	dp.Description = descScanned
	dp.WorkspaceID = wsID
	json.Unmarshal(schemaRaw, &dp.SchemaIDs)
	json.Unmarshal(metricRaw, &dp.MetricIDs)
	json.Unmarshal(datasetRaw, &dp.DatasetIDs)
	return &dp, nil
}

/* Get returns a data product by ID */
func (s *DataProductService) Get(ctx context.Context, id uuid.UUID) (*DataProduct, error) {
	query := `
		SELECT id, name, description, owner_id, workspace_id, version, schema_ids, metric_ids, dataset_ids, sla_freshness_hours, visibility, created_at, updated_at
		FROM neuronip.data_products WHERE id = $1
	`
	var dp DataProduct
	var descVal *string
	var wsID *uuid.UUID
	var schemaRaw, metricRaw, datasetRaw []byte
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&dp.ID, &dp.Name, &descVal, &dp.OwnerID, &wsID, &dp.Version, &schemaRaw, &metricRaw, &datasetRaw, &dp.SLAFreshnessHours, &dp.Visibility, &dp.CreatedAt, &dp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get data product: %w", err)
	}
	dp.Description = descVal
	dp.WorkspaceID = wsID
	json.Unmarshal(schemaRaw, &dp.SchemaIDs)
	json.Unmarshal(metricRaw, &dp.MetricIDs)
	json.Unmarshal(datasetRaw, &dp.DatasetIDs)
	return &dp, nil
}

/* List lists data products for a workspace or owner */
func (s *DataProductService) List(ctx context.Context, workspaceID *uuid.UUID, ownerID string) ([]DataProduct, error) {
	query := `
		SELECT id, name, description, owner_id, workspace_id, version, schema_ids, metric_ids, dataset_ids, sla_freshness_hours, visibility, created_at, updated_at
		FROM neuronip.data_products WHERE 1=1
	`
	args := []interface{}{}
	n := 1
	if workspaceID != nil {
		query += fmt.Sprintf(" AND (workspace_id = $%d OR visibility = 'public')", n)
		args = append(args, workspaceID)
		n++
	}
	if ownerID != "" {
		query += fmt.Sprintf(" AND owner_id = $%d", n)
		args = append(args, ownerID)
		n++
	}
	query += " ORDER BY name, version DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DataProduct
	for rows.Next() {
		var dp DataProduct
		var descVal *string
		var wsID *uuid.UUID
		var schemaRaw, metricRaw, datasetRaw []byte
		if err := rows.Scan(&dp.ID, &dp.Name, &descVal, &dp.OwnerID, &wsID, &dp.Version, &schemaRaw, &metricRaw, &datasetRaw, &dp.SLAFreshnessHours, &dp.Visibility, &dp.CreatedAt, &dp.UpdatedAt); err != nil {
			return nil, err
		}
		dp.Description = descVal
		dp.WorkspaceID = wsID
		json.Unmarshal(schemaRaw, &dp.SchemaIDs)
		json.Unmarshal(metricRaw, &dp.MetricIDs)
		json.Unmarshal(datasetRaw, &dp.DatasetIDs)
		list = append(list, dp)
	}
	return list, rows.Err()
}

/* Share grants a consumer access to a data product */
func (s *DataProductService) Share(ctx context.Context, dataProductID uuid.UUID, consumerWorkspaceID *uuid.UUID, consumerUserID *string, permissions, grantedBy string, expiresAt *time.Time) (*DataProductConsumer, error) {
	if permissions == "" {
		permissions = "read"
	}
	id := uuid.New()
	query := `
		INSERT INTO neuronip.data_product_consumers (id, data_product_id, consumer_workspace_id, consumer_user_id, permissions, granted_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, data_product_id, consumer_workspace_id, consumer_user_id, permissions, granted_at, granted_by, expires_at
	`
	var c DataProductConsumer
	var gBy *string
	err := s.pool.QueryRow(ctx, query, id, dataProductID, consumerWorkspaceID, consumerUserID, permissions, grantedBy, expiresAt).Scan(
		&c.ID, &c.DataProductID, &c.ConsumerWorkspaceID, &c.ConsumerUserID, &c.Permissions, &c.GrantedAt, &gBy, &c.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("share data product: %w", err)
	}
	c.GrantedBy = gBy
	return &c, nil
}

/* Revoke removes a consumer's access */
func (s *DataProductService) Revoke(ctx context.Context, dataProductID uuid.UUID, consumerWorkspaceID *uuid.UUID, consumerUserID *string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM neuronip.data_product_consumers
		WHERE data_product_id = $1 AND (consumer_workspace_id IS NOT DISTINCT FROM $2) AND (consumer_user_id IS NOT DISTINCT FROM $3)
	`, dataProductID, consumerWorkspaceID, consumerUserID)
	return err
}

/* ListConsumers lists consumers of a data product */
func (s *DataProductService) ListConsumers(ctx context.Context, dataProductID uuid.UUID) ([]DataProductConsumer, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, data_product_id, consumer_workspace_id, consumer_user_id, permissions, granted_at, granted_by, expires_at
		FROM neuronip.data_product_consumers WHERE data_product_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`, dataProductID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []DataProductConsumer
	for rows.Next() {
		var c DataProductConsumer
		var gBy *string
		if err := rows.Scan(&c.ID, &c.DataProductID, &c.ConsumerWorkspaceID, &c.ConsumerUserID, &c.Permissions, &c.GrantedAt, &gBy, &c.ExpiresAt); err != nil {
			return nil, err
		}
		c.GrantedBy = gBy
		list = append(list, c)
	}
	return list, rows.Err()
}
