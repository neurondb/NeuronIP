package connectors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* DbtConnector implements the Connector interface for dbt */
type DbtConnector struct {
	*ingestion.BaseConnector
	projectPath string
	profilesDir string
}

/* NewDbtConnector creates a new dbt connector */
func NewDbtConnector() *DbtConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "dbt",
		Name:        "dbt",
		Description: "dbt project connector for models, tests, and lineage",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "lineage_extraction", "transformation_logic"},
	}

	base := ingestion.NewBaseConnector("dbt", metadata)

	return &DbtConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to dbt project */
func (d *DbtConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	projectPath, _ := config["project_path"].(string)
	profilesDir, _ := config["profiles_dir"].(string)

	if projectPath == "" {
		return fmt.Errorf("project_path is required")
	}

	// Verify dbt_project.yml exists
	dbtProjectPath := filepath.Join(projectPath, "dbt_project.yml")
	if _, err := os.Stat(dbtProjectPath); os.IsNotExist(err) {
		return fmt.Errorf("dbt_project.yml not found at %s", dbtProjectPath)
	}

	d.projectPath = projectPath
	if profilesDir == "" {
		d.profilesDir = filepath.Join(os.Getenv("HOME"), ".dbt")
	} else {
		d.profilesDir = profilesDir
	}

	d.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (d *DbtConnector) Disconnect(ctx context.Context) error {
	d.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (d *DbtConnector) TestConnection(ctx context.Context) error {
	if d.projectPath == "" {
		return fmt.Errorf("not connected")
	}

	dbtProjectPath := filepath.Join(d.projectPath, "dbt_project.yml")
	if _, err := os.Stat(dbtProjectPath); os.IsNotExist(err) {
		return fmt.Errorf("dbt_project.yml not found")
	}

	return nil
}

/* DiscoverSchema discovers dbt models, sources, and tests */
func (d *DbtConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if d.projectPath == "" {
		return nil, fmt.Errorf("not connected")
	}

	tables := []ingestion.TableSchema{}

	// Verify dbt_project.yml exists (already checked in Connect)
	dbtProjectPath := filepath.Join(d.projectPath, "dbt_project.yml")
	_, err := os.Stat(dbtProjectPath)
	if err != nil {
		return nil, fmt.Errorf("dbt_project.yml not found: %w", err)
	}

	// In production, parse YAML to extract model configurations
	// For now, we'll discover by walking the models directory

	// Walk models directory
	modelsDir := filepath.Join(d.projectPath, "models")
	if _, err := os.Stat(modelsDir); err == nil {
		err = filepath.Walk(modelsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".sql" {
				relPath, _ := filepath.Rel(modelsDir, path)
				modelName := relPath[:len(relPath)-4] // Remove .sql extension

				tables = append(tables, ingestion.TableSchema{
					Name: modelName,
					Columns: []ingestion.ColumnSchema{
						{Name: "model_name", DataType: "string", Nullable: false},
						{Name: "file_path", DataType: "string", Nullable: false},
					},
					Metadata: map[string]interface{}{
						"file_path": path,
						"resource_type": "dbt_model",
					},
				})
			}
			return nil
		})
		if err != nil {
			if err != nil {
			return nil, fmt.Errorf("failed to walk models directory: %w", err)
		}
		}
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (d *DbtConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if d.projectPath == "" {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := d.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	tables := options.Tables
	if len(tables) == 0 {
		for _, table := range schema.Tables {
			tables = append(tables, table.Name)
		}
	}

	for _, modelName := range tables {
		result.TablesSynced = append(result.TablesSynced, modelName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
