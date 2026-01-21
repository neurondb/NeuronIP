package connectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ConnectorType represents a data source connector type */
type ConnectorType string

const (
	ConnectorPostgreSQL  ConnectorType = "postgresql"
	ConnectorMySQL       ConnectorType = "mysql"
	ConnectorSQLServer   ConnectorType = "sqlserver"
	ConnectorOracle      ConnectorType = "oracle"
	ConnectorSnowflake   ConnectorType = "snowflake"
	ConnectorBigQuery    ConnectorType = "bigquery"
	ConnectorRedshift    ConnectorType = "redshift"
	ConnectorDatabricks  ConnectorType = "databricks"
	ConnectorMongoDB     ConnectorType = "mongodb"
	ConnectorElasticsearch ConnectorType = "elasticsearch"
	ConnectorCustom      ConnectorType = "custom"
)

/* DataSourceConnector represents a data source connector */
type DataSourceConnector struct {
	ID             uuid.UUID              `json:"id"`
	Name           string                 `json:"name"`
	ConnectorType  ConnectorType          `json:"connector_type"`
	Enabled        bool                   `json:"enabled"`
	Configuration  map[string]interface{} `json:"configuration"`
	ConnectionString *string              `json:"connection_string,omitempty"`
	Credentials    map[string]interface{} `json:"credentials,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	LastSyncAt     *time.Time             `json:"last_sync_at,omitempty"`
	SyncStatus     string                 `json:"sync_status"`
	SyncError      *string                `json:"sync_error,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

/* ConnectorService provides connector management functionality */
type ConnectorService struct {
	pool      *pgxpool.Pool
	registry  *ConnectorRegistry
}

/* NewConnectorService creates a new connector service */
func NewConnectorService(pool *pgxpool.Pool) *ConnectorService {
	registry := NewConnectorRegistry()
	return &ConnectorService{
		pool:     pool,
		registry: registry,
	}
}

/* CreateConnector creates a new data source connector */
func (s *ConnectorService) CreateConnector(ctx context.Context, connector DataSourceConnector) (*DataSourceConnector, error) {
	connector.ID = uuid.New()
	connector.CreatedAt = time.Now()
	connector.UpdatedAt = time.Now()
	connector.SyncStatus = "idle"

	configJSON, _ := json.Marshal(connector.Configuration)
	credsJSON, _ := json.Marshal(connector.Credentials)
	metadataJSON, _ := json.Marshal(connector.Metadata)

	var connectionString sql.NullString
	if connector.ConnectionString != nil {
		connectionString = sql.NullString{String: *connector.ConnectionString, Valid: true}
	}

	query := `
		INSERT INTO neuronip.data_source_connectors
		(id, name, connector_type, enabled, configuration, connection_string, 
		 credentials, metadata, sync_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		connector.ID, connector.Name, string(connector.ConnectorType), connector.Enabled,
		configJSON, connectionString, credsJSON, metadataJSON,
		connector.SyncStatus, connector.CreatedAt, connector.UpdatedAt,
	).Scan(&connector.ID, &connector.CreatedAt, &connector.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	return &connector, nil
}

/* GetConnector retrieves a connector by ID */
func (s *ConnectorService) GetConnector(ctx context.Context, id uuid.UUID) (*DataSourceConnector, error) {
	var connector DataSourceConnector
	var configJSON, credsJSON, metadataJSON []byte
	var connectionString sql.NullString
	var lastSyncAt sql.NullTime
	var syncError sql.NullString

	query := `
		SELECT id, name, connector_type, enabled, configuration, connection_string,
		       credentials, metadata, last_sync_at, sync_status, sync_error,
		       created_at, updated_at
		FROM neuronip.data_source_connectors
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&connector.ID, &connector.Name, &connector.ConnectorType, &connector.Enabled,
		&configJSON, &connectionString, &credsJSON, &metadataJSON,
		&lastSyncAt, &connector.SyncStatus, &syncError,
		&connector.CreatedAt, &connector.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get connector: %w", err)
	}

	if configJSON != nil {
		json.Unmarshal(configJSON, &connector.Configuration)
	}
	if credsJSON != nil {
		json.Unmarshal(credsJSON, &connector.Credentials)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &connector.Metadata)
	}
	if connectionString.Valid {
		connector.ConnectionString = &connectionString.String
	}
	if lastSyncAt.Valid {
		connector.LastSyncAt = &lastSyncAt.Time
	}
	if syncError.Valid {
		connector.SyncError = &syncError.String
	}

	return &connector, nil
}

/* ListConnectors lists all connectors */
func (s *ConnectorService) ListConnectors(ctx context.Context, enabledOnly bool) ([]DataSourceConnector, error) {
	query := `
		SELECT id, name, connector_type, enabled, configuration, connection_string,
		       credentials, metadata, last_sync_at, sync_status, sync_error,
		       created_at, updated_at
		FROM neuronip.data_source_connectors`
	
	if enabledOnly {
		query += " WHERE enabled = true"
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list connectors: %w", err)
	}
	defer rows.Close()

	var connectors []DataSourceConnector
	for rows.Next() {
		var connector DataSourceConnector
		var configJSON, credsJSON, metadataJSON []byte
		var connectionString sql.NullString
		var lastSyncAt sql.NullTime
		var syncError sql.NullString

		err := rows.Scan(
			&connector.ID, &connector.Name, &connector.ConnectorType, &connector.Enabled,
			&configJSON, &connectionString, &credsJSON, &metadataJSON,
			&lastSyncAt, &connector.SyncStatus, &syncError,
			&connector.CreatedAt, &connector.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if configJSON != nil {
			json.Unmarshal(configJSON, &connector.Configuration)
		}
		if credsJSON != nil {
			json.Unmarshal(credsJSON, &connector.Credentials)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &connector.Metadata)
		}
		if connectionString.Valid {
			connector.ConnectionString = &connectionString.String
		}
		if lastSyncAt.Valid {
			connector.LastSyncAt = &lastSyncAt.Time
		}
		if syncError.Valid {
			connector.SyncError = &syncError.String
		}

		connectors = append(connectors, connector)
	}

	return connectors, nil
}

/* SyncConnector synchronizes a connector (discovers schema) */
func (s *ConnectorService) SyncConnector(ctx context.Context, connectorID uuid.UUID, syncType string) error {
	connector, err := s.GetConnector(ctx, connectorID)
	if err != nil {
		return fmt.Errorf("failed to get connector: %w", err)
	}

	if !connector.Enabled {
		return fmt.Errorf("connector is disabled")
	}

	// Update sync status
	_, err = s.pool.Exec(ctx, `
		UPDATE neuronip.data_source_connectors
		SET sync_status = 'syncing', updated_at = NOW()
		WHERE id = $1`, connectorID)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	// Create sync history record
	syncHistoryID := uuid.New()
	startedAt := time.Now()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO neuronip.connector_sync_history
		(id, connector_id, sync_type, status, started_at)
		VALUES ($1, $2, $3, 'running', $4)`,
		syncHistoryID, connectorID, syncType, startedAt)
	if err != nil {
		return fmt.Errorf("failed to create sync history: %w", err)
	}

	// Get connector implementation from registry
	impl, err := s.registry.GetConnector(connector.ConnectorType)
	if err != nil {
		s.updateSyncError(ctx, connectorID, syncHistoryID, err.Error())
		return fmt.Errorf("connector type not supported: %w", err)
	}

	// Discover schema
	schema, err := impl.DiscoverSchema(ctx, connector)
	if err != nil {
		s.updateSyncError(ctx, connectorID, syncHistoryID, err.Error())
		return fmt.Errorf("failed to discover schema: %w", err)
	}

	// Store discovered schema
	tablesCount, columnsCount, err := s.storeDiscoveredSchema(ctx, connectorID, schema)
	if err != nil {
		s.updateSyncError(ctx, connectorID, syncHistoryID, err.Error())
		return fmt.Errorf("failed to store schema: %w", err)
	}

	// Update sync status to success
	completedAt := time.Now()
	durationMs := int(completedAt.Sub(startedAt).Milliseconds())
	_, err = s.pool.Exec(ctx, `
		UPDATE neuronip.data_source_connectors
		SET sync_status = 'success', last_sync_at = $1, updated_at = $1
		WHERE id = $2`, completedAt, connectorID)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	// Update sync history
	_, err = s.pool.Exec(ctx, `
		UPDATE neuronip.connector_sync_history
		SET status = 'success', completed_at = $1, duration_ms = $2,
		    tables_discovered = $3, columns_discovered = $4
		WHERE id = $5`,
		completedAt, durationMs, tablesCount, columnsCount, syncHistoryID)
	if err != nil {
		return fmt.Errorf("failed to update sync history: %w", err)
	}

	return nil
}

/* updateSyncError updates sync status with error */
func (s *ConnectorService) updateSyncError(ctx context.Context, connectorID, syncHistoryID uuid.UUID, errorMsg string) {
	s.pool.Exec(ctx, `
		UPDATE neuronip.data_source_connectors
		SET sync_status = 'error', sync_error = $1, updated_at = NOW()
		WHERE id = $2`, errorMsg, connectorID)

	s.pool.Exec(ctx, `
		UPDATE neuronip.connector_sync_history
		SET status = 'error', error_message = $1, completed_at = NOW()
		WHERE id = $2`, errorMsg, syncHistoryID)
}

/* storeDiscoveredSchema stores discovered schema in catalog */
func (s *ConnectorService) storeDiscoveredSchema(ctx context.Context, connectorID uuid.UUID, schema *Schema) (int, int, error) {
	tablesCount := 0
	columnsCount := 0

	for _, table := range schema.Tables {
		// Insert or update table
		tableID := uuid.New()
		tableQuery := `
			INSERT INTO neuronip.catalog_tables
			(id, connector_id, schema_name, table_name, table_type, description,
			 row_count, size_bytes, owner, created_at, updated_at, last_analyzed_at,
			 metadata, discovered_at, updated_discovered_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
			ON CONFLICT (connector_id, schema_name, table_name)
			DO UPDATE SET
				table_type = EXCLUDED.table_type,
				description = EXCLUDED.description,
				row_count = EXCLUDED.row_count,
				size_bytes = EXCLUDED.size_bytes,
				owner = EXCLUDED.owner,
				updated_at = EXCLUDED.updated_at,
				last_analyzed_at = EXCLUDED.last_analyzed_at,
				metadata = EXCLUDED.metadata,
				updated_discovered_at = NOW()
			RETURNING id`

		metadataJSON, _ := json.Marshal(table.Metadata)
		var createdAt, updatedAt, lastAnalyzedAt sql.NullTime
		if table.CreatedAt != nil {
			createdAt = sql.NullTime{Time: *table.CreatedAt, Valid: true}
		}
		if table.UpdatedAt != nil {
			updatedAt = sql.NullTime{Time: *table.UpdatedAt, Valid: true}
		}
		if table.LastAnalyzedAt != nil {
			lastAnalyzedAt = sql.NullTime{Time: *table.LastAnalyzedAt, Valid: true}
		}

		err := s.pool.QueryRow(ctx, tableQuery,
			tableID, connectorID, table.SchemaName, table.TableName, table.TableType,
			table.Description, table.RowCount, table.SizeBytes, table.Owner,
			createdAt, updatedAt, lastAnalyzedAt, metadataJSON,
		).Scan(&tableID)
		if err != nil {
			continue
		}

		tablesCount++

		// Insert or update columns
		for _, column := range table.Columns {
			columnQuery := `
				INSERT INTO neuronip.catalog_columns
				(id, table_id, column_name, column_type, ordinal_position,
				 is_nullable, is_primary_key, is_foreign_key, default_value,
				 description, metadata, discovered_at, updated_discovered_at)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
				ON CONFLICT (table_id, column_name)
				DO UPDATE SET
					column_type = EXCLUDED.column_type,
					ordinal_position = EXCLUDED.ordinal_position,
					is_nullable = EXCLUDED.is_nullable,
					is_primary_key = EXCLUDED.is_primary_key,
					is_foreign_key = EXCLUDED.is_foreign_key,
					default_value = EXCLUDED.default_value,
					description = EXCLUDED.description,
					metadata = EXCLUDED.metadata,
					updated_discovered_at = NOW()`

			columnMetadataJSON, _ := json.Marshal(column.Metadata)
			var defaultValue sql.NullString
			if column.DefaultValue != nil {
				defaultValue = sql.NullString{String: *column.DefaultValue, Valid: true}
			}

			_, err = s.pool.Exec(ctx, columnQuery,
				tableID, column.ColumnName, column.ColumnType, column.OrdinalPosition,
				column.IsNullable, column.IsPrimaryKey, column.IsForeignKey,
				defaultValue, column.Description, columnMetadataJSON,
			)
			if err != nil {
				continue
			}

			columnsCount++
		}
	}

	return tablesCount, columnsCount, nil
}

/* Schema represents a discovered database schema */
type Schema struct {
	Tables []Table `json:"tables"`
}

/* Table represents a discovered table */
type Table struct {
	SchemaName     string                 `json:"schema_name"`
	TableName      string                 `json:"table_name"`
	TableType      string                 `json:"table_type"`
	Description    *string                `json:"description,omitempty"`
	RowCount       *int64                 `json:"row_count,omitempty"`
	SizeBytes      *int64                 `json:"size_bytes,omitempty"`
	Owner          *string                `json:"owner,omitempty"`
	CreatedAt      *time.Time              `json:"created_at,omitempty"`
	UpdatedAt      *time.Time              `json:"updated_at,omitempty"`
	LastAnalyzedAt *time.Time              `json:"last_analyzed_at,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Columns        []Column                `json:"columns"`
}

/* Column represents a discovered column */
type Column struct {
	ColumnName     string                 `json:"column_name"`
	ColumnType     string                 `json:"column_type"`
	OrdinalPosition int                   `json:"ordinal_position"`
	IsNullable     bool                   `json:"is_nullable"`
	IsPrimaryKey   bool                   `json:"is_primary_key"`
	IsForeignKey   bool                   `json:"is_foreign_key"`
	DefaultValue   *string                `json:"default_value,omitempty"`
	Description    *string                `json:"description,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
