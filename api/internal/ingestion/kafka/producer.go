package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

/* KafkaProducer provides Kafka event publishing functionality */
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

/* NewKafkaProducer creates a new Kafka producer */
func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	
	return &KafkaProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

/* PublishIngestionEvent publishes an ingestion event to Kafka */
func (kp *KafkaProducer) PublishIngestionEvent(ctx context.Context, eventType string, data map[string]interface{}) error {
	event := map[string]interface{}{
		"event_type": eventType,
		"timestamp":  fmt.Sprintf("%d", ctx.Value("timestamp")),
		"data":       data,
	}
	
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	message := &sarama.ProducerMessage{
		Topic: kp.topic,
		Value: sarama.ByteEncoder(eventJSON),
	}
	
	partition, offset, err := kp.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	
	_ = partition
	_ = offset
	
	return nil
}

/* Close closes the Kafka producer */
func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}
