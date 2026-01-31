package collaboration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* AnnotationService provides inline annotation functionality */
type AnnotationService struct {
	pool *pgxpool.Pool
}

/* NewAnnotationService creates a new annotation service */
func NewAnnotationService(pool *pgxpool.Pool) *AnnotationService {
	return &AnnotationService{pool: pool}
}

/* Annotation represents an inline annotation */
type Annotation struct {
	ID            uuid.UUID              `json:"id"`
	ResourceType  string                 `json:"resource_type"` // "query_result", "dashboard", "document", "metric"
	ResourceID    uuid.UUID              `json:"resource_id"`
	TargetType    string                 `json:"target_type"` // "cell", "row", "column", "section"
	TargetPath    string                 `json:"target_path"` // JSON path to target element
	AnnotationText string                `json:"annotation_text"`
	AuthorID      string                 `json:"author_id"`
	Tags          []string               `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* CreateAnnotation creates a new annotation */
func (as *AnnotationService) CreateAnnotation(ctx context.Context, annotation Annotation) (*Annotation, error) {
	annotation.ID = uuid.New()
	annotation.CreatedAt = time.Now()
	annotation.UpdatedAt = time.Now()
	
	tagsJSON, _ := json.Marshal(annotation.Tags)
	metadataJSON, _ := json.Marshal(annotation.Metadata)
	
	query := `
		INSERT INTO neuronip.annotations 
		(id, resource_type, resource_id, target_type, target_path, annotation_text, author_id, tags, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	
	err := as.pool.QueryRow(ctx, query,
		annotation.ID, annotation.ResourceType, annotation.ResourceID, annotation.TargetType,
		annotation.TargetPath, annotation.AnnotationText, annotation.AuthorID, tagsJSON,
		metadataJSON, annotation.CreatedAt, annotation.UpdatedAt,
	).Scan(&annotation.ID, &annotation.CreatedAt, &annotation.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create annotation: %w", err)
	}
	
	return &annotation, nil
}

/* GetAnnotations retrieves annotations for a resource */
func (as *AnnotationService) GetAnnotations(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]Annotation, error) {
	query := `
		SELECT id, resource_type, resource_id, target_type, target_path, annotation_text, author_id, tags, metadata, created_at, updated_at
		FROM neuronip.annotations
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY created_at DESC
	`
	
	rows, err := as.pool.Query(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotations: %w", err)
	}
	defer rows.Close()
	
	var annotations []Annotation
	for rows.Next() {
		var annotation Annotation
		var tagsJSON, metadataJSON json.RawMessage
		
		err := rows.Scan(
			&annotation.ID, &annotation.ResourceType, &annotation.ResourceID, &annotation.TargetType,
			&annotation.TargetPath, &annotation.AnnotationText, &annotation.AuthorID, &tagsJSON,
			&metadataJSON, &annotation.CreatedAt, &annotation.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		if tagsJSON != nil {
			json.Unmarshal(tagsJSON, &annotation.Tags)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &annotation.Metadata)
		}
		
		annotations = append(annotations, annotation)
	}
	
	return annotations, nil
}

/* UpdateAnnotation updates an annotation */
func (as *AnnotationService) UpdateAnnotation(ctx context.Context, annotationID uuid.UUID, annotationText string, tags []string) error {
	tagsJSON, _ := json.Marshal(tags)
	now := time.Now()
	
	query := `
		UPDATE neuronip.annotations
		SET annotation_text = $1, tags = $2, updated_at = $3
		WHERE id = $4
	`
	
	_, err := as.pool.Exec(ctx, query, annotationText, tagsJSON, now, annotationID)
	return err
}

/* DeleteAnnotation deletes an annotation */
func (as *AnnotationService) DeleteAnnotation(ctx context.Context, annotationID uuid.UUID) error {
	query := `DELETE FROM neuronip.annotations WHERE id = $1`
	_, err := as.pool.Exec(ctx, query, annotationID)
	return err
}
