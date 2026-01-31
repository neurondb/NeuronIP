package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* TracingService provides agent execution tracing functionality */
type TracingService struct {
	pool *pgxpool.Pool
}

/* NewTracingService creates a new tracing service */
func NewTracingService(pool *pgxpool.Pool) *TracingService {
	return &TracingService{pool: pool}
}

/* Trace represents an agent execution trace */
type Trace struct {
	ID            uuid.UUID              `json:"id"`
	AgentID       string                 `json:"agent_id"`
	SessionID     *string                `json:"session_id,omitempty"`
	Task          string                 `json:"task"`
	Steps         []TraceStep            `json:"steps"`
	Status        string                 `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Duration      *time.Duration         `json:"duration,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* TraceStep represents a single step in an execution trace */
type TraceStep struct {
	ID          uuid.UUID              `json:"id"`
	TraceID     uuid.UUID              `json:"trace_id"`
	StepNumber  int                    `json:"step_number"`
	StepType    string                 `json:"step_type"` // "tool_call", "reasoning", "response"
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	ToolName    *string                `json:"tool_name,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Timestamp   time.Time              `json:"timestamp"`
	Error       *string                `json:"error,omitempty"`
}

/* StartTrace starts a new execution trace */
func (ts *TracingService) StartTrace(ctx context.Context, agentID, task string, sessionID *string) (*Trace, error) {
	traceID := uuid.New()
	now := time.Now()
	
	trace := &Trace{
		ID:        traceID,
		AgentID:   agentID,
		SessionID: sessionID,
		Task:      task,
		Steps:     []TraceStep{},
		Status:    "running",
		StartTime: now,
		Metadata:  make(map[string]interface{}),
	}
	
	// Store trace
	metadataJSON, _ := json.Marshal(trace.Metadata)
	query := `
		INSERT INTO neuronip.agent_traces 
		(id, agent_id, session_id, task, status, start_time, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := ts.pool.Exec(ctx, query,
		traceID, agentID, sessionID, task, trace.Status, now, metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start trace: %w", err)
	}
	
	return trace, nil
}

/* AddStep adds a step to a trace */
func (ts *TracingService) AddStep(ctx context.Context, traceID uuid.UUID, step TraceStep) error {
	step.ID = uuid.New()
	step.TraceID = traceID
	step.Timestamp = time.Now()
	
	inputJSON, _ := json.Marshal(step.Input)
	outputJSON, _ := json.Marshal(step.Output)
	
	query := `
		INSERT INTO neuronip.agent_trace_steps 
		(id, trace_id, step_number, step_type, description, input_data, output_data, tool_name, duration, timestamp, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	
	_, err := ts.pool.Exec(ctx, query,
		step.ID, step.TraceID, step.StepNumber, step.StepType, step.Description,
		inputJSON, outputJSON, step.ToolName, step.Duration, step.Timestamp, step.Error,
	)
	return err
}

/* CompleteTrace completes a trace */
func (ts *TracingService) CompleteTrace(ctx context.Context, traceID uuid.UUID, status string) error {
	now := time.Now()
	
	// Get start time
	var startTime time.Time
	err := ts.pool.QueryRow(ctx, `SELECT start_time FROM neuronip.agent_traces WHERE id = $1`, traceID).Scan(&startTime)
	if err != nil {
		return fmt.Errorf("failed to get start time: %w", err)
	}
	
	duration := now.Sub(startTime)
	
	query := `
		UPDATE neuronip.agent_traces
		SET status = $1, end_time = $2, duration_ms = $3
		WHERE id = $4
	`
	
	_, err = ts.pool.Exec(ctx, query, status, now, duration.Milliseconds(), traceID)
	return err
}

/* GetTrace retrieves a trace with all steps */
func (ts *TracingService) GetTrace(ctx context.Context, traceID uuid.UUID) (*Trace, error) {
	// Get trace
	var trace Trace
	var sessionID *string
	var endTime *time.Time
	var durationMs *int64
	var metadataJSON json.RawMessage
	
	query := `
		SELECT id, agent_id, session_id, task, status, start_time, end_time, duration_ms, metadata
		FROM neuronip.agent_traces
		WHERE id = $1
	`
	
	err := ts.pool.QueryRow(ctx, query, traceID).Scan(
		&trace.ID, &trace.AgentID, &sessionID, &trace.Task, &trace.Status,
		&trace.StartTime, &endTime, &durationMs, &metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}
	
	trace.SessionID = sessionID
	trace.EndTime = endTime
	if durationMs != nil {
		duration := time.Duration(*durationMs) * time.Millisecond
		trace.Duration = &duration
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &trace.Metadata)
	}
	
	// Get steps
	stepsQuery := `
		SELECT id, trace_id, step_number, step_type, description, input_data, output_data, tool_name, duration, timestamp, error_message
		FROM neuronip.agent_trace_steps
		WHERE trace_id = $1
		ORDER BY step_number ASC
	`
	
	rows, err := ts.pool.Query(ctx, stepsQuery, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get steps: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var step TraceStep
		var inputJSON, outputJSON json.RawMessage
		var toolName *string
		var errorMsg *string
		
		err := rows.Scan(
			&step.ID, &step.TraceID, &step.StepNumber, &step.StepType, &step.Description,
			&inputJSON, &outputJSON, &toolName, &step.Duration, &step.Timestamp, &errorMsg,
		)
		if err != nil {
			continue
		}
		
		step.ToolName = toolName
		step.Error = errorMsg
		if inputJSON != nil {
			json.Unmarshal(inputJSON, &step.Input)
		}
		if outputJSON != nil {
			json.Unmarshal(outputJSON, &step.Output)
		}
		
		trace.Steps = append(trace.Steps, step)
	}
	
	return &trace, nil
}

/* ListTraces lists traces for an agent */
func (ts *TracingService) ListTraces(ctx context.Context, agentID *string, limit int) ([]Trace, error) {
	if limit <= 0 {
		limit = 100
	}
	
	query := `
		SELECT id, agent_id, session_id, task, status, start_time, end_time, duration_ms, metadata
		FROM neuronip.agent_traces
		WHERE 1=1
	`
	
	args := []interface{}{}
	argIndex := 1
	
	if agentID != nil {
		query += fmt.Sprintf(" AND agent_id = $%d", argIndex)
		args = append(args, *agentID)
		argIndex++
	}
	
	query += " ORDER BY start_time DESC LIMIT $" + fmt.Sprintf("%d", argIndex)
	args = append(args, limit)
	
	rows, err := ts.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list traces: %w", err)
	}
	defer rows.Close()
	
	var traces []Trace
	for rows.Next() {
		var trace Trace
		var sessionID *string
		var endTime *time.Time
		var durationMs *int64
		var metadataJSON json.RawMessage
		
		err := rows.Scan(
			&trace.ID, &trace.AgentID, &sessionID, &trace.Task, &trace.Status,
			&trace.StartTime, &endTime, &durationMs, &metadataJSON,
		)
		if err != nil {
			continue
		}
		
		trace.SessionID = sessionID
		trace.EndTime = endTime
		if durationMs != nil {
			duration := time.Duration(*durationMs) * time.Millisecond
			trace.Duration = &duration
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &trace.Metadata)
		}
		
		traces = append(traces, trace)
	}
	
	return traces, nil
}
