package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* AgentsService provides agent management functionality */
type AgentsService struct {
	pool *pgxpool.Pool
}

/* NewAgentsService creates a new agents service */
func NewAgentsService(pool *pgxpool.Pool) *AgentsService {
	return &AgentsService{pool: pool}
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
