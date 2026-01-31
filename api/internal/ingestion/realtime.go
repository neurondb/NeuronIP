package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* RealtimeIngestionService provides real-time ingestion functionality */
type RealtimeIngestionService struct {
	pool *pgxpool.Pool
}

/* NewRealtimeIngestionService creates a new real-time ingestion service */
func NewRealtimeIngestionService(pool *pgxpool.Pool) *RealtimeIngestionService {
	return &RealtimeIngestionService{pool: pool}
}

/* IngestEvent ingests a real-time event */
func (ris *RealtimeIngestionService) IngestEvent(ctx context.Context, event RealtimeEvent) error {
	event.ID = uuid.New()
	event.Timestamp = time.Now()

	payloadJSON, _ := json.Marshal(event.Payload)

	query := `
		INSERT INTO neuronip.realtime_events 
		(id, event_type, source, payload, timestamp, processed)
		VALUES ($1, $2, $3, $4, $5, false)`

	_, err := ris.pool.Exec(ctx, query,
		event.ID, event.EventType, event.Source, payloadJSON, event.Timestamp,
	)
	return err
}

/* ProcessEvents processes pending real-time events */
func (ris *RealtimeIngestionService) ProcessEvents(ctx context.Context, batchSize int) (int, error) {
	query := `
		UPDATE neuronip.realtime_events
		SET processed = true, processed_at = NOW()
		WHERE id IN (
			SELECT id FROM neuronip.realtime_events
			WHERE processed = false
			ORDER BY timestamp ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id`

	rows, err := ris.pool.Query(ctx, query, batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to process events: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	return count, nil
}

/* RealtimeEvent represents a real-time event */
type RealtimeEvent struct {
	ID        uuid.UUID              `json:"id"`
	EventType string                 `json:"event_type"`
	Source    string                 `json:"source"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}
