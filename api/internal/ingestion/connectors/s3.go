package connectors

import (
	"context"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* S3Connector implements the Connector interface for AWS S3 */
type S3Connector struct {
	*ingestion.BaseConnector
	client *s3.Client
}

/* NewS3Connector creates a new S3 connector */
func NewS3Connector() *S3Connector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "s3",
		Name:        "Amazon S3",
		Description: "AWS S3 object storage connector for bucket and object discovery",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("s3", metadata)

	return &S3Connector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to S3 */
func (s *S3Connector) Connect(ctx context.Context, config map[string]interface{}) error {
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
		staticCreds := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
		opts = append(opts, awsconfig.WithCredentialsProvider(staticCreds))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s.client = s3.NewFromConfig(cfg)
	s.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (s *S3Connector) Disconnect(ctx context.Context) error {
	s.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (s *S3Connector) TestConnection(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("not connected")
	}

	// Try to list buckets
	_, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	return err
}

/* DiscoverSchema discovers S3 schema (buckets and objects) */
func (s *S3Connector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if s.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// List all buckets
	listBucketsOutput, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	tables := []ingestion.TableSchema{}

	for _, bucket := range listBucketsOutput.Buckets {
		bucketName := *bucket.Name

		// List objects in bucket (with pagination)
		paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
			Bucket: &bucketName,
		})

		// Group objects by prefix (simulate "tables")
		prefixMap := make(map[string]bool)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, obj := range page.Contents {
				key := *obj.Key
				// Extract prefix (everything before last /)
				prefix := ""
				if idx := len(key) - 1; idx >= 0 {
					for i := idx; i >= 0; i-- {
						if key[i] == '/' {
							prefix = key[:i]
							break
						}
					}
				}

				if prefix != "" && !prefixMap[prefix] {
					prefixMap[prefix] = true
				}
			}
		}

		// Create "tables" for each prefix
		for prefix := range prefixMap {
			tables = append(tables, ingestion.TableSchema{
				Name: fmt.Sprintf("%s/%s", bucketName, prefix),
				Columns: []ingestion.ColumnSchema{
					{Name: "key", DataType: "text"},
					{Name: "size", DataType: "bigint"},
					{Name: "last_modified", DataType: "timestamp"},
					{Name: "etag", DataType: "text"},
				},
			})
		}

		// Also add bucket root as a table
		if len(prefixMap) == 0 {
			tables = append(tables, ingestion.TableSchema{
				Name: bucketName,
				Columns: []ingestion.ColumnSchema{
					{Name: "key", DataType: "text"},
					{Name: "size", DataType: "bigint"},
					{Name: "last_modified", DataType: "timestamp"},
					{Name: "etag", DataType: "text"},
				},
			})
		}
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (s *S3Connector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if s.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	// Get schema to determine buckets/prefixes
	schema, err := s.DiscoverSchema(ctx)
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
		// Parse bucket/prefix
		var bucketName, prefix string
		if idx := len(tableName) - 1; idx >= 0 {
			for i := idx; i >= 0; i-- {
				if tableName[i] == '/' {
					bucketName = tableName[:i]
					prefix = tableName[i+1:]
					break
				}
			}
		}

		if bucketName == "" {
			bucketName = tableName
		}

		// Count objects
		count := int64(0)
		paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
			Bucket: &bucketName,
			Prefix: &prefix,
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				result.Errors = append(result.Errors, ingestion.SyncError{
					Table:   tableName,
					Message: err.Error(),
				})
				break
			}
			count += int64(len(page.Contents))
		}

		result.RowsSynced += count
		result.TablesSynced = append(result.TablesSynced, tableName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
