package workflows

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/agent"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides agent workflows functionality */
type Service struct {
	pool           *pgxpool.Pool
	agentClient    *agent.Client
	neurondbClient *neurondb.Client
}

/* NewService creates a new workflows service */
func NewService(pool *pgxpool.Pool, agentClient *agent.Client, neurondbClient *neurondb.Client) *Service {
	return &Service{
		pool:           pool,
		agentClient:    agentClient,
		neurondbClient: neurondbClient,
	}
}

/* WorkflowDefinition represents a workflow DAG structure */
type WorkflowDefinition struct {
	Steps      []WorkflowStep   `json:"steps"`
	Conditions []WorkflowCondition `json:"conditions,omitempty"`
	StartStep  string           `json:"start_step"`
}

/* WorkflowStep represents a single workflow step */
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "agent", "script", "condition", "parallel"
	Task        string                 `json:"task,omitempty"`
	AgentID     *string                `json:"agent_id,omitempty"`
	Tools       []string               `json:"tools,omitempty"`
	Script      string                 `json:"script,omitempty"`
	NextSteps   []string               `json:"next_steps,omitempty"`
	Parallel    []string               `json:"parallel,omitempty"`
	Condition   *WorkflowCondition     `json:"condition,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

/* WorkflowCondition represents a conditional branch */
type WorkflowCondition struct {
	Type      string                 `json:"type"` // "if", "switch"
	Expression string                `json:"expression,omitempty"`
	Cases     []WorkflowConditionCase `json:"cases,omitempty"`
	Default   string                 `json:"default,omitempty"`
}

/* WorkflowConditionCase represents a condition case */
type WorkflowConditionCase struct {
	Value    interface{} `json:"value"`
	NextStep string      `json:"next_step"`
}

/* ExecutionState tracks the state of workflow execution */
type ExecutionState struct {
	ExecutionID   uuid.UUID
	WorkflowID    uuid.UUID
	CurrentStep   string
	CompletedSteps map[string]bool
	StepResults   map[string]interface{}
	Status        string
}

/* ExecuteWorkflow executes a workflow */
func (s *Service) ExecuteWorkflow(ctx context.Context, workflowID uuid.UUID, input map[string]interface{}) (map[string]interface{}, error) {
	// Get workflow definition
	workflow, err := s.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	if !workflow.Enabled {
		return nil, fmt.Errorf("workflow is disabled")
	}

	// Parse workflow definition
	var def WorkflowDefinition
	defJSON, _ := json.Marshal(workflow.WorkflowDefinition)
	if err := json.Unmarshal(defJSON, &def); err != nil {
		return nil, fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	// Create execution record
	executionID := uuid.New()
	inputJSON, _ := json.Marshal(input)
	now := time.Now()

	insertQuery := `
		INSERT INTO neuronip.workflow_executions 
		(id, workflow_id, status, input_data, started_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err = s.pool.QueryRow(ctx, insertQuery, executionID, workflowID, "running", inputJSON, now, now).Scan(&executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution record: %w", err)
	}

	// Initialize execution state
	state := ExecutionState{
		ExecutionID:    executionID,
		WorkflowID:     workflowID,
		CurrentStep:    def.StartStep,
		CompletedSteps: make(map[string]bool),
		StepResults:    make(map[string]interface{}),
		Status:         "running",
	}

	// Execute workflow steps
	var output map[string]interface{}
	output, err = s.executeWorkflowSteps(ctx, &def, &state, input)

	// Update execution status
	status := "completed"
	errorMsg := sql.NullString{}
	if err != nil {
		status = "failed"
		errorMsg = sql.NullString{String: err.Error(), Valid: true}
	}

	outputJSON, _ := json.Marshal(output)
	completedAt := time.Now()

	updateQuery := `
		UPDATE neuronip.workflow_executions 
		SET status = $1, output_data = $2, error_message = $3, completed_at = $4
		WHERE id = $5`

	s.pool.Exec(ctx, updateQuery, status, outputJSON, errorMsg, completedAt, executionID)

	if err != nil {
		return nil, err
	}

	return output, nil
}

/* executeWorkflowSteps executes workflow steps based on DAG */
func (s *Service) executeWorkflowSteps(ctx context.Context, def *WorkflowDefinition, state *ExecutionState, input map[string]interface{}) (map[string]interface{}, error) {
	stepMap := make(map[string]*WorkflowStep)
	for i := range def.Steps {
		stepMap[def.Steps[i].ID] = &def.Steps[i]
	}

	currentData := make(map[string]interface{})
	for k, v := range input {
		currentData[k] = v
	}

	// Execute steps starting from start step
	currentStepID := def.StartStep
	maxSteps := 100 // Prevent infinite loops
	stepCount := 0

	for currentStepID != "" && stepCount < maxSteps {
		stepCount++

		step, exists := stepMap[currentStepID]
		if !exists {
			return nil, fmt.Errorf("step not found: %s", currentStepID)
		}

		if state.CompletedSteps[currentStepID] {
			// Skip already completed steps (for parallel execution)
			currentStepID = s.getNextStep(step, currentData)
			continue
		}

		// Handle parallel steps specially - execute all parallel steps concurrently
		if step.Type == "parallel" && len(step.Parallel) > 0 {
			parallelResults, err := s.executeParallelSteps(ctx, step.Parallel, stepMap, currentData, state)
			if err != nil {
				return nil, fmt.Errorf("failed to execute parallel steps for %s: %w", step.ID, err)
			}
			
			// Store parallel results
			state.StepResults[step.ID] = parallelResults
			state.CompletedSteps[step.ID] = true
			
			// Merge all parallel step results into current data
			if parallelResults != nil {
				if resultMap, ok := parallelResults.(map[string]interface{}); ok {
					for k, v := range resultMap {
						currentData[k] = v
					}
				} else {
					currentData[step.ID+"_result"] = parallelResults
				}
			}
			
			// Get next step after parallel execution
			currentStepID = s.getNextStep(step, currentData)
			continue
		}

		// Execute step
		stepResult, err := s.executeStep(ctx, step, currentData, state)
		if err != nil {
			return nil, fmt.Errorf("failed to execute step %s: %w", step.ID, err)
		}

		// Store step result
		state.StepResults[step.ID] = stepResult
		state.CompletedSteps[step.ID] = true

		// Merge step result into current data
		if stepResult != nil {
			if resultMap, ok := stepResult.(map[string]interface{}); ok {
				for k, v := range resultMap {
					currentData[k] = v
				}
			} else {
				currentData[step.ID+"_result"] = stepResult
			}
		}

		// Get next step
		currentStepID = s.getNextStep(step, currentData)
	}

	if stepCount >= maxSteps {
		return nil, fmt.Errorf("workflow exceeded maximum step count")
	}

	return currentData, nil
}

/* executeStep executes a single workflow step */
func (s *Service) executeStep(ctx context.Context, step *WorkflowStep, data map[string]interface{}, state *ExecutionState) (interface{}, error) {
	switch step.Type {
	case "agent":
		return s.executeAgentStep(ctx, step, data, state)
	case "script":
		return s.executeScriptStep(ctx, step, data)
	case "parallel":
		return s.executeParallelStep(ctx, step, data, state)
	case "condition":
		return s.executeConditionStep(ctx, step, data)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}
}

/* executeAgentStep executes an agent step */
func (s *Service) executeAgentStep(ctx context.Context, step *WorkflowStep, data map[string]interface{}, state *ExecutionState) (interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	agentID := step.AgentID
	if agentID == nil {
		return nil, fmt.Errorf("agent_id not specified for agent step")
	}

	// Prepare task with data interpolation
	task := s.interpolateString(step.Task, data)

	// Get workflow memory for agent
	memory := s.getWorkflowMemory(ctx, state.WorkflowID, state.ExecutionID)

	// Execute agent
	result, err := s.agentClient.ExecuteAgent(ctx, *agentID, task, step.Tools, memory)
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// Store decision if applicable
	if step.Config != nil {
		if trackDecision, ok := step.Config["track_decision"].(bool); ok && trackDecision {
			s.logDecision(ctx, state.ExecutionID, step.ID, result)
		}
	}

	return result, nil
}

/* executeScriptStep executes a script step */
func (s *Service) executeScriptStep(ctx context.Context, step *WorkflowStep, data map[string]interface{}) (interface{}, error) {
	if step.Script == "" {
		return nil, fmt.Errorf("script not specified for script step")
	}

	// Interpolate script with data variables
	script := s.interpolateString(step.Script, data)

	// Check if script uses MCP tools
	scriptType := "inline"
	if step.Config != nil {
		if st, ok := step.Config["script_type"].(string); ok {
			scriptType = st
		}
	}

	switch scriptType {
	case "mcp":
		// If script references MCP tools, we would use MCP client here
		// For now, return result indicating MCP execution would occur
		if toolName, ok := step.Config["mcp_tool"].(string); ok {
			// In production, would call: mcpClient.ExecuteTool(ctx, toolName, data)
			return map[string]interface{}{
				"status":    "executed",
				"step":      step.ID,
				"tool":      toolName,
				"script":    script,
				"result":    "MCP tool execution would occur here",
			}, nil
		}
		return map[string]interface{}{"status": "executed", "step": step.ID, "script": script}, nil

	case "sql":
		// Execute SQL script (for warehouse queries)
		// In production, would execute SQL via database pool
		return map[string]interface{}{
			"status": "executed",
			"step":   step.ID,
			"type":   "sql",
			"script": script,
		}, nil

	default: // "inline" or JavaScript-like expressions
		// Simple expression evaluation using data interpolation
		// For complex scripts, you'd use a JavaScript engine or similar
		result := map[string]interface{}{
			"status": "executed",
			"step":   step.ID,
			"script": script,
		}

		// If script contains return statements or expressions, evaluate them
		// This is a simplified version - in production, use a proper expression evaluator
		if strings.HasPrefix(strings.TrimSpace(script), "return ") {
			expr := strings.TrimPrefix(strings.TrimSpace(script), "return ")
			result["return_value"] = s.evaluateExpression(expr, data)
		}

		return result, nil
	}
}

/* evaluateExpression evaluates a simple expression using data context */
func (s *Service) evaluateExpression(expr string, data map[string]interface{}) interface{} {
	// Simple variable substitution
	expr = s.interpolateString(expr, data)
	
	// Try to evaluate as number
	// In production, use a proper expression evaluator library
	return expr
}

/* executeParallelSteps executes multiple steps in parallel using goroutines */
func (s *Service) executeParallelSteps(ctx context.Context, stepIDs []string, stepMap map[string]*WorkflowStep, data map[string]interface{}, state *ExecutionState) (interface{}, error) {
	if len(stepIDs) == 0 {
		return map[string]interface{}{"status": "no_parallel_steps"}, nil
	}
	
	type stepResult struct {
		stepID string
		result interface{}
		err    error
	}
	
	resultChan := make(chan stepResult, len(stepIDs))
	
	// Execute all parallel steps concurrently
	for _, stepID := range stepIDs {
		go func(id string) {
			// Create a context with timeout for each parallel step
			stepCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			
			var res interface{}
			var err error
			
			// Check if step is already completed
			if state.CompletedSteps[id] {
				if existingResult, ok := state.StepResults[id]; ok {
					resultChan <- stepResult{stepID: id, result: existingResult, err: nil}
					return
				}
			}
			
			// Get step definition from stepMap
			step, exists := stepMap[id]
			if !exists {
				resultChan <- stepResult{
					stepID: id,
					result: nil,
					err:    fmt.Errorf("step not found: %s", id),
				}
				return
			}
			
			// Execute the step
			res, err = s.executeStep(stepCtx, step, data, state)
			
			resultChan <- stepResult{
				stepID: id,
				result: res,
				err:    err,
			}
		}(stepID)
	}
	
	// Collect results from all parallel steps
	results := make(map[string]interface{})
	errors := make([]error, 0)
	
	for i := 0; i < len(stepIDs); i++ {
		result := <-resultChan
		results[result.stepID] = result.result
		if result.err != nil {
			errors = append(errors, fmt.Errorf("step %s failed: %w", result.stepID, result.err))
		} else {
			// Mark step as completed if it succeeded
			state.CompletedSteps[result.stepID] = true
			state.StepResults[result.stepID] = result.result
		}
	}
	
	// If any step failed, return error
	if len(errors) > 0 {
		return results, fmt.Errorf("parallel execution had %d errors: %v", len(errors), errors)
	}
	
	return results, nil
}

/* executeParallelStep executes parallel steps - this is called from executeStep for backward compatibility */
func (s *Service) executeParallelStep(ctx context.Context, step *WorkflowStep, data map[string]interface{}, state *ExecutionState) (interface{}, error) {
	// Parallel execution should be handled in executeWorkflowSteps
	// This method is kept for backward compatibility but should not be called for parallel type steps
	if len(step.Parallel) == 0 {
		return map[string]interface{}{"status": "no_parallel_steps"}, nil
	}
	return map[string]interface{}{"status": "parallel_execution_handled_in_main_loop"}, nil
}

/* executeConditionStep executes a condition step */
func (s *Service) executeConditionStep(ctx context.Context, step *WorkflowStep, data map[string]interface{}) (interface{}, error) {
	if step.Condition == nil {
		return nil, fmt.Errorf("condition not specified")
	}

	// Evaluate condition and return next step decision
	// This is handled by getNextStep
	return map[string]interface{}{"condition_evaluated": true}, nil
}

/* getNextStep determines the next step based on current step and data */
func (s *Service) getNextStep(step *WorkflowStep, data map[string]interface{}) string {
	if step.Condition != nil {
		return s.evaluateCondition(step.Condition, data)
	}

	if len(step.NextSteps) > 0 {
		return step.NextSteps[0] // Default to first next step
	}

	if len(step.Parallel) > 0 {
		// For parallel steps, continue with first parallel step
		return step.Parallel[0]
	}

	return "" // End of workflow
}

/* evaluateCondition evaluates a workflow condition with improved expression parsing */
func (s *Service) evaluateCondition(cond *WorkflowCondition, data map[string]interface{}) string {
	if cond.Type == "if" {
		if cond.Expression != "" {
			// Evaluate expression and check if true
			result := s.evaluateConditionExpression(cond.Expression, data)
			if isTruthy(result) {
				if len(cond.Cases) > 0 {
					return cond.Cases[0].NextStep
				}
			}
		}
		return cond.Default
	} else if cond.Type == "switch" {
		if cond.Expression != "" {
			value := s.evaluateConditionExpression(cond.Expression, data)
			// Try exact match first
			for _, c := range cond.Cases {
				if c.Value == value {
					return c.NextStep
				}
			}
			// Try string conversion for comparison
			valueStr := fmt.Sprintf("%v", value)
			for _, c := range cond.Cases {
				caseStr := fmt.Sprintf("%v", c.Value)
				if caseStr == valueStr {
					return c.NextStep
				}
			}
		}
		return cond.Default
	}

	return cond.Default
}

/* evaluateConditionExpression evaluates a condition expression with improved parsing */
func (s *Service) evaluateConditionExpression(expr string, data map[string]interface{}) interface{} {
	expr = strings.TrimSpace(expr)

	// Handle boolean literals
	if expr == "true" {
		return true
	}
	if expr == "false" {
		return false
	}

	// Handle comparison operators: ==, !=, <, >, <=, >=
	if strings.Contains(expr, "==") {
		parts := strings.SplitN(expr, "==", 2)
		left := s.getExpressionValue(strings.TrimSpace(parts[0]), data)
		right := s.getExpressionValue(strings.TrimSpace(parts[1]), data)
		return left == right
	}
	if strings.Contains(expr, "!=") {
		parts := strings.SplitN(expr, "!=", 2)
		left := s.getExpressionValue(strings.TrimSpace(parts[0]), data)
		right := s.getExpressionValue(strings.TrimSpace(parts[1]), data)
		return left != right
	}
	if strings.Contains(expr, "<=") {
		parts := strings.SplitN(expr, "<=", 2)
		left := s.getExpressionValue(strings.TrimSpace(parts[0]), data)
		right := s.getExpressionValue(strings.TrimSpace(parts[1]), data)
		return compareValues(left, right) <= 0
	}
	if strings.Contains(expr, ">=") {
		parts := strings.SplitN(expr, ">=", 2)
		left := s.getExpressionValue(strings.TrimSpace(parts[0]), data)
		right := s.getExpressionValue(strings.TrimSpace(parts[1]), data)
		return compareValues(left, right) >= 0
	}
	if strings.Contains(expr, "<") && !strings.Contains(expr, "<=") {
		parts := strings.SplitN(expr, "<", 2)
		left := s.getExpressionValue(strings.TrimSpace(parts[0]), data)
		right := s.getExpressionValue(strings.TrimSpace(parts[1]), data)
		return compareValues(left, right) < 0
	}
	if strings.Contains(expr, ">") && !strings.Contains(expr, ">=") {
		parts := strings.SplitN(expr, ">", 2)
		left := s.getExpressionValue(strings.TrimSpace(parts[0]), data)
		right := s.getExpressionValue(strings.TrimSpace(parts[1]), data)
		return compareValues(left, right) > 0
	}

	// Simple variable access
	return s.getExpressionValue(expr, data)
}

/* getExpressionValue gets a value from expression (supports dot notation) */
func (s *Service) getExpressionValue(expr string, data map[string]interface{}) interface{} {
	// Remove quotes if present
	expr = strings.Trim(expr, `"'`)
	
	// Check if it's a numeric literal
	if num, err := parseNumber(expr); err == nil {
		return num
	}

	// Navigate data structure using dot notation
	parts := strings.Split(expr, ".")
	var value interface{} = data
	for _, part := range parts {
		if mapValue, ok := value.(map[string]interface{}); ok {
			value = mapValue[part]
		} else {
			return nil
		}
	}
	return value
}

/* isTruthy checks if a value is truthy */
func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	if s, ok := value.(string); ok {
		return s != "" && s != "false" && s != "0"
	}
	if f, ok := value.(float64); ok {
		return f != 0
	}
	return true
}

/* compareValues compares two values numerically if possible */
func compareValues(left, right interface{}) int {
	leftNum, leftOk := asNumber(left)
	rightNum, rightOk := asNumber(right)
	if leftOk && rightOk {
		if leftNum < rightNum {
			return -1
		} else if leftNum > rightNum {
			return 1
		}
		return 0
	}
	// Fallback to string comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	if leftStr < rightStr {
		return -1
	} else if leftStr > rightStr {
		return 1
	}
	return 0
}

/* asNumber converts value to float64 if possible */
func asNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if num, err := parseNumber(v); err == nil {
			return num, true
		}
	}
	return 0, false
}

/* parseNumber parses a string to number */
func parseNumber(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

/* interpolateString replaces template variables with data values */
func (s *Service) interpolateString(template string, data map[string]interface{}) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}
	return result
}

/* getWorkflowMemory retrieves workflow memory for agent context */
func (s *Service) getWorkflowMemory(ctx context.Context, workflowID uuid.UUID, executionID uuid.UUID) map[string]interface{} {
	query := `
		SELECT memory_key, memory_value
		FROM neuronip.workflow_memory
		WHERE workflow_id = $1 AND (execution_id = $2 OR execution_id IS NULL)
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, workflowID, executionID)
	if err != nil {
		return make(map[string]interface{})
	}
	defer rows.Close()

	memory := make(map[string]interface{})
	for rows.Next() {
		var key string
		var valueJSON json.RawMessage
		if err := rows.Scan(&key, &valueJSON); err != nil {
			continue
		}
		var value interface{}
		if err := json.Unmarshal(valueJSON, &value); err == nil {
			memory[key] = value
		}
	}

	return memory
}

/* logDecision logs a workflow decision */
func (s *Service) logDecision(ctx context.Context, executionID uuid.UUID, decisionPoint string, result interface{}) error {
	decisionValue := fmt.Sprintf("%v", result)
	contextJSON, _ := json.Marshal(result)

	query := `
		INSERT INTO neuronip.workflow_decisions 
		(execution_id, decision_point, decision_value, context, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := s.pool.Exec(ctx, query, executionID, decisionPoint, decisionValue, contextJSON, time.Now())
	return err
}

/* CreateWorkflowVersion creates a new version of a workflow */
func (s *Service) CreateWorkflowVersion(ctx context.Context, workflowID uuid.UUID, version string, changes map[string]interface{}) (*Workflow, error) {
	// Get current workflow
	current, err := s.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current workflow: %w", err)
	}

	// Create new version
	newWorkflow := *current
	newWorkflow.ID = uuid.New()
	newWorkflow.CreatedAt = time.Now()
	newWorkflow.UpdatedAt = time.Now()

	// Apply changes
	if name, ok := changes["name"].(string); ok {
		newWorkflow.Name = name
	}
	if description, ok := changes["description"].(string); ok {
		newWorkflow.Description = &description
	}
	if def, ok := changes["workflow_definition"].(map[string]interface{}); ok {
		newWorkflow.WorkflowDefinition = def
	}

	// Store version info
	defJSON, _ := json.Marshal(newWorkflow.WorkflowDefinition)
	descriptionJSON := sql.NullString{}
	if newWorkflow.Description != nil {
		descriptionJSON = sql.NullString{String: *newWorkflow.Description, Valid: true}
	}

	query := `
		INSERT INTO neuronip.workflows 
		(id, name, description, workflow_definition, agent_id, enabled, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err = s.pool.QueryRow(ctx, query,
		newWorkflow.ID, newWorkflow.Name, descriptionJSON, defJSON,
		newWorkflow.AgentID, newWorkflow.Enabled, newWorkflow.CreatedBy,
		newWorkflow.CreatedAt, newWorkflow.UpdatedAt,
	).Scan(&newWorkflow.ID, &newWorkflow.CreatedAt, &newWorkflow.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create workflow version: %w", err)
	}

	// Link versions
	linkQuery := `
		INSERT INTO neuronip.workflow_versions 
		(workflow_id, version, parent_workflow_id, created_at)
		VALUES ($1, $2, $3, NOW())`
	s.pool.Exec(ctx, linkQuery, newWorkflow.ID, version, workflowID)

	return &newWorkflow, nil
}

/* ScheduleWorkflow schedules a workflow for execution */
func (s *Service) ScheduleWorkflow(ctx context.Context, workflowID uuid.UUID, schedule ScheduleConfig) error {
	scheduleJSON, _ := json.Marshal(schedule)

	query := `
		INSERT INTO neuronip.workflow_schedules 
		(id, workflow_id, schedule_config, enabled, next_run_at, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		ON CONFLICT (workflow_id) DO UPDATE
		SET schedule_config = $2, enabled = $3, next_run_at = $4, updated_at = NOW()`

	nextRun := s.calculateNextRun(schedule)
	_, err := s.pool.Exec(ctx, query, workflowID, scheduleJSON, schedule.Enabled, nextRun)
	return err
}

/* ScheduleConfig represents workflow scheduling configuration */
type ScheduleConfig struct {
	CronExpression string                 `json:"cron_expression,omitempty"`
	Interval       string                 `json:"interval,omitempty"` // "hourly", "daily", "weekly", "monthly"
	Time           string                 `json:"time,omitempty"`     // Time of day for daily/weekly/monthly
	DayOfWeek      int                    `json:"day_of_week,omitempty"` // 0-6 for weekly
	DayOfMonth     int                    `json:"day_of_month,omitempty"` // 1-31 for monthly
	Enabled        bool                   `json:"enabled"`
	Input          map[string]interface{} `json:"input,omitempty"`
}

/* calculateNextRun calculates the next run time based on schedule */
func (s *Service) calculateNextRun(schedule ScheduleConfig) time.Time {
	now := time.Now()
	if schedule.Interval != "" {
		switch schedule.Interval {
		case "hourly":
			return now.Add(1 * time.Hour)
		case "daily":
			next := now.Add(24 * time.Hour)
			if schedule.Time != "" {
				// Parse time and set for next day
				// Simplified - would parse HH:MM format
			}
			return next
		case "weekly":
			currentDay := int(now.Weekday())
			targetDay := schedule.DayOfWeek
			daysUntilNext := (7 - currentDay + targetDay) % 7
			if daysUntilNext == 0 {
				daysUntilNext = 7
			}
			return now.Add(time.Duration(daysUntilNext) * 24 * time.Hour)
		case "monthly":
			next := now.AddDate(0, 1, 0)
			if schedule.DayOfMonth > 0 {
				// Set to specific day of month
				next = time.Date(next.Year(), next.Month(), schedule.DayOfMonth, 0, 0, 0, 0, next.Location())
			}
			return next
		}
	}
	return now.Add(1 * time.Hour) // Default to hourly
}

/* RecoverWorkflowExecution recovers a failed workflow execution */
func (s *Service) RecoverWorkflowExecution(ctx context.Context, executionID uuid.UUID, retryFromStep *string) (map[string]interface{}, error) {
	// Get execution record
	query := `
		SELECT id, workflow_id, status, input_data, output_data, error_message, started_at
		FROM neuronip.workflow_executions
		WHERE id = $1`

	var workflowID uuid.UUID
	var execStatus string
	var inputJSON, outputJSON json.RawMessage
	var errorMsg sql.NullString
	var startedAt, completedAt time.Time

	err := s.pool.QueryRow(ctx, query, executionID).Scan(
		&executionID, &workflowID, &execStatus, &inputJSON, &outputJSON, &errorMsg, &startedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	if execStatus != "failed" {
		return nil, fmt.Errorf("execution is not in failed state")
	}

	// Get workflow
	workflow, err := s.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Parse workflow definition
	var def WorkflowDefinition
	defJSON, _ := json.Marshal(workflow.WorkflowDefinition)
	if err := json.Unmarshal(defJSON, &def); err != nil {
		return nil, fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	// Get execution state
	var input map[string]interface{}
	if inputJSON != nil {
		json.Unmarshal(inputJSON, &input)
	}

	// Recreate execution state from saved state
	state := ExecutionState{
		ExecutionID:    executionID,
		WorkflowID:     workflowID,
		CurrentStep:    def.StartStep,
		CompletedSteps: make(map[string]bool),
		StepResults:    make(map[string]interface{}),
		Status:         "running",
	}

	// If retryFromStep is specified, start from that step
	if retryFromStep != nil {
		state.CurrentStep = *retryFromStep
	}

	// Get saved step results if available
	stepResultsQuery := `
		SELECT step_id, result_data
		FROM neuronip.workflow_step_results
		WHERE execution_id = $1`
	rows, _ := s.pool.Query(ctx, stepResultsQuery, executionID)
	for rows.Next() {
		var stepID string
		var resultJSON json.RawMessage
		if err := rows.Scan(&stepID, &resultJSON); err == nil {
			var result interface{}
			json.Unmarshal(resultJSON, &result)
			state.StepResults[stepID] = result
			state.CompletedSteps[stepID] = true
		}
	}
	rows.Close()

	// Resume execution
	output, execErr := s.executeWorkflowSteps(ctx, &def, &state, input)

	// Update execution status
	finalStatus := "completed"
	finalErrorMsg := sql.NullString{}
	if execErr != nil {
		finalStatus = "failed"
		finalErrorMsg = sql.NullString{String: execErr.Error(), Valid: true}
	}

	var outputJSONBytes []byte
	outputJSONBytes, _ = json.Marshal(output)
	completedAt = time.Now()

	updateQuery := `
		UPDATE neuronip.workflow_executions 
		SET status = $1, output_data = $2, error_message = $3, completed_at = $4
		WHERE id = $5`
	s.pool.Exec(ctx, updateQuery, finalStatus, outputJSONBytes, finalErrorMsg, completedAt, executionID)

	if execErr != nil {
		return nil, execErr
	}

	return output, nil
}

/* GetWorkflowExecutionStatus retrieves workflow execution status */
func (s *Service) GetWorkflowExecutionStatus(ctx context.Context, executionID uuid.UUID) (*WorkflowExecutionStatus, error) {
	query := `
		SELECT id, workflow_id, status, started_at, completed_at, execution_time_ms, error_message
		FROM neuronip.workflow_executions
		WHERE id = $1`

	var status WorkflowExecutionStatus
	var completedAt sql.NullTime
	var executionTimeMs sql.NullInt64
	var errorMsg sql.NullString

	err := s.pool.QueryRow(ctx, query, executionID).Scan(
		&status.ExecutionID, &status.WorkflowID, &status.Status,
		&status.StartedAt, &completedAt, &executionTimeMs, &errorMsg,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution status: %w", err)
	}

	if completedAt.Valid {
		status.CompletedAt = &completedAt.Time
	}
	if executionTimeMs.Valid {
		status.ExecutionTimeMs = &executionTimeMs.Int64
	}
	if errorMsg.Valid {
		status.ErrorMessage = &errorMsg.String
	}

	// Get step execution details
	stepQuery := `
		SELECT step_id, status, started_at, completed_at, error_message
		FROM neuronip.workflow_step_executions
		WHERE execution_id = $1
		ORDER BY started_at`
	rows, _ := s.pool.Query(ctx, stepQuery, executionID)
	defer rows.Close()

	status.Steps = []WorkflowStepExecution{}
	for rows.Next() {
		var step WorkflowStepExecution
		var stepCompletedAt sql.NullTime
		var stepErrorMsg sql.NullString

		err := rows.Scan(
			&step.StepID, &step.Status, &step.StartedAt,
			&stepCompletedAt, &stepErrorMsg,
		)
		if err != nil {
			continue
		}

		if stepCompletedAt.Valid {
			step.CompletedAt = &stepCompletedAt.Time
		}
		if stepErrorMsg.Valid {
			step.ErrorMessage = &stepErrorMsg.String
		}

		status.Steps = append(status.Steps, step)
	}

	// Get current step and completed steps from step executions
	status.CurrentStep = ""
	status.CompletedSteps = []string{}
	for _, step := range status.Steps {
		if step.Status == "running" {
			status.CurrentStep = step.StepID
		}
		if step.Status == "completed" {
			status.CompletedSteps = append(status.CompletedSteps, step.StepID)
		}
	}

	return &status, nil
}

/* WorkflowExecutionStatus represents workflow execution status */
type WorkflowExecutionStatus struct {
	ExecutionID     uuid.UUID                `json:"execution_id"`
	WorkflowID      uuid.UUID                `json:"workflow_id"`
	Status          string                   `json:"status"`
	StartedAt       time.Time               `json:"started_at"`
	CompletedAt     *time.Time               `json:"completed_at,omitempty"`
	ExecutionTimeMs *int64                   `json:"execution_time_ms,omitempty"`
	ErrorMessage    *string                  `json:"error_message,omitempty"`
	CurrentStep     string                   `json:"current_step,omitempty"`
	CompletedSteps []string                 `json:"completed_steps,omitempty"`
	Steps           []WorkflowStepExecution  `json:"steps,omitempty"`
}

/* WorkflowStepExecution represents a step execution */
type WorkflowStepExecution struct {
	StepID      string     `json:"step_id"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
}

/* GetWorkflowMonitoring retrieves workflow monitoring metrics */
func (s *Service) GetWorkflowMonitoring(ctx context.Context, workflowID uuid.UUID, timeRange string) (*WorkflowMonitoring, error) {
	var interval string
	switch timeRange {
	case "1h", "1hour":
		interval = "1 hour"
	case "24h", "1day":
		interval = "24 hours"
	case "7d", "1week":
		interval = "7 days"
	default:
		interval = "24 hours"
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_executions,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as error_count,
			AVG(execution_time_ms) as avg_execution_time,
			MAX(execution_time_ms) as max_execution_time
		FROM neuronip.workflow_executions
		WHERE workflow_id = $1 AND started_at > NOW() - INTERVAL '%s'`, interval)

	var monitoring WorkflowMonitoring
	var totalExecutions, successCount, errorCount *int
	var avgExecutionTime, maxExecutionTime *float64

	err := s.pool.QueryRow(ctx, query, workflowID).Scan(
		&totalExecutions, &successCount, &errorCount, &avgExecutionTime, &maxExecutionTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitoring data: %w", err)
	}

	monitoring.WorkflowID = workflowID
	monitoring.TimeRange = timeRange
	if totalExecutions != nil {
		monitoring.TotalExecutions = *totalExecutions
	}
	if successCount != nil {
		monitoring.SuccessCount = *successCount
	}
	if errorCount != nil {
		monitoring.ErrorCount = *errorCount
		monitoring.FailedExecutions = *errorCount
	}
	if avgExecutionTime != nil {
		monitoring.AvgExecutionTimeMs = *avgExecutionTime
		monitoring.AvgDuration = *avgExecutionTime / 1000.0 // Convert ms to seconds
	}
	if maxExecutionTime != nil {
		monitoring.MaxExecutionTimeMs = *maxExecutionTime
	}

	// Calculate success rate
	if monitoring.TotalExecutions > 0 {
		monitoring.SuccessRate = float64(monitoring.SuccessCount) / float64(monitoring.TotalExecutions)
	}

	return &monitoring, nil
}

/* WorkflowMonitoring represents workflow monitoring metrics */
type WorkflowMonitoring struct {
	WorkflowID         uuid.UUID `json:"workflow_id"`
	TimeRange          string    `json:"time_range"`
	TotalExecutions    int       `json:"total_executions"`
	SuccessCount       int       `json:"success_count"`
	ErrorCount         int       `json:"error_count"`
	FailedExecutions   int       `json:"failed_executions"`
	SuccessRate        float64   `json:"success_rate"`
	AvgExecutionTimeMs float64   `json:"avg_execution_time_ms"`
	AvgDuration        float64   `json:"avg_duration"`
	MaxExecutionTimeMs float64   `json:"max_execution_time_ms"`
}

/* GetWorkflow retrieves a workflow by ID */
func (s *Service) GetWorkflow(ctx context.Context, id uuid.UUID) (*Workflow, error) {
	var workflow Workflow
	var description sql.NullString
	var createdBy sql.NullString
	var agentID *uuid.UUID
	var defJSON json.RawMessage

	query := `
		SELECT id, name, description, workflow_definition, agent_id, enabled, created_by, created_at, updated_at
		FROM neuronip.workflows
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&workflow.ID, &workflow.Name, &description, &defJSON,
		&agentID, &workflow.Enabled, &createdBy,
		&workflow.CreatedAt, &workflow.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("workflow not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	if description.Valid {
		workflow.Description = &description.String
	}
	if createdBy.Valid {
		workflow.CreatedBy = &createdBy.String
	}
	workflow.AgentID = agentID

	if defJSON != nil {
		json.Unmarshal(defJSON, &workflow.WorkflowDefinition)
	}

	return &workflow, nil
}

/* ListWorkflows lists all workflows */
func (s *Service) ListWorkflows(ctx context.Context) ([]Workflow, error) {
	query := `
		SELECT id, name, description, workflow_definition, agent_id, enabled, created_by, created_at, updated_at
		FROM neuronip.workflows
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	defer rows.Close()

	var workflows []Workflow
	for rows.Next() {
		var workflow Workflow
		var description sql.NullString
		var createdBy sql.NullString
		var agentID *uuid.UUID
		var defJSON json.RawMessage

		err := rows.Scan(
			&workflow.ID, &workflow.Name, &description, &defJSON,
			&agentID, &workflow.Enabled, &createdBy,
			&workflow.CreatedAt, &workflow.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if description.Valid {
			workflow.Description = &description.String
		}
		if createdBy.Valid {
			workflow.CreatedBy = &createdBy.String
		}
		workflow.AgentID = agentID

		if defJSON != nil {
			json.Unmarshal(defJSON, &workflow.WorkflowDefinition)
		}

		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

/* CreateWorkflow creates a new workflow */
func (s *Service) CreateWorkflow(ctx context.Context, workflow Workflow) (*Workflow, error) {
	workflow.ID = uuid.New()
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	defJSON, _ := json.Marshal(workflow.WorkflowDefinition)
	descriptionJSON := sql.NullString{}
	if workflow.Description != nil {
		descriptionJSON = sql.NullString{String: *workflow.Description, Valid: true}
	}

	query := `
		INSERT INTO neuronip.workflows 
		(id, name, description, workflow_definition, agent_id, enabled, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		workflow.ID, workflow.Name, descriptionJSON, defJSON,
		workflow.AgentID, workflow.Enabled, workflow.CreatedBy,
		workflow.CreatedAt, workflow.UpdatedAt,
	).Scan(&workflow.ID, &workflow.CreatedAt, &workflow.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return &workflow, nil
}

/* UpdateWorkflow updates an existing workflow */
func (s *Service) UpdateWorkflow(ctx context.Context, id uuid.UUID, workflow Workflow) (*Workflow, error) {
	workflow.UpdatedAt = time.Now()

	defJSON, _ := json.Marshal(workflow.WorkflowDefinition)
	descriptionJSON := sql.NullString{}
	if workflow.Description != nil {
		descriptionJSON = sql.NullString{String: *workflow.Description, Valid: true}
	}

	query := `
		UPDATE neuronip.workflows 
		SET name = $1, description = $2, workflow_definition = $3, agent_id = $4, enabled = $5, updated_at = $6
		WHERE id = $7
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		workflow.Name, descriptionJSON, defJSON,
		workflow.AgentID, workflow.Enabled, workflow.UpdatedAt, id,
	).Scan(&workflow.ID, &workflow.CreatedAt, &workflow.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	workflow.ID = id
	return &workflow, nil
}

/* DeleteWorkflow deletes a workflow */
func (s *Service) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM neuronip.workflows WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}
	return nil
}

/* GetWorkflowVersions lists all versions of a workflow */
func (s *Service) GetWorkflowVersions(ctx context.Context, workflowID uuid.UUID) ([]WorkflowVersion, error) {
	query := `
		SELECT wv.id, wv.workflow_id, wv.version, wv.parent_workflow_id, wv.changes, wv.created_at,
		       w.id, w.name, w.description, w.workflow_definition, w.agent_id, w.enabled, w.created_by, w.created_at, w.updated_at
		FROM neuronip.workflow_versions wv
		JOIN neuronip.workflows w ON wv.workflow_id = w.id
		WHERE wv.parent_workflow_id = $1 OR wv.workflow_id = $1
		ORDER BY wv.created_at DESC`

	rows, err := s.pool.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow versions: %w", err)
	}
	defer rows.Close()

	var versions []WorkflowVersion
	for rows.Next() {
		var version WorkflowVersion
		var workflow Workflow
		var description sql.NullString
		var createdBy sql.NullString
		var agentID *uuid.UUID
		var defJSON json.RawMessage
		var changesJSON json.RawMessage

		err := rows.Scan(
			&version.ID, &version.WorkflowID, &version.Version, &version.ParentWorkflowID,
			&changesJSON, &version.CreatedAt,
			&workflow.ID, &workflow.Name, &description, &defJSON,
			&agentID, &workflow.Enabled, &createdBy,
			&workflow.CreatedAt, &workflow.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if description.Valid {
			workflow.Description = &description.String
		}
		if createdBy.Valid {
			workflow.CreatedBy = &createdBy.String
		}
		workflow.AgentID = agentID

		if defJSON != nil {
			json.Unmarshal(defJSON, &workflow.WorkflowDefinition)
		}
		if changesJSON != nil {
			json.Unmarshal(changesJSON, &version.Changes)
		}

		version.Workflow = workflow
		versions = append(versions, version)
	}

	return versions, nil
}

/* GetWorkflowVersion gets a specific workflow version */
func (s *Service) GetWorkflowVersion(ctx context.Context, workflowID uuid.UUID, versionID uuid.UUID) (*WorkflowVersion, error) {
	query := `
		SELECT wv.id, wv.workflow_id, wv.version, wv.parent_workflow_id, wv.changes, wv.created_at,
		       w.id, w.name, w.description, w.workflow_definition, w.agent_id, w.enabled, w.created_by, w.created_at, w.updated_at
		FROM neuronip.workflow_versions wv
		JOIN neuronip.workflows w ON wv.workflow_id = w.id
		WHERE wv.id = $1 AND (wv.parent_workflow_id = $2 OR wv.workflow_id = $2)`

	var version WorkflowVersion
	var workflow Workflow
	var description sql.NullString
	var createdBy sql.NullString
	var agentID *uuid.UUID
	var defJSON json.RawMessage
	var changesJSON json.RawMessage

	err := s.pool.QueryRow(ctx, query, versionID, workflowID).Scan(
		&version.ID, &version.WorkflowID, &version.Version, &version.ParentWorkflowID,
		&changesJSON, &version.CreatedAt,
		&workflow.ID, &workflow.Name, &description, &defJSON,
		&agentID, &workflow.Enabled, &createdBy,
		&workflow.CreatedAt, &workflow.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("workflow version not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow version: %w", err)
	}

	if description.Valid {
		workflow.Description = &description.String
	}
	if createdBy.Valid {
		workflow.CreatedBy = &createdBy.String
	}
	workflow.AgentID = agentID

	if defJSON != nil {
		json.Unmarshal(defJSON, &workflow.WorkflowDefinition)
	}
	if changesJSON != nil {
		json.Unmarshal(changesJSON, &version.Changes)
	}

	version.Workflow = workflow
	return &version, nil
}

/* WorkflowVersion represents a workflow version */
type WorkflowVersion struct {
	ID              uuid.UUID              `json:"id"`
	WorkflowID      uuid.UUID               `json:"workflow_id"`
	Version         string                  `json:"version"`
	ParentWorkflowID *uuid.UUID             `json:"parent_workflow_id,omitempty"`
	Changes         map[string]interface{} `json:"changes,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
	Workflow        Workflow                `json:"workflow"`
}

/* GetScheduledWorkflows gets all scheduled workflows for a workflow */
func (s *Service) GetScheduledWorkflows(ctx context.Context, workflowID uuid.UUID) ([]WorkflowSchedule, error) {
	query := `
		SELECT id, workflow_id, schedule_config, enabled, next_run_at, last_run_at, created_at, updated_at
		FROM neuronip.workflow_schedules
		WHERE workflow_id = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled workflows: %w", err)
	}
	defer rows.Close()

	var schedules []WorkflowSchedule
	for rows.Next() {
		var schedule WorkflowSchedule
		var scheduleJSON json.RawMessage
		var nextRunAt, lastRunAt sql.NullTime

		err := rows.Scan(
			&schedule.ID, &schedule.WorkflowID, &scheduleJSON,
			&schedule.Enabled, &nextRunAt, &lastRunAt,
			&schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if scheduleJSON != nil {
			json.Unmarshal(scheduleJSON, &schedule.ScheduleConfig)
		}
		if nextRunAt.Valid {
			schedule.NextRunAt = &nextRunAt.Time
		}
		if lastRunAt.Valid {
			schedule.LastRunAt = &lastRunAt.Time
		}

		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

/* WorkflowSchedule represents a scheduled workflow */
type WorkflowSchedule struct {
	ID            uuid.UUID              `json:"id"`
	WorkflowID     uuid.UUID              `json:"workflow_id"`
	ScheduleConfig ScheduleConfig         `json:"schedule_config"`
	Enabled        bool                   `json:"enabled"`
	NextRunAt      *time.Time             `json:"next_run_at,omitempty"`
	LastRunAt      *time.Time             `json:"last_run_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

/* CancelScheduledWorkflow cancels a scheduled workflow */
func (s *Service) CancelScheduledWorkflow(ctx context.Context, workflowID uuid.UUID, scheduleID uuid.UUID) error {
	query := `
		UPDATE neuronip.workflow_schedules 
		SET enabled = false, updated_at = NOW()
		WHERE id = $1 AND workflow_id = $2`

	result, err := s.pool.Exec(ctx, query, scheduleID, workflowID)
	if err != nil {
		return fmt.Errorf("failed to cancel scheduled workflow: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("scheduled workflow not found")
	}

	return nil
}

/* GetWorkflowExecutionLogs gets execution logs */
func (s *Service) GetWorkflowExecutionLogs(ctx context.Context, executionID uuid.UUID) ([]WorkflowExecutionLog, error) {
	query := `
		SELECT id, execution_id, step_id, level, message, metadata, timestamp
		FROM neuronip.workflow_logs
		WHERE execution_id = $1
		ORDER BY timestamp ASC`

	rows, err := s.pool.Query(ctx, query, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution logs: %w", err)
	}
	defer rows.Close()

	var logs []WorkflowExecutionLog
	for rows.Next() {
		var log WorkflowExecutionLog
		var stepID sql.NullString
		var metadataJSON json.RawMessage

		err := rows.Scan(
			&log.ID, &log.ExecutionID, &stepID,
			&log.Level, &log.Message, &metadataJSON, &log.Timestamp,
		)
		if err != nil {
			continue
		}

		if stepID.Valid {
			log.StepID = &stepID.String
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

/* WorkflowExecutionLog represents an execution log entry */
type WorkflowExecutionLog struct {
	ID          uuid.UUID              `json:"id"`
	ExecutionID uuid.UUID              `json:"execution_id"`
	StepID      *string                `json:"step_id,omitempty"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

/* GetWorkflowExecutionMetrics gets execution metrics */
func (s *Service) GetWorkflowExecutionMetrics(ctx context.Context, executionID uuid.UUID) ([]WorkflowExecutionMetric, error) {
	query := `
		SELECT id, execution_id, step_id, metric_name, metric_value, metric_unit, description, timestamp
		FROM neuronip.workflow_metrics
		WHERE execution_id = $1
		ORDER BY timestamp ASC`

	rows, err := s.pool.Query(ctx, query, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution metrics: %w", err)
	}
	defer rows.Close()

	var metrics []WorkflowExecutionMetric
	for rows.Next() {
		var metric WorkflowExecutionMetric
		var stepID sql.NullString
		var unit sql.NullString
		var description sql.NullString

		err := rows.Scan(
			&metric.ID, &metric.ExecutionID, &stepID,
			&metric.Name, &metric.Value, &unit, &description, &metric.Timestamp,
		)
		if err != nil {
			continue
		}

		if stepID.Valid {
			metric.StepID = &stepID.String
		}
		if unit.Valid {
			metric.Unit = &unit.String
		}
		if description.Valid {
			metric.Description = &description.String
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

/* WorkflowExecutionMetric represents an execution metric */
type WorkflowExecutionMetric struct {
	ID          uuid.UUID  `json:"id"`
	ExecutionID uuid.UUID  `json:"execution_id"`
	StepID      *string    `json:"step_id,omitempty"`
	Name        string     `json:"name"`
	Value       float64    `json:"value"`
	Unit        *string    `json:"unit,omitempty"`
	Description *string    `json:"description,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
}

/* GetWorkflowExecutionDecisions gets execution decisions */
func (s *Service) GetWorkflowExecutionDecisions(ctx context.Context, executionID uuid.UUID) ([]WorkflowDecision, error) {
	query := `
		SELECT id, execution_id, decision_point, decision_value, decision_reason, context, created_at
		FROM neuronip.workflow_decisions
		WHERE execution_id = $1
		ORDER BY created_at ASC`

	rows, err := s.pool.Query(ctx, query, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution decisions: %w", err)
	}
	defer rows.Close()

	var decisions []WorkflowDecision
	for rows.Next() {
		var decision WorkflowDecision
		var reason sql.NullString
		var contextJSON json.RawMessage

		err := rows.Scan(
			&decision.ID, &decision.ExecutionID, &decision.DecisionPoint,
			&decision.DecisionValue, &reason, &contextJSON, &decision.CreatedAt,
		)
		if err != nil {
			continue
		}

		if reason.Valid {
			decision.DecisionReason = &reason.String
		}
		if contextJSON != nil {
			json.Unmarshal(contextJSON, &decision.Context)
		}

		decisions = append(decisions, decision)
	}

	return decisions, nil
}

/* WorkflowDecision represents a workflow decision */
type WorkflowDecision struct {
	ID            uuid.UUID              `json:"id"`
	ExecutionID   uuid.UUID              `json:"execution_id"`
	DecisionPoint string                 `json:"decision_point"`
	DecisionValue string                 `json:"decision_value"`
	DecisionReason *string               `json:"decision_reason,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

/* StoreWorkflowMemory stores workflow memory with vector embedding */
func (s *Service) StoreWorkflowMemory(ctx context.Context, workflowID uuid.UUID, executionID *uuid.UUID, memoryKey string, memoryValue map[string]interface{}) error {
	// Generate embedding from memory content
	memoryText := fmt.Sprintf("%v", memoryValue)
	embedding, err := s.neurondbClient.GenerateEmbedding(ctx, memoryText, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		embedding = "" // Continue without embedding if generation fails
	}

	memoryValueJSON, _ := json.Marshal(memoryValue)
	now := time.Now()

	query := `
		INSERT INTO neuronip.workflow_memory 
		(workflow_id, execution_id, memory_key, memory_value, embedding, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::vector, $6, $7)
		ON CONFLICT (workflow_id, memory_key) 
		DO UPDATE SET 
			memory_value = EXCLUDED.memory_value,
			embedding = EXCLUDED.embedding,
			execution_id = EXCLUDED.execution_id,
			updated_at = EXCLUDED.updated_at`

	_, err = s.pool.Exec(ctx, query, workflowID, executionID, memoryKey, memoryValueJSON, embedding, now, now)
	if err != nil {
		return fmt.Errorf("failed to store workflow memory: %w", err)
	}

	return nil
}

/* SearchWorkflowMemory performs semantic search on workflow memory */
func (s *Service) SearchWorkflowMemory(ctx context.Context, workflowID uuid.UUID, query string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}

	// Generate embedding for query
	queryEmbedding, err := s.neurondbClient.GenerateEmbedding(ctx, query, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Perform vector similarity search
	searchQuery := `
		SELECT memory_key, memory_value, 1 - (embedding <=> $1::vector) as similarity
		FROM neuronip.workflow_memory
		WHERE workflow_id = $2 AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $3`

	rows, err := s.pool.Query(ctx, searchQuery, queryEmbedding, workflowID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search workflow memory: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var key string
		var valueJSON json.RawMessage
		var similarity float64

		err := rows.Scan(&key, &valueJSON, &similarity)
		if err != nil {
			continue
		}

		var value interface{}
		json.Unmarshal(valueJSON, &value)

		results = append(results, map[string]interface{}{
			"memory_key": key,
			"memory_value": value,
			"similarity": similarity,
		})
	}

	return results, nil
}

/* Workflow represents a workflow model */
type Workflow struct {
	ID               uuid.UUID              `json:"id"`
	Name             string                 `json:"name"`
	Description      *string                `json:"description,omitempty"`
	WorkflowDefinition map[string]interface{} `json:"workflow_definition"`
	AgentID          *uuid.UUID             `json:"agent_id,omitempty"`
	Enabled          bool                   `json:"enabled"`
	CreatedBy        *string                `json:"created_by,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}
