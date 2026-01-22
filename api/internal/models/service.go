package models

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/mcp"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides model management functionality */
type Service struct {
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
	mcpClient      *mcp.Client
}

/* NewService creates a new models service */
func NewService(pool *pgxpool.Pool, neurondbClient *neurondb.Client, mcpClient *mcp.Client) *Service {
	return &Service{
		pool:           pool,
		neurondbClient: neurondbClient,
		mcpClient:      mcpClient,
	}
}

/* Model represents a ML model */
type Model struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	ModelType   string                 `json:"model_type"` // "classification", "regression", "embedding"
	ModelPath   string                 `json:"model_path"`
	Version     string                 `json:"version"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Performance map[string]interface{} `json:"performance,omitempty"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* RegisterModel registers a new model */
func (s *Service) RegisterModel(ctx context.Context, model Model) (*Model, error) {
	model.ID = uuid.New()
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(model.Metadata)
	performanceJSON, _ := json.Marshal(model.Performance)

	// Create model_registry table if it doesn't exist
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS neuronip.model_registry (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL UNIQUE,
			model_type TEXT NOT NULL CHECK (model_type IN ('classification', 'regression', 'embedding')),
			model_path TEXT NOT NULL,
			version TEXT NOT NULL DEFAULT '1.0.0',
			metadata JSONB DEFAULT '{}',
			performance JSONB DEFAULT '{}',
			enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`
	
	s.pool.Exec(ctx, createTableQuery)

	// Insert model
	query := `
		INSERT INTO neuronip.model_registry 
		(id, name, model_type, model_path, version, metadata, performance, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		model.ID, model.Name, model.ModelType, model.ModelPath, model.Version,
		metadataJSON, performanceJSON, model.Enabled, model.CreatedAt, model.UpdatedAt,
	).Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to register model: %w", err)
	}

	return &model, nil
}

/* GetModel retrieves a model by ID */
func (s *Service) GetModel(ctx context.Context, id uuid.UUID) (*Model, error) {
	var model Model
	var metadataJSON, performanceJSON json.RawMessage

	query := `
		SELECT id, name, model_type, model_path, version, metadata, performance, enabled, created_at, updated_at
		FROM neuronip.model_registry
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&model.ID, &model.Name, &model.ModelType, &model.ModelPath, &model.Version,
		&metadataJSON, &performanceJSON, &model.Enabled, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &model.Metadata)
	}
	if performanceJSON != nil {
		json.Unmarshal(performanceJSON, &model.Performance)
	}

	return &model, nil
}

/* ListModelsFromNeuronDB lists models directly from NeuronDB */
func (s *Service) ListModelsFromNeuronDB(ctx context.Context) ([]map[string]interface{}, error) {
	return s.neurondbClient.ListModels(ctx)
}

/* ListModels lists all registered models */
func (s *Service) ListModels(ctx context.Context, modelType string, enabledOnly bool) ([]Model, error) {
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	if modelType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("model_type = $%d", argIndex))
		args = append(args, modelType)
		argIndex++
	}
	if enabledOnly {
		whereClauses = append(whereClauses, fmt.Sprintf("enabled = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, name, model_type, model_path, version, metadata, performance, enabled, created_at, updated_at
		FROM neuronip.model_registry
		%s
		ORDER BY created_at DESC`, whereClause)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var model Model
		var metadataJSON, performanceJSON json.RawMessage

		err := rows.Scan(
			&model.ID, &model.Name, &model.ModelType, &model.ModelPath, &model.Version,
			&metadataJSON, &performanceJSON, &model.Enabled, &model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &model.Metadata)
		}
		if performanceJSON != nil {
			json.Unmarshal(performanceJSON, &model.Performance)
		}

		models = append(models, model)
	}

	return models, nil
}

/* EvaluateModel evaluates model performance using NeuronDB evaluation */
func (s *Service) EvaluateModel(ctx context.Context, modelID uuid.UUID, testData []map[string]interface{}) (map[string]interface{}, error) {
	model, err := s.GetModel(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	// Try using NeuronDB EvaluateModel if test data is in a table
	// This is more efficient for large datasets
	if testTable, ok := testData[0]["_test_table"].(string); ok && len(testData) == 1 {
		if targetColumn, ok := testData[0]["_target_column"].(string); ok {
			// Use NeuronDB EvaluateModel for table-based evaluation
			results, err := s.neurondbClient.EvaluateModel(ctx, model.ModelPath, testTable, targetColumn)
			if err == nil {
				// Update model performance
				model.Performance = results
				performanceJSON, _ := json.Marshal(results)
				s.pool.Exec(ctx, `UPDATE neuronip.model_registry SET performance = $1, updated_at = NOW() WHERE id = $2`,
					performanceJSON, modelID)
				return results, nil
			}
			// Fall through to individual evaluation if table-based fails
		}
	}

	// Use NeuronDB ML functions for evaluation based on model type
	var results map[string]interface{}
	switch model.ModelType {
	case "classification":
		// Extract ground truth labels from test data (assumes "label" or "ground_truth" field)
		var groundTruth []interface{}
		var featuresList []map[string]interface{}
		for _, item := range testData {
			if label, ok := item["label"]; ok {
				groundTruth = append(groundTruth, label)
			} else if gt, ok := item["ground_truth"]; ok {
				groundTruth = append(groundTruth, gt)
			}
			// Create feature map without label
			features := make(map[string]interface{})
			for k, v := range item {
				if k != "label" && k != "ground_truth" {
					features[k] = v
				}
			}
			featuresList = append(featuresList, features)
		}
		
		// Use BatchPredict if available for better performance
		var predictions []map[string]interface{}
		if len(featuresList) > 1 {
			// Try batch prediction first
			predResults, err := s.neurondbClient.BatchPredict(ctx, model.ModelPath, featuresList)
			if err == nil {
				// Convert batch results to predictions format
				for _, pred := range predResults {
					if predMap, ok := pred.(map[string]interface{}); ok {
						predictions = append(predictions, predMap)
					} else {
						predictions = append(predictions, map[string]interface{}{"prediction": pred})
					}
				}
			} else {
				// Fall back to individual predictions
				for _, features := range featuresList {
					featuresJSON, _ := json.Marshal(features)
					featuresStr := string(featuresJSON)
					prediction, err := s.neurondbClient.Classify(ctx, featuresStr, model.ModelPath)
					if err != nil {
						return nil, fmt.Errorf("evaluation failed: %w", err)
					}
					predictions = append(predictions, prediction)
				}
			}
		} else if len(featuresList) == 1 {
			// Single prediction
			featuresJSON, _ := json.Marshal(featuresList[0])
			featuresStr := string(featuresJSON)
			prediction, err := s.neurondbClient.Classify(ctx, featuresStr, model.ModelPath)
			if err != nil {
				return nil, fmt.Errorf("evaluation failed: %w", err)
			}
			predictions = append(predictions, prediction)
		}
		
		// Calculate metrics from predictions vs ground truth
		if len(groundTruth) == 0 {
			// No ground truth provided, return prediction results without metrics
			results = map[string]interface{}{"predictions": predictions}
		} else {
			// Calculate actual classification metrics from predictions vs ground truth
			// Convert predictions and groundTruth to []interface{} for the metrics function
			predInterfaces := make([]interface{}, len(predictions))
			for i, p := range predictions {
				predInterfaces[i] = p
			}
			gtInterfaces := make([]interface{}, len(groundTruth))
			for i, gt := range groundTruth {
				gtInterfaces[i] = gt
			}
			
			metrics := calculateClassificationMetrics(predInterfaces, gtInterfaces)
			results = map[string]interface{}{
				"accuracy":  metrics["accuracy"],
				"precision": metrics["precision"],
				"recall":    metrics["recall"],
				"f1_score":  metrics["f1_score"],
				"predictions": predictions,
			}
		}

	case "regression":
		// Extract ground truth targets from test data (assumes "target" or "ground_truth" field)
		var groundTruth []float64
		var featuresList []map[string]interface{}
		for _, item := range testData {
			if target, ok := item["target"]; ok {
				if t, ok := toFloat64(target); ok {
					groundTruth = append(groundTruth, t)
				}
			} else if gt, ok := item["ground_truth"]; ok {
				if t, ok := toFloat64(gt); ok {
					groundTruth = append(groundTruth, t)
				}
			}
			// Create feature map without target
			features := make(map[string]interface{})
			for k, v := range item {
				if k != "target" && k != "ground_truth" {
					features[k] = v
				}
			}
			featuresList = append(featuresList, features)
		}
		
		// Use BatchPredict if available for better performance
		var predictions []float64
		if len(featuresList) > 1 {
			// Try batch prediction first
			predResults, err := s.neurondbClient.BatchPredict(ctx, model.ModelPath, featuresList)
			if err == nil {
				// Convert batch results to float64 array
				for _, pred := range predResults {
					if predFloat, ok := toFloat64(pred); ok {
						predictions = append(predictions, predFloat)
					} else if predMap, ok := pred.(map[string]interface{}); ok {
						if val, ok := predMap["prediction"]; ok {
							if predFloat, ok := toFloat64(val); ok {
								predictions = append(predictions, predFloat)
							}
						}
					}
				}
			} else {
				// Fall back to individual predictions
				for _, features := range featuresList {
					prediction, err := s.neurondbClient.Regress(ctx, features, model.ModelPath)
					if err != nil {
						return nil, fmt.Errorf("evaluation failed: %w", err)
					}
					predictions = append(predictions, prediction)
				}
			}
		} else if len(featuresList) == 1 {
			// Single prediction
			prediction, err := s.neurondbClient.Regress(ctx, featuresList[0], model.ModelPath)
			if err != nil {
				return nil, fmt.Errorf("evaluation failed: %w", err)
			}
			predictions = append(predictions, prediction)
		}
		
		// Calculate metrics from predictions vs ground truth
		if len(groundTruth) == 0 || len(groundTruth) != len(predictions) {
			// No ground truth or mismatch, return predictions only
			results = map[string]interface{}{
				"predictions": predictions,
				"count": len(predictions),
			}
		} else {
			// Calculate RMSE, MAE, R2
			metrics := calculateRegressionMetrics(predictions, groundTruth)
			results = metrics
			results["predictions"] = predictions
		}

	default:
		return nil, fmt.Errorf("evaluation not supported for model type: %s", model.ModelType)
	}

	// Update model performance
	model.Performance = results
	performanceJSON, _ := json.Marshal(results)
	s.pool.Exec(ctx, `UPDATE neuronip.model_registry SET performance = $1, updated_at = NOW() WHERE id = $2`,
		performanceJSON, modelID)

	return results, nil
}

/* TrainModelViaMCP trains a model using MCP TrainModel tool */
func (s *Service) TrainModelViaMCP(ctx context.Context, algorithm string, tableName string, targetColumn string, featureColumns []string, options map[string]interface{}) (string, error) {
	if s.mcpClient == nil {
		return "", fmt.Errorf("MCP client not configured")
	}

	result, err := s.mcpClient.TrainModel(ctx, algorithm, tableName, targetColumn, featureColumns, options)
	if err != nil {
		return "", fmt.Errorf("failed to train model via MCP: %w", err)
	}

	if modelPath, ok := result["model_path"].(string); ok {
		return modelPath, nil
	}
	return "", fmt.Errorf("model_path not found in MCP result")
}

/* TrainModel trains a model using NeuronDB ML training */
func (s *Service) TrainModel(ctx context.Context, trainingData []map[string]interface{}, config map[string]interface{}) (*Model, error) {
	modelType, ok := config["model_type"].(string)
	if !ok {
		modelType = "classification"
	}

	modelName, ok := config["name"].(string)
	if !ok {
		modelName = fmt.Sprintf("model_%s", uuid.New().String()[:8])
	}

	// Determine algorithm based on model type
	algorithm := "random_forest"
	if alg, ok := config["algorithm"].(string); ok {
		algorithm = alg
	} else {
		switch modelType {
		case "classification":
			algorithm = "random_forest"
		case "regression":
			algorithm = "linear_regression"
		}
	}

	// Get table name for training data
	tableName, ok := config["table_name"].(string)
	if !ok {
		// If no table provided, we need to create a temporary table or use a different approach
		return nil, fmt.Errorf("table_name required in config for NeuronDB training")
	}

	// Get target column
	targetColumn, ok := config["target_column"].(string)
	if !ok {
		return nil, fmt.Errorf("target_column required in config for NeuronDB training")
	}

	// Get feature columns
	featureColumns := []string{}
	if features, ok := config["feature_columns"].([]interface{}); ok {
		for _, f := range features {
			if fStr, ok := f.(string); ok {
				featureColumns = append(featureColumns, fStr)
			}
		}
	}
	if len(featureColumns) == 0 {
		return nil, fmt.Errorf("feature_columns required in config for NeuronDB training")
	}

	// Build training options
	options := make(map[string]interface{})
	if opts, ok := config["options"].(map[string]interface{}); ok {
		options = opts
	}
	options["training_data_size"] = len(trainingData)

	// Use NeuronDB ML training function
	modelPath, err := s.neurondbClient.TrainModel(ctx, algorithm, tableName, targetColumn, featureColumns, options)
	if err != nil {
		return nil, fmt.Errorf("failed to train model using NeuronDB: %w", err)
	}

	// Create model record
	model := Model{
		Name:      modelName,
		ModelType: modelType,
		ModelPath: modelPath,
		Version:   "1.0.0",
		Metadata: map[string]interface{}{
			"training_data_size": len(trainingData),
			"config":             config,
			"algorithm":          algorithm,
			"table_name":         tableName,
			"target_column":      targetColumn,
			"feature_columns":    featureColumns,
		},
		Enabled: true,
	}

	return s.RegisterModel(ctx, model)
}

/* InferModel performs inference using a model */
func (s *Service) InferModel(ctx context.Context, modelID uuid.UUID, input map[string]interface{}) (interface{}, error) {
	model, err := s.GetModel(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	// Use NeuronDB ML functions based on model type
	switch model.ModelType {
	case "classification":
		// Use neurondb_classify
		inputJSON, _ := json.Marshal(input)
		inputStr := string(inputJSON)
		result, err := s.neurondbClient.Classify(ctx, inputStr, model.ModelPath)
		return result, err

	case "regression":
		// Use neurondb_regress
		result, err := s.neurondbClient.Regress(ctx, input, model.ModelPath)
		return result, err

	case "embedding":
		// Use neurondb_embed
		inputText := fmt.Sprintf("%v", input)
		result, err := s.neurondbClient.GenerateEmbedding(ctx, inputText, model.ModelPath)
		return result, err

	default:
		return nil, fmt.Errorf("unsupported model type: %s", model.ModelType)
	}
}

/* BatchInferModel performs batch inference using a model */
func (s *Service) BatchInferModel(ctx context.Context, modelID uuid.UUID, inputs []map[string]interface{}) ([]interface{}, error) {
	model, err := s.GetModel(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	// Use NeuronDB BatchPredict for efficient batch inference
	results, err := s.neurondbClient.BatchPredict(ctx, model.ModelPath, inputs)
	if err != nil {
		return nil, fmt.Errorf("batch inference failed: %w", err)
	}

	return results, nil
}

/* GetModelInfoFromNeuronDB retrieves model information from NeuronDB */
func (s *Service) GetModelInfoFromNeuronDB(ctx context.Context, modelPath string) (map[string]interface{}, error) {
	return s.neurondbClient.GetModelInfo(ctx, modelPath)
}

/* DeleteModelFromNeuronDB deletes a model from NeuronDB */
func (s *Service) DeleteModelFromNeuronDB(ctx context.Context, modelPath string) error {
	return s.neurondbClient.DeleteModel(ctx, modelPath)
}

/* CreateModelVersion creates a new version of a model */
func (s *Service) CreateModelVersion(ctx context.Context, modelID uuid.UUID, version string, modelPath string, metadata map[string]interface{}, changelog *string) (*ModelVersion, error) {
	modelVersion := &ModelVersion{
		ModelID:   modelID,
		Version:   version,
		ModelPath: modelPath,
		Metadata:  metadata,
		Changelog: changelog,
		CreatedAt: time.Now(),
	}
	
	metadataJSON, _ := json.Marshal(metadata)
	
	query := `
		INSERT INTO neuronip.model_versions 
		(id, model_id, version, model_path, metadata, changelog, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	
	err := s.pool.QueryRow(ctx, query,
		modelID, version, modelPath, metadataJSON, changelog, modelVersion.CreatedAt,
	).Scan(&modelVersion.ID, &modelVersion.CreatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create model version: %w", err)
	}
	
	// Update model version count
	s.pool.Exec(ctx, `UPDATE neuronip.model_registry SET version_count = version_count + 1 WHERE id = $1`, modelID)
	
	return modelVersion, nil
}

/* ListModelVersions lists versions for a model */
func (s *Service) ListModelVersions(ctx context.Context, modelID uuid.UUID) ([]ModelVersion, error) {
	query := `
		SELECT id, model_id, version, model_path, metadata, performance, changelog, created_by, created_at
		FROM neuronip.model_versions
		WHERE model_id = $1
		ORDER BY created_at DESC`
	
	rows, err := s.pool.Query(ctx, query, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to list model versions: %w", err)
	}
	defer rows.Close()
	
	versions := make([]ModelVersion, 0)
	for rows.Next() {
		var v ModelVersion
		var metadataJSON, performanceJSON json.RawMessage
		
		err := rows.Scan(
			&v.ID, &v.ModelID, &v.Version, &v.ModelPath, &metadataJSON, &performanceJSON,
			&v.Changelog, &v.CreatedBy, &v.CreatedAt,
		)
		if err != nil {
			continue
		}
		
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &v.Metadata)
		}
		if performanceJSON != nil {
			json.Unmarshal(performanceJSON, &v.Performance)
		}
		
		versions = append(versions, v)
	}
	
	return versions, nil
}

/* CreateExperiment creates an A/B testing experiment */
func (s *Service) CreateExperiment(ctx context.Context, experiment ModelExperiment) (*ModelExperiment, error) {
	experiment.ID = uuid.New()
	experiment.CreatedAt = time.Now()
	experiment.UpdatedAt = time.Now()
	
	if experiment.Status == "" {
		experiment.Status = "draft"
	}
	
	trafficSplitJSON, _ := json.Marshal(experiment.TrafficSplit)
	resultsJSON, _ := json.Marshal(experiment.Results)
	
	query := `
		INSERT INTO neuronip.model_experiments 
		(id, name, description, model_a_id, model_b_id, traffic_split, status,
		 start_date, end_date, results, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`
	
	err := s.pool.QueryRow(ctx, query,
		experiment.ID, experiment.Name, experiment.Description, experiment.ModelAID, experiment.ModelBID,
		trafficSplitJSON, experiment.Status, experiment.StartDate, experiment.EndDate, resultsJSON,
		experiment.CreatedBy, experiment.CreatedAt, experiment.UpdatedAt,
	).Scan(&experiment.ID, &experiment.CreatedAt, &experiment.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create experiment: %w", err)
	}
	
	return &experiment, nil
}

/* RecordModelMetric records a performance metric for a model */
func (s *Service) RecordModelMetric(ctx context.Context, modelID uuid.UUID, metricName string, metricValue float64, metadata map[string]interface{}) error {
	metadataJSON, _ := json.Marshal(metadata)
	
	query := `
		INSERT INTO neuronip.model_metrics 
		(id, model_id, metric_name, metric_value, metadata, timestamp)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())`
	
	_, err := s.pool.Exec(ctx, query, modelID, metricName, metricValue, metadataJSON)
	return err
}

/* GetModelMetrics retrieves metrics for a model */
func (s *Service) GetModelMetrics(ctx context.Context, modelID uuid.UUID, metricName *string, startTime *time.Time, endTime *time.Time, limit int) ([]ModelMetric, error) {
	query := `
		SELECT id, model_id, metric_name, metric_value, metadata, timestamp
		FROM neuronip.model_metrics
		WHERE model_id = $1`
	
	args := []interface{}{modelID}
	argIndex := 2
	
	if metricName != nil {
		query += fmt.Sprintf(" AND metric_name = $%d", argIndex)
		args = append(args, *metricName)
		argIndex++
	}
	
	if startTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *startTime)
		argIndex++
	}
	
	if endTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *endTime)
		argIndex++
	}
	
	query += " ORDER BY timestamp DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}
	
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get model metrics: %w", err)
	}
	defer rows.Close()
	
	metrics := make([]ModelMetric, 0)
	for rows.Next() {
		var m ModelMetric
		var metadataJSON json.RawMessage
		
		err := rows.Scan(&m.ID, &m.ModelID, &m.MetricName, &m.MetricValue, &metadataJSON, &m.Timestamp)
		if err != nil {
			continue
		}
		
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &m.Metadata)
		}
		
		metrics = append(metrics, m)
	}
	
	return metrics, nil
}

/* ModelVersion represents a model version */
type ModelVersion struct {
	ID          uuid.UUID              `json:"id"`
	ModelID     uuid.UUID              `json:"model_id"`
	Version     string                 `json:"version"`
	ModelPath   string                 `json:"model_path"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Performance map[string]interface{} `json:"performance,omitempty"`
	Changelog   *string                `json:"changelog,omitempty"`
	CreatedBy   *string                `json:"created_by,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

/* ModelExperiment represents an A/B testing experiment */
type ModelExperiment struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	ModelAID    uuid.UUID              `json:"model_a_id"`
	ModelBID    uuid.UUID              `json:"model_b_id"`
	TrafficSplit map[string]int        `json:"traffic_split"`
	Status      string                 `json:"status"`
	StartDate   *time.Time             `json:"start_date,omitempty"`
	EndDate     *time.Time             `json:"end_date,omitempty"`
	Results     map[string]interface{} `json:"results,omitempty"`
	CreatedBy   *string                `json:"created_by,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* ModelMetric represents a model performance metric */
type ModelMetric struct {
	ID          uuid.UUID              `json:"id"`
	ModelID     uuid.UUID              `json:"model_id"`
	MetricName  string                 `json:"metric_name"`
	MetricValue float64                `json:"metric_value"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}


/* calculateClassificationMetrics calculates classification metrics from predictions and ground truth */
func calculateClassificationMetrics(predictionsRaw interface{}, groundTruth []interface{}) map[string]interface{} {
	if len(groundTruth) == 0 {
		return map[string]interface{}{}
	}
	
	// Extract predicted labels from predictions
	// Predictions might be a single map with class probabilities or an array
	predictions := make([]interface{}, 0)
	
	if predMap, ok := predictionsRaw.(map[string]interface{}); ok {
		// Single prediction with class probabilities - find highest probability class
		if predClass, ok := predMap["class"]; ok {
			predictions = append(predictions, predClass)
		} else if predClass, ok := predMap["prediction"]; ok {
			predictions = append(predictions, predClass)
		} else {
			// Find class with highest probability
			var maxProb float64
			var maxClass interface{}
			for k, v := range predMap {
				if prob, ok := toFloat64(v); ok && prob > maxProb {
					maxProb = prob
					maxClass = k
				}
			}
			if maxClass != nil {
				predictions = append(predictions, maxClass)
			}
		}
	} else if predArray, ok := predictionsRaw.([]interface{}); ok {
		predictions = predArray
	}
	
	if len(predictions) == 0 || len(predictions) != len(groundTruth) {
		return map[string]interface{}{"error": "predictions and ground truth length mismatch"}
	}
	
	// Calculate metrics
	correct := 0
	tp := make(map[interface{}]float64) // true positives per class
	fp := make(map[interface{}]float64) // false positives per class
	fn := make(map[interface{}]float64) // false negatives per class
	
	for i := 0; i < len(predictions); i++ {
		pred := predictions[i]
		truth := groundTruth[i]
		
		if fmt.Sprintf("%v", pred) == fmt.Sprintf("%v", truth) {
			correct++
			tp[truth]++
		} else {
			fp[pred]++
			fn[truth]++
		}
	}
	
	accuracy := float64(correct) / float64(len(groundTruth))
	
	// Calculate precision, recall, F1 per class and macro-averaged
	var precisionSum, recallSum, f1Sum float64
	classCount := 0
	
	allClasses := make(map[interface{}]bool)
	for k := range tp {
		allClasses[k] = true
	}
	for k := range fp {
		allClasses[k] = true
	}
	for k := range fn {
		allClasses[k] = true
	}
	
	for class := range allClasses {
		tpVal := tp[class]
		fpVal := fp[class]
		fnVal := fn[class]
		
		precision := 0.0
		if tpVal+fpVal > 0 {
			precision = tpVal / (tpVal + fpVal)
		}
		
		recall := 0.0
		if tpVal+fnVal > 0 {
			recall = tpVal / (tpVal + fnVal)
		}
		
		f1 := 0.0
		if precision+recall > 0 {
			f1 = 2 * (precision * recall) / (precision + recall)
		}
		
		precisionSum += precision
		recallSum += recall
		f1Sum += f1
		classCount++
	}
	
	var precision, recall, f1Score float64
	if classCount > 0 {
		precision = precisionSum / float64(classCount)
		recall = recallSum / float64(classCount)
		f1Score = f1Sum / float64(classCount)
	}
	
	return map[string]interface{}{
		"accuracy":  accuracy,
		"precision": precision,
		"recall":    recall,
		"f1_score":  f1Score,
	}
}

/* calculateRegressionMetrics calculates regression metrics from predictions and ground truth */
func calculateRegressionMetrics(predictions, groundTruth []float64) map[string]interface{} {
	if len(predictions) != len(groundTruth) || len(predictions) == 0 {
		return map[string]interface{}{"error": "predictions and ground truth length mismatch"}
	}
	
	n := float64(len(predictions))
	var mse, mae, sumSquaredErrors, sumSquaredTotal float64
	var meanPred, meanTruth float64
	
	// Calculate means
	for i := 0; i < len(predictions); i++ {
		meanPred += predictions[i]
		meanTruth += groundTruth[i]
	}
	meanPred /= n
	meanTruth /= n
	
	// Calculate MSE, MAE, and R2 components
	for i := 0; i < len(predictions); i++ {
		diff := predictions[i] - groundTruth[i]
		mse += diff * diff
		mae += math.Abs(diff)
		sumSquaredErrors += diff * diff
		sumSquaredTotal += (groundTruth[i] - meanTruth) * (groundTruth[i] - meanTruth)
	}
	
	mse /= n
	mae /= n
	rmse := math.Sqrt(mse)
	
	// Calculate R2 (coefficient of determination)
	r2 := 0.0
	if sumSquaredTotal > 0 {
		r2 = 1 - (sumSquaredErrors / sumSquaredTotal)
	}
	
	return map[string]interface{}{
		"rmse": rmse,
		"mae":  mae,
		"mse":  mse,
		"r2":   r2,
	}
}

/* toFloat64 converts various numeric types to float64 */
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case string:
		// Try parsing as float
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

/* EnsureFeatureVectorIndex ensures a vector index exists for feature vectors used by models */
func (s *Service) EnsureFeatureVectorIndex(ctx context.Context, tableName string, columnName string) error {
	if s.neurondbClient == nil {
		return fmt.Errorf("NeuronDB client not configured")
	}

	// Check if index already exists
	indexName := fmt.Sprintf("idx_%s_%s_vector", tableName, columnName)
	checkQuery := `
		SELECT COUNT(*) FROM pg_indexes 
		WHERE indexname = $1 AND tablename = $2`
	
	var count int
	err := s.pool.QueryRow(ctx, checkQuery, indexName, tableName).Scan(&count)
	if err == nil && count > 0 {
		return nil // Index already exists
	}

	// Get row count to determine index parameters
	var rowCount int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s IS NOT NULL`, tableName, columnName)
	err = s.pool.QueryRow(ctx, countQuery).Scan(&rowCount)
	if err != nil {
		return fmt.Errorf("failed to get row count: %w", err)
	}

	// Choose index type and parameters based on data size
	var indexType string
	var options map[string]interface{}
	
	if rowCount < 10000 {
		// Small dataset - use IVF
		indexType = "ivf"
		options = map[string]interface{}{
			"lists": 100,
		}
	} else {
		// Large dataset - use HNSW
		indexType = "hnsw"
		options = map[string]interface{}{
			"m":              16,
			"ef_construction": 64,
		}
	}

	// Create index using NeuronDB client
	err = s.neurondbClient.CreateVectorIndex(ctx, indexName, tableName, columnName, indexType, options)
	if err != nil {
		return fmt.Errorf("failed to create vector index: %w", err)
	}

	return nil
}

/* OptimizeVectorIndex tunes an existing vector index based on usage patterns */
func (s *Service) OptimizeVectorIndex(ctx context.Context, indexName string, tableName string, columnName string) error {
	if s.mcpClient == nil {
		return fmt.Errorf("MCP client not configured")
	}

	// Get index status first
	status, err := s.mcpClient.IndexStatus(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to get index status: %w", err)
	}

	// Check if index needs tuning based on status
	// This is a simplified check - in production, analyze actual usage patterns
	if statusIndexType, ok := status["index_type"].(string); ok {
		indexType := statusIndexType // Use the status index type
		if statusIndexType == "hnsw" || indexType == "hnsw" {
			// Tune HNSW index - need to drop and recreate with new parameters
			// In production, would use TuneHNSWIndex if available
			dropQuery := fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)
			_, err = s.pool.Exec(ctx, dropQuery)
			if err != nil {
				return fmt.Errorf("failed to drop index: %w", err)
			}

			// Recreate with optimized parameters
			_, err = s.mcpClient.CreateHNSWIndex(ctx, indexName, tableName, columnName, map[string]interface{}{
				"m":              32, // Increase for better recall
				"ef_construction": 128, // Increase for better quality
			})
			return err
		} else if statusIndexType == "ivf" || indexType == "ivf" {
			// Tune IVF index
			dropQuery := fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)
			_, err = s.pool.Exec(ctx, dropQuery)
			if err != nil {
				return fmt.Errorf("failed to drop index: %w", err)
			}

			// Recreate with optimized parameters
			_, err = s.mcpClient.CreateIVFIndex(ctx, indexName, tableName, columnName, map[string]interface{}{
				"lists": 200, // Increase for larger datasets
			})
			return err
		}
	}

	return nil
}
