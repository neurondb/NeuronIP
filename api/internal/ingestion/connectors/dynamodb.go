package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* DynamoDBConnector implements the Connector interface for Amazon DynamoDB */
type DynamoDBConnector struct {
	*ingestion.BaseConnector
	client *dynamodb.Client
}

/* NewDynamoDBConnector creates a new DynamoDB connector */
func NewDynamoDBConnector() *DynamoDBConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "dynamodb",
		Name:        "Amazon DynamoDB",
		Description: "Amazon DynamoDB NoSQL database connector",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("dynamodb", metadata)

	return &DynamoDBConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to DynamoDB */
func (d *DynamoDBConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	region, _ := config["region"].(string)
	if region == "" {
		region = "us-east-1"
	}

	accessKeyID, _ := config["access_key_id"].(string)
	secretAccessKey, _ := config["secret_access_key"].(string)

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}

	if accessKeyID != "" && secretAccessKey != "" {
		// Use static credentials provider
		staticCreds := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
		opts = append(opts, awsconfig.WithCredentialsProvider(staticCreds))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	d.client = dynamodb.NewFromConfig(cfg)
	d.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (d *DynamoDBConnector) Disconnect(ctx context.Context) error {
	d.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (d *DynamoDBConnector) TestConnection(ctx context.Context) error {
	if d.client == nil {
		return fmt.Errorf("not connected")
	}

	// List tables to test connection
	_, err := d.client.ListTables(ctx, &dynamodb.ListTablesInput{Limit: aws.Int32(1)})
	return err
}

/* DiscoverSchema discovers DynamoDB schema */
func (d *DynamoDBConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if d.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	tables := []ingestion.TableSchema{}

	// List all tables
	paginator := dynamodb.NewListTablesPaginator(d.client, &dynamodb.ListTablesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tables: %w", err)
		}

		for _, tableName := range page.TableNames {
			// Get table description
			desc, err := d.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				continue
			}

			columns := []ingestion.ColumnSchema{}
			for _, attr := range desc.Table.AttributeDefinitions {
				columns = append(columns, ingestion.ColumnSchema{
					Name:     *attr.AttributeName,
					DataType: string(attr.AttributeType),
					Nullable: false, // DynamoDB keys are not nullable
				})
			}

			// Sample items to discover additional attributes
			scanInput := &dynamodb.ScanInput{
				TableName:        aws.String(tableName),
				Limit:            aws.Int32(10),
				AttributesToGet:  nil, // Get all attributes
			}

			scanResult, err := d.client.Scan(ctx, scanInput)
			if err == nil {
				knownAttrs := make(map[string]bool)
				for _, col := range columns {
					knownAttrs[col.Name] = true
				}

				// Discover additional attributes from sample items
				for _, item := range scanResult.Items {
					for key := range item {
						if !knownAttrs[key] {
							knownAttrs[key] = true
							columns = append(columns, ingestion.ColumnSchema{
								Name:     key,
								DataType: "S", // Default to string for discovered attributes
								Nullable: true,
							})
						}
					}
				}
			}

			tables = append(tables, ingestion.TableSchema{
				Name:    tableName,
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
func (d *DynamoDBConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if d.client == nil {
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

	for _, tableName := range tables {
		// Get item count (approximate)
		desc, err := d.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   tableName,
				Message: err.Error(),
			})
			continue
		}

		if desc.Table.ItemCount != nil {
			result.RowsSynced += *desc.Table.ItemCount
		}
		result.TablesSynced = append(result.TablesSynced, tableName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
