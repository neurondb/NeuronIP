package cdc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* CDCManager coordinates CDC operations across different database types */
type CDCManager struct {
	postgresCDC *PostgresCDC
	mysqlCDC    *MySQLCDC
	pool        *pgxpool.Pool
}

/* NewCDCManager creates a new CDC manager */
func NewCDCManager(pool *pgxpool.Pool) *CDCManager {
	return &CDCManager{
		postgresCDC: NewPostgresCDC(pool),
		mysqlCDC:    NewMySQLCDC(),
		pool:        pool,
	}
}

/* StartCDCForDataSource starts CDC for a data source */
func (m *CDCManager) StartCDCForDataSource(ctx context.Context, dataSourceType string, config map[string]interface{}) error {
	switch dataSourceType {
	case "postgresql", "postgres":
		return m.postgresCDC.StartCDC(ctx, config)
	case "mysql":
		return m.mysqlCDC.StartCDC(ctx, config)
	default:
		return fmt.Errorf("CDC not supported for data source type: %s", dataSourceType)
	}
}

/* StopCDCForDataSource stops CDC for a data source */
func (m *CDCManager) StopCDCForDataSource(ctx context.Context, dataSourceType string) error {
	switch dataSourceType {
	case "postgresql", "postgres":
		return m.postgresCDC.StopCDC(ctx)
	case "mysql":
		return m.mysqlCDC.StopCDC(ctx)
	default:
		return fmt.Errorf("CDC not supported for data source type: %s", dataSourceType)
	}
}

/* ProcessChanges processes CDC changes and applies them */
func (m *CDCManager) ProcessChanges(ctx context.Context, dataSourceID string, dataSourceType string, changes []ChangeEvent) error {
	// Process each change event
	for _, change := range changes {
		// Save checkpoint
		checkpoint := map[string]interface{}{
			"lsn":       change.LSN,
			"timestamp": change.Timestamp,
		}
		
		var cdc CDC
		switch dataSourceType {
		case "postgresql", "postgres":
			cdc = m.postgresCDC
		case "mysql":
			cdc = m.mysqlCDC
		default:
			continue
		}
		
		if err := cdc.SaveCheckpoint(ctx, dataSourceID, change.Table, checkpoint); err != nil {
			return fmt.Errorf("failed to save checkpoint: %w", err)
		}
		
		// In production, would apply changes to target system
		// This could be:
		// 1. Apply to warehouse tables
		// 2. Trigger ETL transformations
		// 3. Update knowledge graph
	}
	
	return nil
}

/* CDC interface for different database types */
type CDC interface {
	StartCDC(ctx context.Context, config map[string]interface{}) error
	StopCDC(ctx context.Context) error
	GetChanges(ctx context.Context, lastPosition interface{}) ([]ChangeEvent, error)
	SaveCheckpoint(ctx context.Context, dataSourceID string, tableName string, checkpoint map[string]interface{}) error
	GetCheckpoint(ctx context.Context, dataSourceID string, tableName string) (map[string]interface{}, error)
}
