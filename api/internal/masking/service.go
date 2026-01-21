package masking

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* MaskingService provides data masking functionality */
type MaskingService struct {
	pool *pgxpool.Pool
}

/* NewMaskingService creates a new masking service */
func NewMaskingService(pool *pgxpool.Pool) *MaskingService {
	return &MaskingService{pool: pool}
}

/* MaskingPolicy represents a data masking policy */
type MaskingPolicy struct {
	ID             uuid.UUID              `json:"id"`
	ConnectorID    *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName     string                 `json:"schema_name"`
	TableName      string                 `json:"table_name"`
	ColumnName     string                 `json:"column_name"`
	MaskingType    string                 `json:"masking_type"` // "tokenization", "encryption", "format_preserving", "partial", "full"
	Algorithm      string                 `json:"algorithm,omitempty"`
	MaskingRule    string                 `json:"masking_rule,omitempty"`
	UserRoles      []string               `json:"user_roles"` // Roles that see masked data
	Enabled        bool                   `json:"enabled"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

/* MaskingAlgorithms provides masking algorithm implementations */
type MaskingAlgorithms struct{}

/* NewMaskingAlgorithms creates masking algorithms */
func NewMaskingAlgorithms() *MaskingAlgorithms {
	return &MaskingAlgorithms{}
}

/* CreateMaskingPolicy creates a masking policy */
func (s *MaskingService) CreateMaskingPolicy(ctx context.Context, policy MaskingPolicy) (*MaskingPolicy, error) {
	policy.ID = uuid.New()
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	userRolesJSON, _ := json.Marshal(policy.UserRoles)
	metadataJSON, _ := json.Marshal(policy.Metadata)

	var connectorID sql.NullString
	if policy.ConnectorID != nil {
		connectorID = sql.NullString{String: policy.ConnectorID.String(), Valid: true}
	}

	query := `
		INSERT INTO neuronip.masking_policies
		(id, connector_id, schema_name, table_name, column_name, masking_type, algorithm,
		 masking_rule, user_roles, enabled, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		policy.ID, connectorID, policy.SchemaName, policy.TableName, policy.ColumnName,
		policy.MaskingType, policy.Algorithm, policy.MaskingRule, userRolesJSON,
		policy.Enabled, metadataJSON, policy.CreatedAt, policy.UpdatedAt,
	).Scan(&policy.ID, &policy.CreatedAt, &policy.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create masking policy: %w", err)
	}

	return &policy, nil
}

/* GetMaskingPolicy gets a masking policy */
func (s *MaskingService) GetMaskingPolicy(ctx context.Context, connectorID *uuid.UUID, schemaName, tableName, columnName string) (*MaskingPolicy, error) {
	var policy MaskingPolicy
	var connectorIDStr sql.NullString
	var userRolesJSON, metadataJSON []byte

	var connectorIDParam sql.NullString
	if connectorID != nil {
		connectorIDParam = sql.NullString{String: connectorID.String(), Valid: true}
	}

	query := `
		SELECT id, connector_id, schema_name, table_name, column_name, masking_type,
		       algorithm, masking_rule, user_roles, enabled, metadata, created_at, updated_at
		FROM neuronip.masking_policies
		WHERE (connector_id = $1 OR ($1 IS NULL AND connector_id IS NULL))
		  AND schema_name = $2 AND table_name = $3 AND column_name = $4
		  AND enabled = true
		LIMIT 1`

	err := s.pool.QueryRow(ctx, query, connectorIDParam, schemaName, tableName, columnName).Scan(
		&policy.ID, &connectorIDStr, &policy.SchemaName, &policy.TableName, &policy.ColumnName,
		&policy.MaskingType, &policy.Algorithm, &policy.MaskingRule, &userRolesJSON,
		&policy.Enabled, &metadataJSON, &policy.CreatedAt, &policy.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("masking policy not found: %w", err)
	}

	if connectorIDStr.Valid {
		connUUID, _ := uuid.Parse(connectorIDStr.String)
		policy.ConnectorID = &connUUID
	}
	if userRolesJSON != nil {
		json.Unmarshal(userRolesJSON, &policy.UserRoles)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &policy.Metadata)
	}

	return &policy, nil
}

/* ApplyMasking applies masking to a value based on policy */
func (s *MaskingService) ApplyMasking(ctx context.Context, userRole string, connectorID *uuid.UUID, schemaName, tableName, columnName string, value interface{}) (interface{}, error) {
	policy, err := s.GetMaskingPolicy(ctx, connectorID, schemaName, tableName, columnName)
	if err != nil {
		// No policy means no masking
		return value, nil
	}

	// Check if user role is exempt from masking
	for _, exemptRole := range policy.UserRoles {
		if exemptRole == userRole || exemptRole == "*" {
			// User can see unmasked data
			return value, nil
		}
	}

	// Apply masking based on type
	algorithms := NewMaskingAlgorithms()
	valueStr := fmt.Sprintf("%v", value)

	switch policy.MaskingType {
	case "tokenization":
		return algorithms.Tokenize(valueStr, policy.Algorithm)
	case "encryption":
		return algorithms.Encrypt(valueStr, policy.Algorithm)
	case "format_preserving":
		return algorithms.FormatPreservingMask(valueStr, policy.MaskingRule)
	case "partial":
		return algorithms.PartialMask(valueStr, policy.MaskingRule)
	case "full":
		return "[MASKED]", nil
	default:
		return value, nil
	}
}
