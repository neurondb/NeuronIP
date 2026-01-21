package compliance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* DSARRequest represents a Data Subject Access Request */
type DSARRequest struct {
	ID              uuid.UUID              `json:"id"`
	RequestType     string                 `json:"request_type"` // "access", "deletion", "correction", "portability"
	SubjectName     string                 `json:"subject_name"`
	SubjectEmail    string                 `json:"subject_email"`
	SubjectID       *string                `json:"subject_id,omitempty"`
	Status          string                 `json:"status"` // "pending", "in_progress", "completed", "rejected"
	RequestedAt     time.Time              `json:"requested_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	RequestDetails  map[string]interface{} `json:"request_details,omitempty"`
	DiscoveredData  []DiscoveredDataItem   `json:"discovered_data,omitempty"`
	ResponseData    map[string]interface{} `json:"response_data,omitempty"`
	AssignedTo      *string                `json:"assigned_to,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

/* DiscoveredDataItem represents discovered personal data */
type DiscoveredDataItem struct {
	Source       string                 `json:"source"`
	Location     string                 `json:"location"`
	DataType     string                 `json:"data_type"`
	DataSample   string                 `json:"data_sample,omitempty"`
	Identifiers  map[string]interface{} `json:"identifiers,omitempty"`
	DiscoveredAt time.Time              `json:"discovered_at"`
}

/* DSARService provides DSAR automation functionality */
type DSARService struct {
	pool *pgxpool.Pool
}

/* NewDSARService creates a new DSAR service */
func NewDSARService(pool *pgxpool.Pool) *DSARService {
	return &DSARService{
		pool: pool,
	}
}

/* CreateDSARRequest creates a new DSAR request */
func (s *DSARService) CreateDSARRequest(ctx context.Context, req DSARRequest) (*DSARRequest, error) {
	req.ID = uuid.New()
	req.Status = "pending"
	req.RequestedAt = time.Now()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	requestDetailsJSON, _ := json.Marshal(req.RequestDetails)
	metadataJSON, _ := json.Marshal(req.Metadata)

	var subjectID sql.NullString
	if req.SubjectID != nil {
		subjectID = sql.NullString{String: *req.SubjectID, Valid: true}
	}

	var assignedTo sql.NullString
	if req.AssignedTo != nil {
		assignedTo = sql.NullString{String: *req.AssignedTo, Valid: true}
	}

	query := `
		INSERT INTO neuronip.dsar_requests
		(id, request_type, subject_name, subject_email, subject_id, status, requested_at,
		 request_details, assigned_to, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		req.ID, req.RequestType, req.SubjectName, req.SubjectEmail, subjectID,
		req.Status, req.RequestedAt, requestDetailsJSON, assignedTo, metadataJSON,
		req.CreatedAt, req.UpdatedAt,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create DSAR request: %w", err)
	}

	// Automatically discover data if subject identifier provided
	if req.SubjectID != nil {
		go s.discoverDataForSubject(context.Background(), req.ID, *req.SubjectID)
	}

	return &req, nil
}

/* discoverDataForSubject discovers personal data for a subject */
func (s *DSARService) discoverDataForSubject(ctx context.Context, requestID uuid.UUID, subjectID string) {
	// Search across all data sources for PII matching this subject
	// This is a simplified version - in production would scan all connectors
	
	discoveredItems := []DiscoveredDataItem{}

	// Query PII detections table for matches
	query := `
		SELECT connector_id, schema_name, table_name, column_name, pii_types
		FROM neuronip.pii_detections
		WHERE metadata->>'subject_id' = $1 OR metadata->>'email' = $1`

	rows, err := s.pool.Query(ctx, query, subjectID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var connectorID, schemaName, tableName, columnName string
			var piiTypesJSON []byte

			if err := rows.Scan(&connectorID, &schemaName, &tableName, &columnName, &piiTypesJSON); err != nil {
				continue
			}

			item := DiscoveredDataItem{
				Source:       fmt.Sprintf("%s.%s.%s", schemaName, tableName, columnName),
				Location:     fmt.Sprintf("connector:%s", connectorID),
				DataType:     "pii",
				DiscoveredAt: time.Now(),
			}
			discoveredItems = append(discoveredItems, item)
		}
	}

	// Update request with discovered data
	s.updateDiscoveredData(ctx, requestID, discoveredItems)
}

/* updateDiscoveredData updates discovered data for a DSAR request */
func (s *DSARService) updateDiscoveredData(ctx context.Context, requestID uuid.UUID, items []DiscoveredDataItem) {
	discoveredJSON, _ := json.Marshal(items)

	query := `
		UPDATE neuronip.dsar_requests
		SET discovered_data = $1, status = 'in_progress', updated_at = NOW()
		WHERE id = $2`

	s.pool.Exec(ctx, query, discoveredJSON, requestID)
}

/* GetDSARRequest gets a DSAR request by ID */
func (s *DSARService) GetDSARRequest(ctx context.Context, requestID uuid.UUID) (*DSARRequest, error) {
	var req DSARRequest
	var subjectID, assignedTo sql.NullString
	var completedAt sql.NullTime
	var requestDetailsJSON, discoveredJSON, responseJSON, metadataJSON []byte

	query := `
		SELECT id, request_type, subject_name, subject_email, subject_id, status,
		       requested_at, completed_at, request_details, discovered_data, response_data,
		       assigned_to, metadata, created_at, updated_at
		FROM neuronip.dsar_requests
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, requestID).Scan(
		&req.ID, &req.RequestType, &req.SubjectName, &req.SubjectEmail, &subjectID,
		&req.Status, &req.RequestedAt, &completedAt, &requestDetailsJSON, &discoveredJSON,
		&responseJSON, &assignedTo, &metadataJSON, &req.CreatedAt, &req.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("DSAR request not found: %w", err)
	}

	if subjectID.Valid {
		req.SubjectID = &subjectID.String
	}
	if assignedTo.Valid {
		req.AssignedTo = &assignedTo.String
	}
	if completedAt.Valid {
		req.CompletedAt = &completedAt.Time
	}
	if requestDetailsJSON != nil {
		json.Unmarshal(requestDetailsJSON, &req.RequestDetails)
	}
	if discoveredJSON != nil {
		json.Unmarshal(discoveredJSON, &req.DiscoveredData)
	}
	if responseJSON != nil {
		json.Unmarshal(responseJSON, &req.ResponseData)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &req.Metadata)
	}

	return &req, nil
}

/* CompleteDSARRequest completes a DSAR request */
func (s *DSARService) CompleteDSARRequest(ctx context.Context, requestID uuid.UUID, responseData map[string]interface{}) error {
	responseJSON, _ := json.Marshal(responseData)
	completedAt := time.Now()

	query := `
		UPDATE neuronip.dsar_requests
		SET status = 'completed', completed_at = $1, response_data = $2, updated_at = NOW()
		WHERE id = $3`

	_, err := s.pool.Exec(ctx, query, completedAt, responseJSON, requestID)
	return err
}

/* ListDSARRequests lists DSAR requests */
func (s *DSARService) ListDSARRequests(ctx context.Context, status string, limit int) ([]DSARRequest, error) {
	query := `
		SELECT id, request_type, subject_name, subject_email, subject_id, status,
		       requested_at, completed_at, assigned_to, created_at, updated_at
		FROM neuronip.dsar_requests`

	if status != "" {
		query += fmt.Sprintf(" WHERE status = '%s'", status)
	}

	query += " ORDER BY requested_at DESC LIMIT $1"

	if limit == 0 {
		limit = 100
	}

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list DSAR requests: %w", err)
	}
	defer rows.Close()

	requests := []DSARRequest{}
	for rows.Next() {
		var req DSARRequest
		var subjectID, assignedTo sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&req.ID, &req.RequestType, &req.SubjectName, &req.SubjectEmail, &subjectID,
			&req.Status, &req.RequestedAt, &completedAt, &assignedTo,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if subjectID.Valid {
			req.SubjectID = &subjectID.String
		}
		if assignedTo.Valid {
			req.AssignedTo = &assignedTo.String
		}
		if completedAt.Valid {
			req.CompletedAt = &completedAt.Time
		}

		requests = append(requests, req)
	}

	return requests, nil
}

