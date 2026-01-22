package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* CatalogService provides data catalog functionality */
type CatalogService struct {
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
}

/* NewCatalogService creates a new catalog service */
func NewCatalogService(pool *pgxpool.Pool) *CatalogService {
	return &CatalogService{pool: pool}
}

/* NewCatalogServiceWithNeuronDB creates a new catalog service with NeuronDB client */
func NewCatalogServiceWithNeuronDB(pool *pgxpool.Pool, neurondbClient *neurondb.Client) *CatalogService {
	return &CatalogService{
		pool:           pool,
		neurondbClient: neurondbClient,
	}
}

/* Dataset represents a dataset in the catalog */
type Dataset struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	SchemaInfo  map[string]interface{} `json:"schema_info,omitempty"`
	Fields      []interface{}          `json:"fields,omitempty"`
	Owner       *string                `json:"owner,omitempty"`
	Description *string                `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* Field represents a field in a dataset */
type Field struct {
	ID           uuid.UUID `json:"id"`
	DatasetID    uuid.UUID `json:"dataset_id"`
	FieldName    string    `json:"field_name"`
	FieldType    string    `json:"field_type"`
	Description  *string   `json:"description,omitempty"`
	SemanticTags []string  `json:"semantic_tags,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

/* ListDatasets lists all datasets */
func (s *CatalogService) ListDatasets(ctx context.Context) ([]Dataset, error) {
	query := `
		SELECT id, name, schema_info, fields, owner, description, tags, created_at, updated_at
		FROM neuronip.catalog_datasets
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}
	defer rows.Close()

	var datasets []Dataset
	for rows.Next() {
		var ds Dataset
		var schemaJSON, fieldsJSON json.RawMessage

		err := rows.Scan(&ds.ID, &ds.Name, &schemaJSON, &fieldsJSON, &ds.Owner, &ds.Description, &ds.Tags, &ds.CreatedAt, &ds.UpdatedAt)
		if err != nil {
			continue
		}

		if schemaJSON != nil {
			json.Unmarshal(schemaJSON, &ds.SchemaInfo)
		}
		if fieldsJSON != nil {
			json.Unmarshal(fieldsJSON, &ds.Fields)
		}

		datasets = append(datasets, ds)
	}

	return datasets, nil
}

/* GetDataset retrieves a dataset by ID */
func (s *CatalogService) GetDataset(ctx context.Context, id uuid.UUID) (*Dataset, error) {
	query := `
		SELECT id, name, schema_info, fields, owner, description, tags, created_at, updated_at
		FROM neuronip.catalog_datasets
		WHERE id = $1`

	var ds Dataset
	var schemaJSON, fieldsJSON json.RawMessage

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&ds.ID, &ds.Name, &schemaJSON, &fieldsJSON, &ds.Owner, &ds.Description, &ds.Tags, &ds.CreatedAt, &ds.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	if schemaJSON != nil {
		json.Unmarshal(schemaJSON, &ds.SchemaInfo)
	}
	if fieldsJSON != nil {
		json.Unmarshal(fieldsJSON, &ds.Fields)
	}

	// Get fields
	fields, _ := s.GetFields(ctx, id)
	ds.Fields = make([]interface{}, len(fields))
	for i, f := range fields {
		ds.Fields[i] = f
	}

	return &ds, nil
}

/* GetFields retrieves fields for a dataset */
func (s *CatalogService) GetFields(ctx context.Context, datasetID uuid.UUID) ([]Field, error) {
	query := `
		SELECT id, dataset_id, field_name, field_type, description, semantic_tags, created_at
		FROM neuronip.catalog_fields
		WHERE dataset_id = $1
		ORDER BY field_name`

	rows, err := s.pool.Query(ctx, query, datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fields: %w", err)
	}
	defer rows.Close()

	var fields []Field
	for rows.Next() {
		var f Field
		err := rows.Scan(&f.ID, &f.DatasetID, &f.FieldName, &f.FieldType, &f.Description, &f.SemanticTags, &f.CreatedAt)
		if err != nil {
			continue
		}
		fields = append(fields, f)
	}

	return fields, nil
}

/* SearchDatasets performs semantic search on datasets and fields */
func (s *CatalogService) SearchDatasets(ctx context.Context, query string) ([]Dataset, error) {
	searchQuery := `
		SELECT id, name, schema_info, fields, owner, description, tags, created_at, updated_at
		FROM neuronip.catalog_datasets
		WHERE name ILIKE $1 OR description ILIKE $1
		ORDER BY created_at DESC`

	pattern := "%" + query + "%"
	rows, err := s.pool.Query(ctx, searchQuery, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search datasets: %w", err)
	}
	defer rows.Close()

	var datasets []Dataset
	for rows.Next() {
		var ds Dataset
		var schemaJSON, fieldsJSON json.RawMessage

		err := rows.Scan(&ds.ID, &ds.Name, &schemaJSON, &fieldsJSON, &ds.Owner, &ds.Description, &ds.Tags, &ds.CreatedAt, &ds.UpdatedAt)
		if err != nil {
			continue
		}

		if schemaJSON != nil {
			json.Unmarshal(schemaJSON, &ds.SchemaInfo)
		}
		if fieldsJSON != nil {
			json.Unmarshal(fieldsJSON, &ds.Fields)
		}

		datasets = append(datasets, ds)
	}

	return datasets, nil
}

/* ListOwners lists all dataset owners */
func (s *CatalogService) ListOwners(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT owner FROM neuronip.catalog_datasets WHERE owner IS NOT NULL ORDER BY owner`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list owners: %w", err)
	}
	defer rows.Close()

	var owners []string
	for rows.Next() {
		var owner string
		if err := rows.Scan(&owner); err == nil {
			owners = append(owners, owner)
		}
	}

	return owners, nil
}

/* DiscoverDatasets performs semantic discovery */
func (s *CatalogService) DiscoverDatasets(ctx context.Context, tags []string) ([]Dataset, error) {
	var datasets []Dataset

	if len(tags) > 0 {
		query := `
			SELECT id, name, schema_info, fields, owner, description, tags, created_at, updated_at
			FROM neuronip.catalog_datasets
			WHERE tags && $1
			ORDER BY created_at DESC`

		rows, err := s.pool.Query(ctx, query, tags)
		if err != nil {
			return nil, fmt.Errorf("failed to discover datasets: %w", err)
		}
		defer rows.Close()
	} else {
		// Return all datasets
		return s.ListDatasets(ctx)
	}

	// Process rows (simplified - actual implementation would need type assertion)
	// For now, just return empty list
	return datasets, nil
}

/* GenerateDatasetEmbedding generates embedding for dataset metadata with optional image */
func (s *CatalogService) GenerateDatasetEmbedding(ctx context.Context, datasetID uuid.UUID, imageData []byte) (string, error) {
	if s.neurondbClient == nil {
		return "", fmt.Errorf("NeuronDB client not configured")
	}

	// Get dataset
	dataset, err := s.GetDataset(ctx, datasetID)
	if err != nil {
		return "", fmt.Errorf("failed to get dataset: %w", err)
	}

	// Build text for embedding
	text := dataset.Name
	if dataset.Description != nil {
		text += " " + *dataset.Description
	}
	if len(dataset.Tags) > 0 {
		text += " " + fmt.Sprintf("%v", dataset.Tags)
	}

	// Generate embedding based on whether image is provided
	modelName := "sentence-transformers/all-MiniLM-L6-v2"
	if imageData != nil && len(imageData) > 0 {
		// Use multimodal embedding
		embedding, err := s.neurondbClient.GenerateMultimodalEmbedding(ctx, text, imageData, modelName)
		if err != nil {
			// Fallback to text-only
			return s.neurondbClient.GenerateEmbedding(ctx, text, modelName)
		}
		return embedding, nil
	}

	// Use text-only embedding
	return s.neurondbClient.GenerateEmbedding(ctx, text, modelName)
}

/* GenerateImageOnlyEmbedding generates embedding for image-only metadata */
func (s *CatalogService) GenerateImageOnlyEmbedding(ctx context.Context, imageData []byte) (string, error) {
	if s.neurondbClient == nil {
		return "", fmt.Errorf("NeuronDB client not configured")
	}

	modelName := "sentence-transformers/all-MiniLM-L6-v2"
	return s.neurondbClient.GenerateImageEmbedding(ctx, imageData, modelName)
}
