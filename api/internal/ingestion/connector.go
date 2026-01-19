package ingestion

import (
	"context"
	"fmt"
	"time"
)

/* Connector defines the interface for all data source connectors */
type Connector interface {
	// Connect establishes a connection to the data source
	Connect(ctx context.Context, config map[string]interface{}) error
	
	// Disconnect closes the connection
	Disconnect(ctx context.Context) error
	
	// TestConnection tests if the connection is valid
	TestConnection(ctx context.Context) error
	
	// DiscoverSchema discovers the schema of the data source
	DiscoverSchema(ctx context.Context) (*Schema, error)
	
	// Sync performs a full or incremental sync
	Sync(ctx context.Context, options SyncOptions) (*SyncResult, error)
	
	// GetConnectorType returns the type identifier for this connector
	GetConnectorType() string
	
	// GetMetadata returns connector-specific metadata
	GetMetadata() ConnectorMetadata
}

/* Schema represents the discovered schema of a data source */
type Schema struct {
	Tables      []TableSchema `json:"tables"`
	Views       []ViewSchema  `json:"views,omitempty"`
	LastUpdated time.Time     `json:"last_updated"`
}

/* TableSchema represents a table in the schema */
type TableSchema struct {
	Name        string            `json:"name"`
	Columns     []ColumnSchema    `json:"columns"`
	PrimaryKeys []string          `json:"primary_keys,omitempty"`
	Indexes     []IndexSchema     `json:"indexes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

/* ColumnSchema represents a column in a table */
type ColumnSchema struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue *string `json:"default_value,omitempty"`
	MaxLength    *int   `json:"max_length,omitempty"`
}

/* IndexSchema represents an index */
type IndexSchema struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

/* ViewSchema represents a view */
type ViewSchema struct {
	Name     string         `json:"name"`
	Columns  []ColumnSchema `json:"columns"`
	Definition string       `json:"definition,omitempty"`
}

/* SyncOptions configures how a sync should be performed */
type SyncOptions struct {
	Mode          SyncMode                `json:"mode"` // "full" or "incremental"
	Tables        []string                `json:"tables,omitempty"` // Empty means all tables
	Since         *time.Time              `json:"since,omitempty"` // For incremental syncs
	BatchSize     int                     `json:"batch_size,omitempty"`
	Transformations map[string]interface{} `json:"transformations,omitempty"`
}

/* SyncMode defines the type of sync */
type SyncMode string

const (
	SyncModeFull        SyncMode = "full"
	SyncModeIncremental SyncMode = "incremental"
)

/* SyncResult contains the results of a sync operation */
type SyncResult struct {
	RowsSynced    int64                  `json:"rows_synced"`
	TablesSynced  []string               `json:"tables_synced"`
	Duration      time.Duration          `json:"duration"`
	Errors        []SyncError            `json:"errors,omitempty"`
	Checkpoint    map[string]interface{} `json:"checkpoint,omitempty"` // For incremental syncs
}

/* SyncError represents an error during sync */
type SyncError struct {
	Table   string `json:"table"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

/* ConnectorMetadata provides information about a connector */
type ConnectorMetadata struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Capabilities []string `json:"capabilities"` // e.g., "incremental", "schema_discovery", "cdc"
}

/* BaseConnector provides common functionality for connectors */
type BaseConnector struct {
	connectorType string
	config        map[string]interface{}
	connected     bool
	metadata      ConnectorMetadata
}

/* NewBaseConnector creates a new base connector */
func NewBaseConnector(connectorType string, metadata ConnectorMetadata) *BaseConnector {
	return &BaseConnector{
		connectorType: connectorType,
		metadata:      metadata,
		connected:     false,
	}
}

/* GetConnectorType returns the connector type */
func (b *BaseConnector) GetConnectorType() string {
	return b.connectorType
}

/* GetMetadata returns connector metadata */
func (b *BaseConnector) GetMetadata() ConnectorMetadata {
	return b.metadata
}

/* SetConfig sets the connector configuration */
func (b *BaseConnector) SetConfig(config map[string]interface{}) {
	b.config = config
}

/* GetConfig returns the connector configuration */
func (b *BaseConnector) GetConfig() map[string]interface{} {
	return b.config
}

/* IsConnected returns whether the connector is connected */
func (b *BaseConnector) IsConnected() bool {
	return b.connected
}

/* SetConnected sets the connected state */
func (b *BaseConnector) SetConnected(connected bool) {
	b.connected = connected
}

/* ConnectorPool manages a pool of connector instances */
type ConnectorPool struct {
	connectors map[string]Connector
	factories  map[string]ConnectorFactory
	maxSize    int
}

/* NewConnectorPool creates a new connector pool */
func NewConnectorPool(maxSize int) *ConnectorPool {
	return &ConnectorPool{
		connectors: make(map[string]Connector),
		factories:  make(map[string]ConnectorFactory),
		maxSize:    maxSize,
	}
}

/* GetConnector retrieves or creates a connector */
func (p *ConnectorPool) GetConnector(ctx context.Context, connectorType string, factory ConnectorFactory, config map[string]interface{}) (Connector, error) {
	key := fmt.Sprintf("%s:%v", connectorType, config)
	
	if conn, exists := p.connectors[key]; exists {
		if baseConn, ok := conn.(interface{ IsConnected() bool }); ok && baseConn.IsConnected() {
			return conn, nil
		}
	}
	
	if len(p.connectors) >= p.maxSize {
		return nil, fmt.Errorf("connector pool is full")
	}
	
	conn := factory(connectorType)
	if baseConn, ok := conn.(interface{ SetConfig(map[string]interface{}) }); ok {
		baseConn.SetConfig(config)
	}
	
	if err := conn.Connect(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	
	p.connectors[key] = conn
	return conn, nil
}

/* ReleaseConnector releases a connector back to the pool */
func (p *ConnectorPool) ReleaseConnector(ctx context.Context, conn Connector) error {
	if err := conn.Disconnect(ctx); err != nil {
		return err
	}
	
	// Remove from pool
	for key, c := range p.connectors {
		if c == conn {
			delete(p.connectors, key)
			break
		}
	}
	
	return nil
}

/* ConnectorFactory creates a new connector instance */
type ConnectorFactory func(connectorType string) Connector

/* RetryConfig configures retry behavior */
type RetryConfig struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

/* DefaultRetryConfig returns default retry configuration */
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

/* Retry executes a function with retry logic */
func Retry(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error
	delay := config.InitialDelay
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		
		if attempt < config.MaxAttempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				delay = time.Duration(float64(delay) * config.Multiplier)
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
			}
		}
	}
	
	return fmt.Errorf("retry failed after %d attempts: %w", config.MaxAttempts, lastErr)
}
