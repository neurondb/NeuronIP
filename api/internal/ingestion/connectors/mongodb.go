package connectors

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* MongoDBConnector implements the Connector interface for MongoDB */
type MongoDBConnector struct {
	*ingestion.BaseConnector
	client *mongo.Client
}

/* NewMongoDBConnector creates a new MongoDB connector */
func NewMongoDBConnector() *MongoDBConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "mongodb",
		Name:        "MongoDB",
		Description: "MongoDB NoSQL database connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("mongodb", metadata)

	return &MongoDBConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to MongoDB */
func (m *MongoDBConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	uri, _ := config["uri"].(string)
	if uri == "" {
		host, _ := config["host"].(string)
		port, _ := config["port"].(float64)
		if port == 0 {
			port = 27017
		}
		user, _ := config["user"].(string)
		password, _ := config["password"].(string)
		database, _ := config["database"].(string)

		if host == "" {
			return fmt.Errorf("host or uri is required")
		}

		if user != "" && password != "" {
			uri = fmt.Sprintf("mongodb://%s:%s@%s:%.0f/%s", user, password, host, port, database)
		} else {
			uri = fmt.Sprintf("mongodb://%s:%.0f/%s", host, port, database)
		}
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (m *MongoDBConnector) Disconnect(ctx context.Context) error {
	if m.client != nil {
		m.client.Disconnect(ctx)
	}
	m.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (m *MongoDBConnector) TestConnection(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("not connected")
	}
	return m.client.Ping(ctx, nil)
}

/* DiscoverSchema discovers MongoDB schema */
func (m *MongoDBConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if m.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	tables := []ingestion.TableSchema{}

	// List all databases
	databases, err := m.client.ListDatabaseNames(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	for _, dbName := range databases {
		if dbName == "admin" || dbName == "local" || dbName == "config" {
			continue
		}

		db := m.client.Database(dbName)
		collections, err := db.ListCollectionNames(ctx, nil)
		if err != nil {
			continue
		}

		for _, collName := range collections {
			// Sample a few documents to infer schema
			coll := db.Collection(collName)
			cursor, err := coll.Find(ctx, nil, options.Find().SetLimit(10))
			if err != nil {
				continue
			}

			columns := []ingestion.ColumnSchema{}
			fieldMap := make(map[string]bool)

			for cursor.Next(ctx) {
				var doc map[string]interface{}
				if err := cursor.Decode(&doc); err != nil {
					continue
				}

				for key := range doc {
					if !fieldMap[key] {
						fieldMap[key] = true
						columns = append(columns, ingestion.ColumnSchema{
							Name:     key,
							DataType: "jsonb", // MongoDB stores as JSON
							Nullable: true,
						})
					}
				}
			}
			cursor.Close(ctx)

			tables = append(tables, ingestion.TableSchema{
				Name:    fmt.Sprintf("%s.%s", dbName, collName),
				Columns: columns,
			})
		}
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (m *MongoDBConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if m.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	// Get schema to determine collections
	schema, err := m.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	tables := options.Tables
	if len(tables) == 0 {
		for _, table := range schema.Tables {
			tables = append(tables, table.Name)
		}
	}

	for _, tableName := range tables {
		// Parse database.collection
		var dbName, collName string
		if idx := len(tableName) - 1; idx >= 0 {
			for i := idx; i >= 0; i-- {
				if tableName[i] == '.' {
					dbName = tableName[:i]
					collName = tableName[i+1:]
					break
				}
			}
		}

		if dbName == "" || collName == "" {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   tableName,
				Message: "invalid table name format (expected database.collection)",
			})
			continue
		}

		coll := m.client.Database(dbName).Collection(collName)
		count, err := coll.CountDocuments(ctx, nil)
		if err != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   tableName,
				Message: err.Error(),
			})
			continue
		}

		result.RowsSynced += count
		result.TablesSynced = append(result.TablesSynced, tableName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
