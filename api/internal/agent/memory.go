package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* MemoryService provides persistent agent memory storage */
type MemoryService struct {
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
}

/* NewMemoryService creates a new agent memory service */
func NewMemoryService(pool *pgxpool.Pool, neurondbClient *neurondb.Client) *MemoryService {
	return &MemoryService{
		pool:           pool,
		neurondbClient: neurondbClient,
	}
}

/* AgentMemory represents agent memory entry */
type AgentMemory struct {
	ID            uuid.UUID              `json:"id"`
	AgentID       string                 `json:"agent_id"`
	MemoryKey     string                 `json:"memory_key"`
	MemoryValue   map[string]interface{} `json:"memory_value"`
	Embedding     string                 `json:"-"` // Not exposed in JSON
	ImportanceScore float64              `json:"importance_score"`
	LastAccessedAt *time.Time            `json:"last_accessed_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* StoreMemory stores agent memory with vector embedding */
func (s *MemoryService) StoreMemory(ctx context.Context, agentID string, memoryKey string, memoryValue map[string]interface{}, importanceScore float64) error {
	// Generate embedding from memory content
	memoryText := fmt.Sprintf("%v", memoryValue)
	embedding, err := s.neurondbClient.GenerateEmbedding(ctx, memoryText, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		embedding = "" // Continue without embedding if generation fails
	}

	memoryValueJSON, _ := json.Marshal(memoryValue)
	now := time.Now()

	// Check if agent_memory table exists, if not create it
	// For now, assume it exists or will be created via migration
	query := `
		INSERT INTO neuronip.agent_memory 
		(id, agent_id, memory_key, memory_value, embedding, importance_score, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4::vector, $5, $6, $7)
		ON CONFLICT (agent_id, memory_key) 
		DO UPDATE SET 
			memory_value = EXCLUDED.memory_value,
			embedding = EXCLUDED.embedding,
			importance_score = EXCLUDED.importance_score,
			last_accessed_at = NOW(),
			updated_at = EXCLUDED.updated_at`

	_, err = s.pool.Exec(ctx, query, agentID, memoryKey, memoryValueJSON, embedding, importanceScore, now, now)
	if err != nil {
		// Table might not exist, return error with suggestion
		return fmt.Errorf("failed to store agent memory (ensure agent_memory table exists): %w", err)
	}

	return nil
}

/* GetMemory retrieves agent memory by key */
func (s *MemoryService) GetMemory(ctx context.Context, agentID string, memoryKey string) (*AgentMemory, error) {
	query := `
		SELECT id, agent_id, memory_key, memory_value, importance_score, last_accessed_at, created_at, updated_at
		FROM neuronip.agent_memory
		WHERE agent_id = $1 AND memory_key = $2`

	var mem AgentMemory
	var memoryValueJSON json.RawMessage
	var lastAccessedAt interface{}

	err := s.pool.QueryRow(ctx, query, agentID, memoryKey).Scan(
		&mem.ID, &mem.AgentID, &mem.MemoryKey, &memoryValueJSON,
		&mem.ImportanceScore, &lastAccessedAt, &mem.CreatedAt, &mem.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("memory not found: %w", err)
	}

	if memoryValueJSON != nil {
		json.Unmarshal(memoryValueJSON, &mem.MemoryValue)
	}

	// Update last_accessed_at
	s.pool.Exec(ctx, `UPDATE neuronip.agent_memory SET last_accessed_at = NOW() WHERE id = $1`, mem.ID)

	return &mem, nil
}

/* SearchMemory performs semantic search on agent memory */
func (s *MemoryService) SearchMemory(ctx context.Context, agentID string, query string, limit int) ([]AgentMemory, error) {
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
		SELECT id, agent_id, memory_key, memory_value, importance_score, last_accessed_at, created_at, updated_at,
		       1 - (embedding <=> $1::vector) as similarity
		FROM neuronip.agent_memory
		WHERE agent_id = $2 AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $3`

	rows, err := s.pool.Query(ctx, searchQuery, queryEmbedding, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search agent memory: %w", err)
	}
	defer rows.Close()

	var memories []AgentMemory
	for rows.Next() {
		var mem AgentMemory
		var memoryValueJSON json.RawMessage
		var similarity float64
		var lastAccessedAt interface{}

		err := rows.Scan(
			&mem.ID, &mem.AgentID, &mem.MemoryKey, &memoryValueJSON,
			&mem.ImportanceScore, &lastAccessedAt, &mem.CreatedAt, &mem.UpdatedAt, &similarity,
		)
		if err != nil {
			continue
		}

		if memoryValueJSON != nil {
			json.Unmarshal(memoryValueJSON, &mem.MemoryValue)
		}

		memories = append(memories, mem)
	}

	return memories, nil
}

/* GetAllMemory retrieves all memory for an agent */
func (s *MemoryService) GetAllMemory(ctx context.Context, agentID string) ([]AgentMemory, error) {
	query := `
		SELECT id, agent_id, memory_key, memory_value, importance_score, last_accessed_at, created_at, updated_at
		FROM neuronip.agent_memory
		WHERE agent_id = $1
		ORDER BY importance_score DESC, updated_at DESC`

	rows, err := s.pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent memory: %w", err)
	}
	defer rows.Close()

	var memories []AgentMemory
	for rows.Next() {
		var mem AgentMemory
		var memoryValueJSON json.RawMessage
		var lastAccessedAt interface{}

		err := rows.Scan(
			&mem.ID, &mem.AgentID, &mem.MemoryKey, &memoryValueJSON,
			&mem.ImportanceScore, &lastAccessedAt, &mem.CreatedAt, &mem.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if memoryValueJSON != nil {
			json.Unmarshal(memoryValueJSON, &mem.MemoryValue)
		}

		memories = append(memories, mem)
	}

	return memories, nil
}

/* DeleteMemory deletes agent memory */
func (s *MemoryService) DeleteMemory(ctx context.Context, agentID string, memoryKey string) error {
	query := `DELETE FROM neuronip.agent_memory WHERE agent_id = $1 AND memory_key = $2`
	_, err := s.pool.Exec(ctx, query, agentID, memoryKey)
	return err
}
