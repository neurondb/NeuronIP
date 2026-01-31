package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	postgresEventsSource = "postgres_events"
	databaseSink         = "database"
	partitionDefault     = "default"
	pollInterval         = 2 * time.Second
	batchSize            = 100
)

/* Runner holds running pipeline cancel functions */
type Runner struct {
	pool    *pgxpool.Pool
	cancels map[uuid.UUID]context.CancelFunc
	mu      sync.Mutex
}

/* NewRunner creates a new pipeline runner */
func NewRunner(pool *pgxpool.Pool) *Runner {
	return &Runner{pool: pool, cancels: make(map[uuid.UUID]context.CancelFunc)}
}

/* Start starts a pipeline worker in the background. Caller should call Stop to cancel. */
func (r *Runner) Start(ctx context.Context, engine *StreamingEngine, pipelineID uuid.UUID) error {
	r.mu.Lock()
	if _, exists := r.cancels[pipelineID]; exists {
		r.mu.Unlock()
		return fmt.Errorf("pipeline %s already running", pipelineID)
	}
	workerCtx, cancel := context.WithCancel(ctx)
	r.cancels[pipelineID] = cancel
	r.mu.Unlock()

	go func() {
		defer func() {
			r.mu.Lock()
			delete(r.cancels, pipelineID)
			r.mu.Unlock()
		}()
		runPipeline(workerCtx, engine.pool, pipelineID)
		// Mark stopped when context is done
		_, _ = engine.pool.Exec(workerCtx, `UPDATE neuronip.streaming_pipelines SET status = 'stopped', updated_at = NOW() WHERE id = $1`, pipelineID)
	}()
	return nil
}

/* Stop cancels a running pipeline worker */
func (r *Runner) Stop(pipelineID uuid.UUID) {
	r.mu.Lock()
	cancel, ok := r.cancels[pipelineID]
	r.mu.Unlock()
	if ok && cancel != nil {
		cancel()
	}
}

func runPipeline(ctx context.Context, pool *pgxpool.Pool, pipelineID uuid.UUID) {
	pl, err := getPipelineByID(ctx, pool, pipelineID)
	if err != nil || pl == nil {
		_, _ = pool.Exec(ctx, `UPDATE neuronip.streaming_pipelines SET status = 'failed', updated_at = NOW() WHERE id = $1`, pipelineID)
		return
	}
	if len(pl.Sources) == 0 {
		return
	}
	src := pl.Sources[0]
	if src.Type != postgresEventsSource {
		// Only postgres_events is implemented; others no-op
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(pollInterval)
			}
		}
	}

	lastOffset := loadCheckpoint(ctx, pool, pipelineID)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		events, maxID, err := pollStreamEvents(ctx, pool, lastOffset, batchSize)
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}
		if len(events) == 0 {
			time.Sleep(pollInterval)
			continue
		}
		transformed := applyTransforms(pl.Transformations, events)
		if err := writeToSink(ctx, pool, pipelineID, pl.Sinks, transformed); err != nil {
			time.Sleep(pollInterval)
			continue
		}
		lastOffset = maxID
		saveCheckpoint(ctx, pool, pipelineID, lastOffset)
	}
}

func getPipelineByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Pipeline, error) {
	query := `SELECT id, pipeline_name, description, sources, transformations, sinks, config, status, created_at, updated_at
			  FROM neuronip.streaming_pipelines WHERE id = $1`
	var sourcesRaw, transformationsRaw, sinksRaw, configRaw json.RawMessage
	var pl Pipeline
	err := pool.QueryRow(ctx, query, id).Scan(
		&pl.ID, &pl.Name, &pl.Description, &sourcesRaw, &transformationsRaw, &sinksRaw, &configRaw,
		&pl.Status, &pl.CreatedAt, &pl.UpdatedAt,
	)
	if err != nil {
		// Fallback when 065 not applied: columns description/sources/transformations/sinks/config may be missing
		fallback := `SELECT id, pipeline_name, status, created_at, updated_at FROM neuronip.streaming_pipelines WHERE id = $1`
		if fallbackErr := pool.QueryRow(ctx, fallback, id).Scan(&pl.ID, &pl.Name, &pl.Status, &pl.CreatedAt, &pl.UpdatedAt); fallbackErr != nil {
			return nil, err
		}
		pl.Sources = nil
		pl.Transformations = nil
		pl.Sinks = nil
		return &pl, nil
	}
	if sourcesRaw != nil {
		_ = json.Unmarshal(sourcesRaw, &pl.Sources)
	}
	if transformationsRaw != nil {
		_ = json.Unmarshal(transformationsRaw, &pl.Transformations)
	}
	if sinksRaw != nil {
		_ = json.Unmarshal(sinksRaw, &pl.Sinks)
	}
	if configRaw != nil {
		_ = json.Unmarshal(configRaw, &pl.Config)
	}
	return &pl, nil
}

func loadCheckpoint(ctx context.Context, pool *pgxpool.Pool, pipelineID uuid.UUID) int64 {
	var offset int64
	query := `SELECT offset FROM neuronip.stream_checkpoints WHERE pipeline_id = $1 AND partition_id = $2`
	err := pool.QueryRow(ctx, query, pipelineID, partitionDefault).Scan(&offset)
	if err != nil {
		return 0
	}
	return offset
}

func saveCheckpoint(ctx context.Context, pool *pgxpool.Pool, pipelineID uuid.UUID, offset int64) {
	query := `
		INSERT INTO neuronip.stream_checkpoints (id, pipeline_id, partition_id, offset, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		ON CONFLICT (pipeline_id, partition_id) DO UPDATE SET offset = $3, updated_at = NOW()`
	_, _ = pool.Exec(ctx, query, pipelineID, partitionDefault, offset)
}

func pollStreamEvents(ctx context.Context, pool *pgxpool.Pool, afterID int64, limit int) ([]map[string]interface{}, int64, error) {
	query := `SELECT id, event_type, source_table, payload, created_at
			  FROM neuronip.stream_events WHERE id > $1 ORDER BY id LIMIT $2`
	rows, err := pool.Query(ctx, query, afterID, limit)
	if err != nil {
		return nil, afterID, err
	}
	defer rows.Close()
	var events []map[string]interface{}
	var maxID int64
	for rows.Next() {
		var id int64
		var eventType, sourceTable string
		var payload json.RawMessage
		var createdAt time.Time
		if err := rows.Scan(&id, &eventType, &sourceTable, &payload, &createdAt); err != nil {
			continue
		}
		if id > maxID {
			maxID = id
		}
		ev := map[string]interface{}{"id": id, "event_type": eventType, "source_table": sourceTable, "created_at": createdAt}
		if len(payload) > 0 {
			var p map[string]interface{}
			_ = json.Unmarshal(payload, &p)
			ev["payload"] = p
		} else {
			ev["payload"] = map[string]interface{}{}
		}
		events = append(events, ev)
	}
	return events, maxID, nil
}

func applyTransforms(transforms []Transformation, events []map[string]interface{}) []map[string]interface{} {
	out := events
	for _, t := range transforms {
		switch t.Type {
		case "filter":
			out = filterEvents(out, t.Config)
		case "map":
			out = mapEvents(out, t.Config)
		default:
			// no-op for aggregate/window/join in single-batch mode
		}
	}
	return out
}

func filterEvents(events []map[string]interface{}, config map[string]interface{}) []map[string]interface{} {
	field, _ := config["field"].(string)
	value, ok := config["value"]
	if !ok || field == "" {
		return events
	}
	out := make([]map[string]interface{}, 0, len(events))
	for _, ev := range events {
		if p, ok := ev["payload"].(map[string]interface{}); ok {
			if p[field] == value {
				out = append(out, ev)
			}
		}
	}
	return out
}

func mapEvents(events []map[string]interface{}, config map[string]interface{}) []map[string]interface{} {
	// Simple map: include only specified keys from payload
	keys, ok := config["include_keys"].([]interface{})
	if !ok {
		return events
	}
	out := make([]map[string]interface{}, 0, len(events))
	for _, ev := range events {
		payload, _ := ev["payload"].(map[string]interface{})
		mapped := make(map[string]interface{})
		for _, k := range keys {
			if s, ok := k.(string); ok && payload[s] != nil {
				mapped[s] = payload[s]
			}
		}
		ev2 := make(map[string]interface{})
		for k, v := range ev {
			ev2[k] = v
		}
		ev2["payload"] = mapped
		out = append(out, ev2)
	}
	return out
}

func writeToSink(ctx context.Context, pool *pgxpool.Pool, pipelineID uuid.UUID, sinks []Sink, events []map[string]interface{}) error {
	if len(sinks) == 0 {
		return writeToStreamSinkResults(ctx, pool, pipelineID, events)
	}
	sink := sinks[0]
	if sink.Type != databaseSink {
		return writeToStreamSinkResults(ctx, pool, pipelineID, events)
	}
	_ = sink.Target // reserved for future dynamic table
	return writeToStreamSinkResults(ctx, pool, pipelineID, events)
}

func writeToStreamSinkResults(ctx context.Context, pool *pgxpool.Pool, pipelineID uuid.UUID, events []map[string]interface{}) error {
	ensureStreamSinkResults(ctx, pool)
	for _, ev := range events {
		payloadBytes, _ := json.Marshal(ev["payload"])
		eventID, _ := ev["id"].(int64)
		_, err := pool.Exec(ctx,
			`INSERT INTO neuronip.stream_sink_results (id, pipeline_id, event_id, payload, created_at)
			 VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
			pipelineID, eventID, payloadBytes,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureStreamSinkResults(ctx context.Context, pool *pgxpool.Pool) {
	_, _ = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS neuronip.stream_sink_results (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			pipeline_id UUID NOT NULL,
			event_id BIGINT NOT NULL,
			payload JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
}
