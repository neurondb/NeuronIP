package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/agent"
)

/* AgentsService provides agent management functionality */
type AgentsService struct {
	pool        *pgxpool.Pool
	agentClient *agent.Client
}

/* NewAgentsService creates a new agents service */
func NewAgentsService(pool *pgxpool.Pool, agentClient *agent.Client) *AgentsService {
	return &AgentsService{
		pool:        pool,
		agentClient: agentClient,
	}
}

/* Agent represents an agent */
type Agent struct {
	ID               uuid.UUID              `json:"id"`
	Name             string                 `json:"name"`
	AgentType        string                 `json:"agent_type"`
	Config           map[string]interface{} `json:"config"`
	Status           string                 `json:"status"`
	PerformanceMetrics map[string]interface{} `json:"performance_metrics,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

/* AgentPerformance represents agent performance metrics */
type AgentPerformance struct {
	ID          uuid.UUID `json:"id"`
	AgentID     uuid.UUID `json:"agent_id"`
	MetricName  string    `json:"metric_name"`
	MetricValue float64   `json:"metric_value"`
	Timestamp   time.Time `json:"timestamp"`
}

/* CreateAgent creates a new agent */
func (s *AgentsService) CreateAgent(ctx context.Context, agent Agent) (*Agent, error) {
	id := uuid.New()
	configJSON, _ := json.Marshal(agent.Config)
	perfJSON, _ := json.Marshal(agent.PerformanceMetrics)
	now := time.Now()

	query := `
		INSERT INTO neuronip.agents (id, name, agent_type, config, status, performance_metrics, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, agent_type, config, status, performance_metrics, created_at, updated_at`

	var result Agent
	var configJSONRaw, perfJSONRaw json.RawMessage

	err := s.pool.QueryRow(ctx, query,
		id, agent.Name, agent.AgentType, configJSON, agent.Status, perfJSON, now, now,
	).Scan(
		&result.ID, &result.Name, &result.AgentType, &configJSONRaw,
		&result.Status, &perfJSONRaw, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	json.Unmarshal(configJSONRaw, &result.Config)
	if perfJSONRaw != nil {
		json.Unmarshal(perfJSONRaw, &result.PerformanceMetrics)
	}

	return &result, nil
}

/* GetAgent retrieves an agent by ID */
func (s *AgentsService) GetAgent(ctx context.Context, id uuid.UUID) (*Agent, error) {
	query := `
		SELECT id, name, agent_type, config, status, performance_metrics, created_at, updated_at
		FROM neuronip.agents
		WHERE id = $1`

	var result Agent
	var configJSONRaw, perfJSONRaw json.RawMessage

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.Name, &result.AgentType, &configJSONRaw,
		&result.Status, &perfJSONRaw, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	json.Unmarshal(configJSONRaw, &result.Config)
	if perfJSONRaw != nil {
		json.Unmarshal(perfJSONRaw, &result.PerformanceMetrics)
	}

	return &result, nil
}

/* ListAgents lists all agents */
func (s *AgentsService) ListAgents(ctx context.Context) ([]Agent, error) {
	query := `
		SELECT id, name, agent_type, config, status, performance_metrics, created_at, updated_at
		FROM neuronip.agents
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		var configJSONRaw, perfJSONRaw json.RawMessage

		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.AgentType, &configJSONRaw,
			&agent.Status, &perfJSONRaw, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(configJSONRaw, &agent.Config)
		if perfJSONRaw != nil {
			json.Unmarshal(perfJSONRaw, &agent.PerformanceMetrics)
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

/* UpdateAgent updates an agent */
func (s *AgentsService) UpdateAgent(ctx context.Context, id uuid.UUID, agent Agent) (*Agent, error) {
	configJSON, _ := json.Marshal(agent.Config)
	perfJSON, _ := json.Marshal(agent.PerformanceMetrics)

	query := `
		UPDATE neuronip.agents
		SET name = $1, agent_type = $2, config = $3, status = $4, performance_metrics = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, name, agent_type, config, status, performance_metrics, created_at, updated_at`

	var result Agent
	var configJSONRaw, perfJSONRaw json.RawMessage

	err := s.pool.QueryRow(ctx, query,
		agent.Name, agent.AgentType, configJSON, agent.Status, perfJSON, id,
	).Scan(
		&result.ID, &result.Name, &result.AgentType, &configJSONRaw,
		&result.Status, &perfJSONRaw, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	json.Unmarshal(configJSONRaw, &result.Config)
	if perfJSONRaw != nil {
		json.Unmarshal(perfJSONRaw, &result.PerformanceMetrics)
	}

	return &result, nil
}

/* DeleteAgent deletes an agent */
func (s *AgentsService) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM neuronip.agents WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id)
	return err
}

/* GetPerformance retrieves performance metrics for an agent */
func (s *AgentsService) GetPerformance(ctx context.Context, id uuid.UUID) ([]AgentPerformance, error) {
	query := `
		SELECT id, agent_id, metric_name, metric_value, timestamp
		FROM neuronip.agent_performance
		WHERE agent_id = $1
		ORDER BY timestamp DESC
		LIMIT 100`

	rows, err := s.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance: %w", err)
	}
	defer rows.Close()

	var perf []AgentPerformance
	for rows.Next() {
		var p AgentPerformance
		err := rows.Scan(
			&p.ID, &p.AgentID, &p.MetricName, &p.MetricValue, &p.Timestamp,
		)
		if err != nil {
			continue
		}
		perf = append(perf, p)
	}

	return perf, nil
}

/* RecordPerformance records a performance metric for an agent */
func (s *AgentsService) RecordPerformance(ctx context.Context, agentID uuid.UUID, metricName string, metricValue float64) error {
	query := `
		INSERT INTO neuronip.agent_performance (id, agent_id, metric_name, metric_value, timestamp)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())`
	_, err := s.pool.Exec(ctx, query, agentID, metricName, metricValue)
	return err
}

/* DeployAgent updates agent status to active */
func (s *AgentsService) DeployAgent(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE neuronip.agents SET status = 'active', updated_at = NOW() WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id)
	return err
}

/* CreateEvalTask creates an evaluation task using NeuronAgent evaluation framework */
func (s *AgentsService) CreateEvalTask(ctx context.Context, taskType string, input string, expectedOutput string, expectedToolSequence map[string]interface{}, goldenSQLSideEffects map[string]interface{}, metadata map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.CreateEvalTask(ctx, taskType, input, expectedOutput, expectedToolSequence, goldenSQLSideEffects, metadata)
}

/* CreateEvalRun creates an evaluation run using NeuronAgent */
func (s *AgentsService) CreateEvalRun(ctx context.Context, datasetVersion string, agentID string, totalTasks int, metadata map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.CreateEvalRun(ctx, datasetVersion, agentID, totalTasks, metadata)
}

/* ListEvalRuns lists evaluation runs */
func (s *AgentsService) ListEvalRuns(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.ListEvalRuns(ctx, filters)
}

/* GetEvalTaskResults retrieves evaluation task results */
func (s *AgentsService) GetEvalTaskResults(ctx context.Context, evalRunID string, filters map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.ListEvalTaskResults(ctx, evalRunID, filters)
}

/* GetEvalTaskResult retrieves a specific evaluation task result */
func (s *AgentsService) GetEvalTaskResult(ctx context.Context, resultID string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.GetEvalTaskResult(ctx, resultID)
}

/* GetEvalRun retrieves an evaluation run */
func (s *AgentsService) GetEvalRun(ctx context.Context, runID string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.GetEvalRun(ctx, runID)
}

/* GetEvalTask retrieves an evaluation task */
func (s *AgentsService) GetEvalTask(ctx context.Context, taskID string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.GetEvalTask(ctx, taskID)
}

/* ListEvalTasks lists evaluation tasks */
func (s *AgentsService) ListEvalTasks(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.ListEvalTasks(ctx, filters)
}

/* GetSpecializations retrieves agent specializations */
func (s *AgentsService) GetSpecializations(ctx context.Context, agentID string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.GetSpecializations(ctx, agentID)
}

/* CreateSpecialization creates a specialization for an agent */
func (s *AgentsService) CreateSpecialization(ctx context.Context, agentID string, name string, config map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.CreateSpecialization(ctx, agentID, name, config)
}

/* UpdateSpecialization updates a specialization */
func (s *AgentsService) UpdateSpecialization(ctx context.Context, specializationID string, updates map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.UpdateSpecialization(ctx, specializationID, updates)
}

/* ListTools lists available tools from NeuronAgent */
func (s *AgentsService) ListTools(ctx context.Context, category string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.ListTools(ctx, category)
}

/* GetTool retrieves tool information from NeuronAgent */
func (s *AgentsService) GetTool(ctx context.Context, toolName string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.GetTool(ctx, toolName)
}

/* ExecuteTool executes a tool via NeuronAgent */
func (s *AgentsService) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}
	return s.agentClient.ExecuteTool(ctx, toolName, args)
}

/* AutoRegisterAllMCPTools automatically discovers and registers all MCP tools with an agent */
func (s *AgentsService) AutoRegisterAllMCPTools(ctx context.Context, agentID uuid.UUID, mcpClient interface{}) error {
	if s.agentClient == nil {
		return fmt.Errorf("agent client not configured")
	}

	// Check if mcpClient has ListTools method (type assertion)
	type MCPClient interface {
		ListTools(ctx context.Context) ([]map[string]interface{}, error)
	}

	mcp, ok := mcpClient.(MCPClient)
	if !ok {
		return fmt.Errorf("MCP client does not implement ListTools")
	}

	// List all available MCP tools
	mcpTools, err := mcp.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list MCP tools: %w", err)
	}

	// Extract tool names
	toolNames := make([]string, 0, len(mcpTools))
	for _, tool := range mcpTools {
		if name, ok := tool["name"].(string); ok {
			toolNames = append(toolNames, name)
		}
	}

	// Register all tools with the agent
	return s.RegisterMCPToolsWithAgent(ctx, agentID, toolNames)
}

/* RegisterMCPToolsWithAgent registers MCP tools with NeuronAgent for agent use */
func (s *AgentsService) RegisterMCPToolsWithAgent(ctx context.Context, agentID uuid.UUID, mcpTools []string) error {
	if s.agentClient == nil {
		return fmt.Errorf("agent client not configured")
	}

	// Get agent to update its tool configuration
	agent, err := s.GetAgent(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Get current tools
	currentTools := []string{}
	if tools, ok := agent.Config["tools"].([]interface{}); ok {
		for _, tool := range tools {
			if toolStr, ok := tool.(string); ok {
				currentTools = append(currentTools, toolStr)
			}
		}
	}

	// Add MCP tools to agent's tool list
	toolMap := make(map[string]bool)
	for _, tool := range currentTools {
		toolMap[tool] = true
	}
	for _, mcpTool := range mcpTools {
		if !toolMap[mcpTool] {
			currentTools = append(currentTools, mcpTool)
		}
	}

	// Update agent config with MCP tools
	agent.Config["tools"] = currentTools
	agent.Config["mcp_tools"] = mcpTools

	_, err = s.UpdateAgent(ctx, agentID, *agent)
	return err
}
