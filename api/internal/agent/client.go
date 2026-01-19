package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

/* Client provides NeuronAgent integration */
type Client struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

/* NewClient creates a new NeuronAgent client */
func NewClient(endpoint string, apiKey string) *Client {
	return &Client{
		endpoint: endpoint,
		apiKey:   apiKey,
		client:   &http.Client{},
	}
}

/* CreateAgent creates an agent via NeuronAgent API */
func (c *Client) CreateAgent(ctx context.Context, name string, config map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"name":   name,
		"config": config,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/v1/agents", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("agent API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

/* ExecuteAgent executes an agent task via NeuronAgent API */
func (c *Client) ExecuteAgent(ctx context.Context, agentID string, task string, tools []string, memory map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"agent_id": agentID,
		"task":     task,
		"tools":    tools,
		"memory":   memory,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/v1/agents/"+agentID+"/execute", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

/* GetAgentMemory retrieves agent memory via NeuronAgent API */
func (c *Client) GetAgentMemory(ctx context.Context, agentID string, memoryKey string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint+"/api/v1/agents/"+agentID+"/memory/"+memoryKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

/* SetAgentMemory stores agent memory via NeuronAgent API */
func (c *Client) SetAgentMemory(ctx context.Context, agentID string, memoryKey string, memoryValue map[string]interface{}) error {
	reqBody := map[string]interface{}{
		"key":   memoryKey,
		"value": memoryValue,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/v1/agents/"+agentID+"/memory", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("agent API returned status %d", resp.StatusCode)
	}

	return nil
}

/* ConvertNLToSQL converts natural language to SQL via NeuronAgent API */
func (c *Client) ConvertNLToSQL(ctx context.Context, query string, schema map[string]interface{}) (string, error) {
	reqBody := map[string]interface{}{
		"query":  query,
		"schema": schema,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/v1/convert/nl-to-sql", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("agent API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	sql, ok := result["sql"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: sql not found")
	}

	return sql, nil
}

/* GenerateReply generates a context-aware reply via NeuronAgent API */
func (c *Client) GenerateReply(ctx context.Context, context []map[string]interface{}, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"context": context,
		"prompt":  prompt,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/v1/generate/reply", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("agent API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	reply, ok := result["reply"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: reply not found")
	}

	return reply, nil
}

// ============================================================================
// Session Management
// ============================================================================

/* CreateSession creates a new session */
func (c *Client) CreateSession(ctx context.Context, agentID string, config map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"agent_id": agentID,
	}
	if config != nil {
		reqBody["config"] = config
	}
	return c.makeRequest(ctx, "POST", "/api/v1/sessions", reqBody)
}

/* GetSession retrieves a session */
func (c *Client) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/sessions/"+sessionID, nil)
}

/* ListSessions lists all sessions */
func (c *Client) ListSessions(ctx context.Context, agentID string, filters map[string]interface{}) (map[string]interface{}, error) {
	endpoint := "/api/v1/sessions"
	if agentID != "" {
		endpoint += "?agent_id=" + url.QueryEscape(agentID)
	}
	return c.makeRequest(ctx, "GET", endpoint, filters)
}

/* UpdateSession updates a session */
func (c *Client) UpdateSession(ctx context.Context, sessionID string, updates map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "PUT", "/api/v1/sessions/"+sessionID, updates)
}

/* DeleteSession deletes a session */
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := c.makeRequest(ctx, "DELETE", "/api/v1/sessions/"+sessionID, nil)
	return err
}

/* CreateMessage creates a message in a session */
func (c *Client) CreateMessage(ctx context.Context, sessionID string, role string, content string, metadata map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"role":    role,
		"content": content,
	}
	if metadata != nil {
		reqBody["metadata"] = metadata
	}
	return c.makeRequest(ctx, "POST", "/api/v1/sessions/"+sessionID+"/messages", reqBody)
}

/* GetMessages retrieves messages from a session */
func (c *Client) GetMessages(ctx context.Context, sessionID string, limit int, offset int) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/api/v1/sessions/%s/messages", sessionID)
	if limit > 0 {
		endpoint += fmt.Sprintf("?limit=%d&offset=%d", limit, offset)
	}
	return c.makeRequest(ctx, "GET", endpoint, nil)
}

/* StreamMessages streams messages from a session */
func (c *Client) StreamMessages(ctx context.Context, sessionID string, handler func(map[string]interface{}) error) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint+"/api/v1/sessions/"+sessionID+"/messages/stream", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent API returned status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to decode message: %w", err)
		}
		if err := handler(msg); err != nil {
			return err
		}
	}
	return nil
}

/* ExecuteSession executes a session-based task */
func (c *Client) ExecuteSession(ctx context.Context, sessionID string, task string, tools []string) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"task": task,
	}
	if tools != nil {
		reqBody["tools"] = tools
	}
	return c.makeRequest(ctx, "POST", "/api/v1/sessions/"+sessionID+"/execute", reqBody)
}

// ============================================================================
// Workflow Management
// ============================================================================

/* CreateWorkflow creates a new workflow */
func (c *Client) CreateWorkflow(ctx context.Context, name string, definition map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"name":       name,
		"definition": definition,
	}
	if config != nil {
		reqBody["config"] = config
	}
	return c.makeRequest(ctx, "POST", "/api/v1/workflows", reqBody)
}

/* GetWorkflow retrieves a workflow */
func (c *Client) GetWorkflow(ctx context.Context, workflowID string) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/workflows/"+workflowID, nil)
}

/* ListWorkflows lists all workflows */
func (c *Client) ListWorkflows(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/workflows", filters)
}

/* UpdateWorkflow updates a workflow */
func (c *Client) UpdateWorkflow(ctx context.Context, workflowID string, updates map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "PUT", "/api/v1/workflows/"+workflowID, updates)
}

/* DeleteWorkflow deletes a workflow */
func (c *Client) DeleteWorkflow(ctx context.Context, workflowID string) error {
	_, err := c.makeRequest(ctx, "DELETE", "/api/v1/workflows/"+workflowID, nil)
	return err
}

/* ExecuteWorkflow executes a workflow */
func (c *Client) ExecuteWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "POST", "/api/v1/workflows/"+workflowID+"/execute", input)
}

/* GetWorkflowExecution retrieves a workflow execution */
func (c *Client) GetWorkflowExecution(ctx context.Context, executionID string) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/workflows/executions/"+executionID, nil)
}

/* ListWorkflowExecutions lists workflow executions */
func (c *Client) ListWorkflowExecutions(ctx context.Context, workflowID string, filters map[string]interface{}) (map[string]interface{}, error) {
	endpoint := "/api/v1/workflows/executions"
	if workflowID != "" {
		endpoint += "?workflow_id=" + url.QueryEscape(workflowID)
	}
	return c.makeRequest(ctx, "GET", endpoint, filters)
}

// ============================================================================
// Evaluation Framework
// ============================================================================

/* CreateEvalTask creates an evaluation task */
func (c *Client) CreateEvalTask(ctx context.Context, taskType string, input string, expectedOutput string, expectedToolSequence map[string]interface{}, goldenSQLSideEffects map[string]interface{}, metadata map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"task_type": taskType,
		"input":     input,
	}
	if expectedOutput != "" {
		reqBody["expected_output"] = expectedOutput
	}
	if expectedToolSequence != nil {
		reqBody["expected_tool_sequence"] = expectedToolSequence
	}
	if goldenSQLSideEffects != nil {
		reqBody["golden_sql_side_effects"] = goldenSQLSideEffects
	}
	if metadata != nil {
		reqBody["metadata"] = metadata
	}
	return c.makeRequest(ctx, "POST", "/api/v1/eval/tasks", reqBody)
}

/* GetEvalTask retrieves an evaluation task */
func (c *Client) GetEvalTask(ctx context.Context, taskID string) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/eval/tasks/"+taskID, nil)
}

/* ListEvalTasks lists evaluation tasks */
func (c *Client) ListEvalTasks(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/eval/tasks", filters)
}

/* CreateEvalRun creates an evaluation run */
func (c *Client) CreateEvalRun(ctx context.Context, datasetVersion string, agentID string, totalTasks int, metadata map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"dataset_version": datasetVersion,
	}
	if agentID != "" {
		reqBody["agent_id"] = agentID
	}
	if totalTasks > 0 {
		reqBody["total_tasks"] = totalTasks
	}
	if metadata != nil {
		reqBody["metadata"] = metadata
	}
	return c.makeRequest(ctx, "POST", "/api/v1/eval/runs", reqBody)
}

/* GetEvalRun retrieves an evaluation run */
func (c *Client) GetEvalRun(ctx context.Context, runID string) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/eval/runs/"+runID, nil)
}

/* ListEvalRuns lists evaluation runs */
func (c *Client) ListEvalRuns(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/eval/runs", filters)
}

/* GetEvalTaskResult retrieves an evaluation task result */
func (c *Client) GetEvalTaskResult(ctx context.Context, resultID string) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/eval/task-results/"+resultID, nil)
}

/* ListEvalTaskResults lists evaluation task results */
func (c *Client) ListEvalTaskResults(ctx context.Context, evalRunID string, filters map[string]interface{}) (map[string]interface{}, error) {
	endpoint := "/api/v1/eval/task-results"
	if evalRunID != "" {
		endpoint += "?eval_run_id=" + url.QueryEscape(evalRunID)
	}
	return c.makeRequest(ctx, "GET", endpoint, filters)
}

// ============================================================================
// Replay Operations
// ============================================================================

/* CreateSnapshot creates a snapshot of session state */
func (c *Client) CreateSnapshot(ctx context.Context, sessionID string, agentID string, userMessage string, executionState map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"session_id":      sessionID,
		"user_message":    userMessage,
		"execution_state": executionState,
	}
	if agentID != "" {
		reqBody["agent_id"] = agentID
	}
	return c.makeRequest(ctx, "POST", "/api/v1/snapshots", reqBody)
}

/* GetSnapshot retrieves a snapshot */
func (c *Client) GetSnapshot(ctx context.Context, snapshotID string) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/snapshots/"+snapshotID, nil)
}

/* ListSnapshots lists snapshots */
func (c *Client) ListSnapshots(ctx context.Context, sessionID string, filters map[string]interface{}) (map[string]interface{}, error) {
	endpoint := "/api/v1/snapshots"
	if sessionID != "" {
		endpoint += "?session_id=" + url.QueryEscape(sessionID)
	}
	return c.makeRequest(ctx, "GET", endpoint, filters)
}

/* ReplaySession replays a session from a snapshot */
func (c *Client) ReplaySession(ctx context.Context, snapshotID string, options map[string]interface{}) (map[string]interface{}, error) {
	reqBody := make(map[string]interface{})
	if options != nil {
		reqBody = options
	}
	return c.makeRequest(ctx, "POST", "/api/v1/snapshots/"+snapshotID+"/replay", reqBody)
}

// ============================================================================
// Specializations
// ============================================================================

/* GetSpecializations retrieves agent specializations */
func (c *Client) GetSpecializations(ctx context.Context, agentID string) (map[string]interface{}, error) {
	endpoint := "/api/v1/specializations"
	if agentID != "" {
		endpoint += "?agent_id=" + url.QueryEscape(agentID)
	}
	return c.makeRequest(ctx, "GET", endpoint, nil)
}

/* CreateSpecialization creates a specialization */
func (c *Client) CreateSpecialization(ctx context.Context, agentID string, name string, config map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"agent_id": agentID,
		"name":     name,
	}
	if config != nil {
		reqBody["config"] = config
	}
	return c.makeRequest(ctx, "POST", "/api/v1/specializations", reqBody)
}

/* UpdateSpecialization updates a specialization */
func (c *Client) UpdateSpecialization(ctx context.Context, specializationID string, updates map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "PUT", "/api/v1/specializations/"+specializationID, updates)
}

// ============================================================================
// Tool Management
// ============================================================================

/* ListTools lists available tools */
func (c *Client) ListTools(ctx context.Context, category string) (map[string]interface{}, error) {
	endpoint := "/api/v1/tools"
	if category != "" {
		endpoint += "?category=" + url.QueryEscape(category)
	}
	return c.makeRequest(ctx, "GET", endpoint, nil)
}

/* GetTool retrieves tool information */
func (c *Client) GetTool(ctx context.Context, toolName string) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "GET", "/api/v1/tools/"+toolName, nil)
}

/* ExecuteTool executes a tool */
func (c *Client) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}
	return c.makeRequest(ctx, "POST", "/api/v1/tools/execute", reqBody)
}

// ============================================================================
// Helper Methods
// ============================================================================

/* makeRequest is a helper method for making HTTP requests */
func (c *Client) makeRequest(ctx context.Context, method string, endpoint string, body interface{}) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if resp.ContentLength > 0 {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return result, nil
}
