package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/* ExecutionStreamService provides real-time execution streaming */
type ExecutionStreamService struct {
	connections map[uuid.UUID]*websocket.Conn
}

/* NewExecutionStreamService creates a new execution stream service */
func NewExecutionStreamService() *ExecutionStreamService {
	return &ExecutionStreamService{
		connections: make(map[uuid.UUID]*websocket.Conn),
	}
}

/* StreamExecutionEvent represents an execution event */
type StreamExecutionEvent struct {
	ExecutionID uuid.UUID              `json:"execution_id"`
	StepID      string                 `json:"step_id"`
	EventType   string                 `json:"event_type"` // "step_started", "step_completed", "step_failed"
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
}

/* StreamExecution streams execution events */
func (ess *ExecutionStreamService) StreamExecution(ctx context.Context, executionID uuid.UUID, conn *websocket.Conn) error {
	ess.connections[executionID] = conn
	defer delete(ess.connections, executionID)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Keep connection alive
			time.Sleep(1 * time.Second)
		}
	}
}

/* SendEvent sends an event to a stream */
func (ess *ExecutionStreamService) SendEvent(executionID uuid.UUID, event StreamExecutionEvent) error {
	conn, exists := ess.connections[executionID]
	if !exists {
		return fmt.Errorf("no connection for execution %s", executionID)
	}
	
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}
	
	return conn.WriteMessage(websocket.TextMessage, eventJSON)
}
