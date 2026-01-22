package cdc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

/* MySQLCDC implements CDC for MySQL using binlog replication */
type MySQLCDC struct {
	syncer   *replication.BinlogSyncer
	streamer *replication.BinlogStreamer
	pool     *pgxpool.Pool              // For storing checkpoints
	config   map[string]interface{}     // Connection config
	mu       sync.RWMutex                // Protects syncer and streamer
}

/* NewMySQLCDC creates a new MySQL CDC instance */
func NewMySQLCDC(pool *pgxpool.Pool) *MySQLCDC {
	return &MySQLCDC{
		pool: pool,
	}
}

/* StartCDC starts the CDC replication process */
func (m *MySQLCDC) StartCDC(ctx context.Context, config map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store config
	m.config = config

	// Extract connection parameters
	host, _ := config["host"].(string)
	port, _ := config["port"].(float64)
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)
	serverID, _ := config["server_id"].(float64)

	if host == "" {
		return fmt.Errorf("host is required for MySQL CDC")
	}
	if user == "" {
		return fmt.Errorf("user is required for MySQL CDC")
	}

	// Default values
	if port == 0 {
		port = 3306
	}
	if serverID == 0 {
		serverID = 100 // Default server ID
	}

	// Create binlog syncer
	cfg := replication.BinlogSyncerConfig{
		ServerID: uint32(serverID),
		Flavor:   "mysql",
		Host:     host,
		Port:     uint16(port),
		User:     user,
		Password: password,
	}

	syncer := replication.NewBinlogSyncer(&cfg)
	m.syncer = syncer

	// Get current binlog position (or use provided position)
	var position mysql.Position
	if pos, ok := config["position"].(map[string]interface{}); ok {
		if file, ok := pos["file"].(string); ok {
			position.Name = file
		}
		if posVal, ok := pos["pos"].(float64); ok {
			position.Pos = uint32(posVal)
		}
	}

	// If no position provided, start from current position
	if position.Name == "" {
		// Connect to MySQL to get current position
		conn, err := mysql.Connect(fmt.Sprintf("%s:%d", host, uint16(port)), user, password, "")
		if err != nil {
			// If we can't get position, start from beginning
			position.Name = "mysql-bin.000001"
			position.Pos = 4
		} else {
			pos, err := conn.Execute("SHOW MASTER STATUS")
			if err == nil && len(pos.Values) > 0 {
				if file, ok := pos.Values[0][0].(string); ok {
					position.Name = file
				}
				if posVal, ok := pos.Values[0][1].(uint64); ok {
					position.Pos = uint32(posVal)
				}
			}
			conn.Close()
		}
	}

	// Start streaming from position
	streamer, err := syncer.StartSync(position)
	if err != nil {
		syncer.Close()
		m.syncer = nil
		return fmt.Errorf("failed to start binlog sync: %w", err)
	}

	m.streamer = streamer
	return nil
}

/* StopCDC stops the CDC replication process */
func (m *MySQLCDC) StopCDC(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.syncer != nil {
		m.syncer.Close()
		m.syncer = nil
	}
	m.streamer = nil
	m.config = nil
	return nil
}

/* GetChanges retrieves changes from the binlog */
func (m *MySQLCDC) GetChanges(ctx context.Context, lastPosition interface{}) ([]ChangeEvent, error) {
	m.mu.RLock()
	streamer := m.streamer
	config := m.config
	m.mu.RUnlock()

	if config == nil {
		return nil, fmt.Errorf("CDC not started - call StartCDC first")
	}

	// Try to get changes from binlog streamer if available
	if streamer != nil {
		changes, err := m.getChangesFromBinlog(ctx, streamer, lastPosition)
		if err == nil {
			return changes, nil
		}
		// If binlog fails, fall back to polling
	}

	// Fallback to polling if binlog is not available or fails
	if m.pool != nil {
		return m.getChangesFromPolling(ctx, lastPosition)
	}

	return []ChangeEvent{}, nil
}

/* getChangesFromBinlog retrieves changes from MySQL binlog stream */
func (m *MySQLCDC) getChangesFromBinlog(ctx context.Context, streamer *replication.BinlogStreamer, lastPosition interface{}) ([]ChangeEvent, error) {
	var changes []ChangeEvent
	maxEvents := 1000 // Limit number of events per call
	eventCount := 0

	// Set timeout for reading events
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for eventCount < maxEvents {
		ev, err := streamer.GetEvent(timeoutCtx)
		if err != nil {
			if err == context.DeadlineExceeded || err == context.Canceled {
				// Timeout is expected - return what we have
				break
			}
			return changes, fmt.Errorf("failed to read binlog event: %w", err)
		}

		// Convert binlog event to ChangeEvent
		change, err := m.convertBinlogEvent(ev)
		if err != nil {
			// Skip events we can't convert
			continue
		}
		if change != nil {
			changes = append(changes, *change)
			eventCount++
		}
	}

	return changes, nil
}

/* getChangesFromPolling retrieves changes using polling approach */
func (m *MySQLCDC) getChangesFromPolling(ctx context.Context, lastPosition interface{}) ([]ChangeEvent, error) {
	binlogFile, _ := lastPosition.(map[string]interface{})["binlog_file"].(string)
	binlogPos, _ := lastPosition.(map[string]interface{})["binlog_pos"].(uint32)
	
	if binlogFile == "" {
		binlogFile = "mysql-bin.000001"
	}
	
	// Query change log table (if exists)
	query := `
		SELECT table_name, operation, binlog_file, binlog_pos, timestamp, old_data, new_data
		FROM neuronip.cdc_changes
		WHERE binlog_file > $1 OR (binlog_file = $1 AND binlog_pos > $2)
		ORDER BY binlog_file, binlog_pos ASC
		LIMIT 1000`
	
	rows, err := m.pool.Query(ctx, query, binlogFile, binlogPos)
	if err != nil {
		return []ChangeEvent{}, nil // Table may not exist
	}
	defer rows.Close()
	
	var changes []ChangeEvent
	for rows.Next() {
		var change ChangeEvent
		var oldDataJSON, newDataJSON []byte
		var binlogFileVal string
		var binlogPosVal uint32
		
		err := rows.Scan(
			&change.Table,
			&change.Operation,
			&binlogFileVal,
			&binlogPosVal,
			&change.Timestamp,
			&oldDataJSON,
			&newDataJSON,
		)
		if err != nil {
			continue
		}
		
		// Convert binlog position to LSN format
		change.LSN = fmt.Sprintf("%s:%d", binlogFileVal, binlogPosVal)
		
		if oldDataJSON != nil {
			json.Unmarshal(oldDataJSON, &change.OldData)
		}
		if newDataJSON != nil {
			json.Unmarshal(newDataJSON, &change.NewData)
		}
		
		changes = append(changes, change)
	}
	
	return changes, nil
}

/* convertBinlogEvent converts a binlog event to ChangeEvent */
func (m *MySQLCDC) convertBinlogEvent(ev *replication.BinlogEvent) (*ChangeEvent, error) {
	if ev == nil {
		return nil, fmt.Errorf("nil binlog event")
	}

	change := &ChangeEvent{
		LSN:       fmt.Sprintf("%s:%d", ev.Header.LogPos, ev.Header.LogPos),
		Timestamp: time.Unix(int64(ev.Header.Timestamp), 0),
		OldData:   make(map[string]interface{}),
		NewData:   make(map[string]interface{}),
	}

	// Handle different event types
	switch e := ev.Event.(type) {
	case *replication.RowsEvent:
		// Extract table name
		change.Table = string(e.Table.Schema) + "." + string(e.Table.Table)

		// Handle row events
		switch ev.Header.EventType {
		case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
			change.Operation = "insert"
			// Extract new row data
			if len(e.Rows) > 0 {
				row := e.Rows[0]
				for i, col := range e.Table.Columns {
					if i < len(row) {
						change.NewData[col.Name] = row[i]
					}
				}
			}

		case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
			change.Operation = "update"
			// Extract old and new row data
			if len(e.Rows) >= 2 {
				oldRow := e.Rows[0]
				newRow := e.Rows[1]
				for i, col := range e.Table.Columns {
					if i < len(oldRow) {
						change.OldData[col.Name] = oldRow[i]
					}
					if i < len(newRow) {
						change.NewData[col.Name] = newRow[i]
					}
				}
			}

		case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
			change.Operation = "delete"
			// Extract old row data
			if len(e.Rows) > 0 {
				row := e.Rows[0]
				for i, col := range e.Table.Columns {
					if i < len(row) {
						change.OldData[col.Name] = row[i]
					}
				}
			}

		default:
			// Unknown event type, skip
			return nil, nil
		}

	default:
		// Not a row event, skip
		return nil, nil
	}

	return change, nil
}

/* SaveCheckpoint saves a CDC checkpoint */
func (m *MySQLCDC) SaveCheckpoint(ctx context.Context, dataSourceID string, tableName string, checkpoint map[string]interface{}) error {
	if m.pool == nil {
		return fmt.Errorf("database pool not available")
	}

	checkpointJSON, _ := json.Marshal(checkpoint)
	binlogFile := ""
	binlogPos := uint32(0)

	if file, ok := checkpoint["binlog_file"].(string); ok {
		binlogFile = file
	}
	if pos, ok := checkpoint["binlog_pos"].(uint32); ok {
		binlogPos = pos
	}

	// Store checkpoint (using same table structure as PostgreSQL)
	query := `
		INSERT INTO neuronip.cdc_checkpoints 
		(data_source_id, table_name, checkpoint_data, last_binlog_file, last_binlog_pos, last_timestamp, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (data_source_id, table_name) 
		DO UPDATE SET 
		checkpoint_data = EXCLUDED.checkpoint_data,
		last_binlog_file = EXCLUDED.last_binlog_file,
		last_binlog_pos = EXCLUDED.last_binlog_pos,
		last_timestamp = EXCLUDED.last_timestamp,
		updated_at = EXCLUDED.updated_at`

	_, err := m.pool.Exec(ctx, query, dataSourceID, tableName, checkpointJSON, binlogFile, binlogPos)
	return err
}

/* GetCheckpoint retrieves a CDC checkpoint */
func (m *MySQLCDC) GetCheckpoint(ctx context.Context, dataSourceID string, tableName string) (map[string]interface{}, error) {
	if m.pool == nil {
		return nil, fmt.Errorf("database pool not available")
	}

	var checkpointJSON []byte
	var binlogFile *string
	var binlogPos *uint32

	query := `
		SELECT checkpoint_data, last_binlog_file, last_binlog_pos
		FROM neuronip.cdc_checkpoints
		WHERE data_source_id = $1 AND table_name = $2`

	err := m.pool.QueryRow(ctx, query, dataSourceID, tableName).Scan(&checkpointJSON, &binlogFile, &binlogPos)
	if err != nil {
		return nil, err
	}

	var checkpoint map[string]interface{}
	if err := json.Unmarshal(checkpointJSON, &checkpoint); err != nil {
		return nil, err
	}

	if binlogFile != nil {
		checkpoint["binlog_file"] = *binlogFile
	}
	if binlogPos != nil {
		checkpoint["binlog_pos"] = *binlogPos
	}

	return checkpoint, nil
}
