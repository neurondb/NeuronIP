package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* KafkaConnector implements the Connector interface for Apache Kafka */
type KafkaConnector struct {
	*ingestion.BaseConnector
	adminClient sarama.ClusterAdmin
	brokers     []string
}

/* NewKafkaConnector creates a new Kafka connector */
func NewKafkaConnector() *KafkaConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "kafka",
		Name:        "Apache Kafka",
		Description: "Apache Kafka message broker connector for topics and partitions",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "cdc"},
	}

	base := ingestion.NewBaseConnector("kafka", metadata)

	return &KafkaConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Kafka */
func (k *KafkaConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	brokers, _ := config["brokers"].([]interface{})
	if len(brokers) == 0 {
		return fmt.Errorf("brokers list is required")
	}

	brokersStr := make([]string, len(brokers))
	for i, b := range brokers {
		brokersStr[i] = b.(string)
	}

	k.brokers = brokersStr

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_6_0_0

	adminClient, err := sarama.NewClusterAdmin(brokersStr, cfg)
	if err != nil {
		return fmt.Errorf("failed to create Kafka admin client: %w", err)
	}

	k.adminClient = adminClient
	k.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (k *KafkaConnector) Disconnect(ctx context.Context) error {
	if k.adminClient != nil {
		k.adminClient.Close()
	}
	k.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (k *KafkaConnector) TestConnection(ctx context.Context) error {
	if k.adminClient == nil {
		return fmt.Errorf("not connected")
	}
	// List topics to test connection
	_, err := k.adminClient.ListTopics()
	return err
}

/* DiscoverSchema discovers Kafka topics */
func (k *KafkaConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if k.adminClient == nil {
		return nil, fmt.Errorf("not connected")
	}

	topics, err := k.adminClient.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	tables := []ingestion.TableSchema{}

	for topicName, topicDetail := range topics {
		// Each topic is represented as a "table"
		columns := []ingestion.ColumnSchema{
			{Name: "key", DataType: "bytes", Nullable: true},
			{Name: "value", DataType: "bytes", Nullable: true},
			{Name: "partition", DataType: "int", Nullable: false},
			{Name: "offset", DataType: "long", Nullable: false},
			{Name: "timestamp", DataType: "timestamp", Nullable: false},
		}

		tables = append(tables, ingestion.TableSchema{
			Name:    topicName,
			Columns: columns,
			Metadata: map[string]interface{}{
				"num_partitions": topicDetail.NumPartitions,
			},
		})
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (k *KafkaConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if k.adminClient == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := k.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	tables := options.Tables
	if len(tables) == 0 {
		for _, table := range schema.Tables {
			tables = append(tables, table.Name)
		}
	}

	// For Kafka, we can't easily count messages without consuming them
	// This is a simplified implementation
	for _, topicName := range tables {
		result.TablesSynced = append(result.TablesSynced, topicName)
		// Note: Actual message count would require consuming from each partition
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
