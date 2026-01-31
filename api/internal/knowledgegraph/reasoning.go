package knowledgegraph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ReasoningService provides graph-based reasoning for agents */
type ReasoningService struct {
	pool  *pgxpool.Pool
	kgService *Service
}

/* NewReasoningService creates a new reasoning service */
func NewReasoningService(pool *pgxpool.Pool, kgService *Service) *ReasoningService {
	return &ReasoningService{
		pool:      pool,
		kgService: kgService,
	}
}

/* ReasonRequest represents a reasoning request */
type ReasonRequest struct {
	Question      string
	StartEntityID *uuid.UUID
	MaxDepth      int
	MaxResults    int
}

/* Reason performs graph-based reasoning */
func (rs *ReasoningService) Reason(ctx context.Context, req ReasonRequest) (map[string]interface{}, error) {
	if req.MaxDepth <= 0 {
		req.MaxDepth = 3
	}
	if req.MaxResults <= 0 {
		req.MaxResults = 10
	}
	
	// Search for relevant entities
	entities, err := rs.kgService.SearchEntities(ctx, req.Question, nil, req.MaxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}
	
	// Perform graph traversal from relevant entities
	reasoningPaths := []map[string]interface{}{}
	
	for _, entity := range entities {
		if req.StartEntityID != nil && entity.ID != *req.StartEntityID {
			continue
		}
		
		// Traverse graph from this entity
		traversal, err := rs.kgService.TraverseGraph(ctx, entity.ID, req.MaxDepth, nil, "")
		if err != nil {
			continue
		}
		
		reasoningPaths = append(reasoningPaths, map[string]interface{}{
			"start_entity": entity,
			"paths":        traversal.Paths,
			"entities":     traversal.Entities,
			"links":        traversal.Links,
		})
	}
	
	return map[string]interface{}{
		"question":       req.Question,
		"reasoning_paths": reasoningPaths,
		"entities_found": len(entities),
	}, nil
}
