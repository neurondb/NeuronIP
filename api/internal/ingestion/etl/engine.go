package etl

import (
	"context"
	"fmt"
)

/* ETLEngine executes ETL transformations */
type ETLEngine struct {
	transformations map[string]Transformation
}

/* NewETLEngine creates a new ETL engine */
func NewETLEngine() *ETLEngine {
	engine := &ETLEngine{
		transformations: make(map[string]Transformation),
	}
	
	// Register built-in transformations
	engine.RegisterTransformation("filter", &FilterTransformation{})
	engine.RegisterTransformation("map", &MapTransformation{})
	engine.RegisterTransformation("aggregate", &AggregateTransformation{})
	engine.RegisterTransformation("join", &JoinTransformation{})
	
	return engine
}

/* RegisterTransformation registers a transformation */
func (e *ETLEngine) RegisterTransformation(name string, transformation Transformation) {
	e.transformations[name] = transformation
}

/* Execute executes a transformation pipeline */
func (e *ETLEngine) Execute(ctx context.Context, pipeline Pipeline, data []map[string]interface{}) ([]map[string]interface{}, error) {
	result := data
	
	for _, step := range pipeline.Steps {
		transformation, exists := e.transformations[step.Type]
		if !exists {
			return nil, fmt.Errorf("unknown transformation type: %s", step.Type)
		}
		
		var err error
		result, err = transformation.Transform(ctx, result, step.Config)
		if err != nil {
			return nil, fmt.Errorf("transformation %s failed: %w", step.Type, err)
		}
	}
	
	return result, nil
}

/* Pipeline represents an ETL transformation pipeline */
type Pipeline struct {
	Steps []PipelineStep `json:"steps"`
}

/* PipelineStep represents a single step in the pipeline */
type PipelineStep struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

/* Transformation defines the interface for ETL transformations */
type Transformation interface {
	Transform(ctx context.Context, data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error)
}
