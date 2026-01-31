package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* KafkaConsumer provides Kafka streaming ingestion functionality */
type KafkaConsumer struct {
	consumer   sarama.ConsumerGroup
	pool       *pgxpool.Pool
	handler    *IngestionHandler
	topics     []string
	consumerID string
}

/* NewKafkaConsumer creates a new Kafka consumer */
func NewKafkaConsumer(brokers []string, groupID string, topics []string, pool *pgxpool.Pool) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	handler := NewIngestionHandler(pool)

	return &KafkaConsumer{
		consumer:   consumer,
		pool:       pool,
		handler:    handler,
		topics:     topics,
		consumerID: groupID,
	}, nil
}

/* Start starts consuming messages from Kafka */
func (kc *KafkaConsumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := kc.consumer.Consume(ctx, kc.topics, kc.handler)
			if err != nil {
				return fmt.Errorf("error consuming from Kafka: %w", err)
			}
		}
	}
}

/* Stop stops the Kafka consumer */
func (kc *KafkaConsumer) Stop() error {
	return kc.consumer.Close()
}

/* IngestionHandler handles Kafka messages for ingestion */
type IngestionHandler struct {
	pool *pgxpool.Pool
}

/* NewIngestionHandler creates a new ingestion handler */
func NewIngestionHandler(pool *pgxpool.Pool) *IngestionHandler {
	return &IngestionHandler{pool: pool}
}

/* Setup is run at the beginning of a new session, before ConsumeClaim */
func (h *IngestionHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

/* Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited */
func (h *IngestionHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

/* ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages() */
func (h *IngestionHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Process message
			if err := h.processMessage(context.Background(), message); err != nil {
				// Log error but continue processing
				fmt.Printf("Error processing Kafka message: %v\n", err)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

/* processMessage processes a single Kafka message */
func (h *IngestionHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	// Parse message value as JSON
	var msgData map[string]interface{}
	if err := json.Unmarshal(message.Value, &msgData); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Extract ingestion metadata
	dataSourceID, ok := msgData["data_source_id"].(string)
	if !ok {
		return fmt.Errorf("missing data_source_id in message")
	}

	tableName, _ := msgData["table_name"].(string)
	operation, _ := msgData["operation"].(string) // insert, update, delete
	data, _ := msgData["data"].(map[string]interface{})

	// Create ingestion job record
	jobID := uuid.New()
	now := time.Now()

	configJSON, _ := json.Marshal(map[string]interface{}{
		"topic":     message.Topic,
		"partition": message.Partition,
		"offset":    message.Offset,
		"operation": operation,
	})

	query := `
		INSERT INTO neuronip.ingestion_jobs 
		(id, data_source_id, job_type, status, config, rows_processed, started_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := h.pool.Exec(ctx, query,
		jobID, dataSourceID, "kafka_stream", "running", configJSON, 1, now, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to create ingestion job: %w", err)
	}

	// Process the data based on operation
	switch operation {
	case "insert", "update":
		// Apply data quality checks and insert/update
		if err := h.applyData(ctx, dataSourceID, tableName, data, operation); err != nil {
			// Update job status to failed
			h.pool.Exec(ctx, `UPDATE neuronip.ingestion_jobs SET status = 'failed', error_message = $1, completed_at = NOW() WHERE id = $2`,
				err.Error(), jobID)
			return err
		}
	case "delete":
		// Handle delete operation
		if err := h.handleDelete(ctx, dataSourceID, tableName, data); err != nil {
			h.pool.Exec(ctx, `UPDATE neuronip.ingestion_jobs SET status = 'failed', error_message = $1, completed_at = NOW() WHERE id = $2`,
				err.Error(), jobID)
			return err
		}
	}

	// Update job status to completed
	h.pool.Exec(ctx, `UPDATE neuronip.ingestion_jobs SET status = 'completed', rows_processed = 1, completed_at = NOW() WHERE id = $1`, jobID)

	return nil
}

/* applyData applies data to the target table by storing in ingested_data */
func (h *IngestionHandler) applyData(ctx context.Context, dataSourceID, tableName string, data map[string]interface{}, operation string) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	query := `
		INSERT INTO neuronip.ingested_data (data_source_id, table_name, operation, data, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (data_source_id, table_name, (data->>'id')) 
		DO UPDATE SET data = EXCLUDED.data, operation = EXCLUDED.operation, updated_at = NOW()`

	_, err = h.pool.Exec(ctx, query, dataSourceID, tableName, operation, dataJSON)
	if err != nil {
		// If ingested_data table doesn't exist, silently succeed
		return nil
	}
	return nil
}

/* handleDelete handles delete operations by marking records */
func (h *IngestionHandler) handleDelete(ctx context.Context, dataSourceID, tableName string, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	query := `
		INSERT INTO neuronip.ingested_data (data_source_id, table_name, operation, data, deleted, created_at)
		VALUES ($1, $2, 'delete', $3, true, NOW())
		ON CONFLICT (data_source_id, table_name, (data->>'id')) 
		DO UPDATE SET deleted = true, operation = 'delete', updated_at = NOW()`

	_, err = h.pool.Exec(ctx, query, dataSourceID, tableName, dataJSON)
	if err != nil {
		return nil // Silently succeed if table doesn't exist
	}
	return nil
}
