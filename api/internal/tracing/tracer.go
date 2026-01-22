package tracing

import (
	"context"
	"time"

	"github.com/google/uuid"
)

/* TracerService provides distributed tracing functionality */
type TracerService struct {
	enabled bool
}

/* NewTracerService creates a new tracer service */
func NewTracerService(enabled bool) *TracerService {
	return &TracerService{enabled: enabled}
}

/* Trace represents a distributed trace */
type Trace struct {
	TraceID    string                 `json:"trace_id"`
	SpanID     string                 `json:"span_id"`
	ParentSpanID *string              `json:"parent_span_id,omitempty"`
	Operation  string                 `json:"operation"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	Duration   *time.Duration         `json:"duration,omitempty"`
	Tags       map[string]interface{} `json:"tags,omitempty"`
	Logs       []TraceLog             `json:"logs,omitempty"`
}

/* TraceLog represents a log entry in a trace */
type TraceLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

/* StartSpan starts a new trace span */
func (s *TracerService) StartSpan(ctx context.Context, operation string) (context.Context, *Trace) {
	if !s.enabled {
		return ctx, nil
	}

	traceID := getTraceID(ctx)
	if traceID == "" {
		traceID = uuid.New().String()
		ctx = setTraceID(ctx, traceID)
	}

	spanID := uuid.New().String()
	parentSpanID := getSpanID(ctx)

	trace := &Trace{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		Operation:    operation,
		StartTime:    time.Now(),
		Tags:         make(map[string]interface{}),
		Logs:         []TraceLog{},
	}

	ctx = setSpanID(ctx, spanID)
	ctx = setTrace(ctx, trace)

	return ctx, trace
}

/* EndSpan ends a trace span */
func (s *TracerService) EndSpan(ctx context.Context, trace *Trace) {
	if !s.enabled || trace == nil {
		return
	}

	endTime := time.Now()
	trace.EndTime = &endTime
	duration := endTime.Sub(trace.StartTime)
	trace.Duration = &duration

	// Store trace (in production, send to tracing backend)
	_ = trace
}

/* AddTag adds a tag to a trace */
func (s *TracerService) AddTag(trace *Trace, key string, value interface{}) {
	if trace != nil && trace.Tags != nil {
		trace.Tags[key] = value
	}
}

/* AddLog adds a log entry to a trace */
func (s *TracerService) AddLog(trace *Trace, message string, fields map[string]interface{}) {
	if trace != nil {
		trace.Logs = append(trace.Logs, TraceLog{
			Timestamp: time.Now(),
			Message:   message,
			Fields:    fields,
		})
	}
}

/* Context keys for trace ID and span ID */
type contextKey string

const (
	traceIDKey contextKey = "trace_id"
	spanIDKey  contextKey = "span_id"
	traceKey   contextKey = "trace"
)

/* getTraceID gets trace ID from context */
func getTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

/* setTraceID sets trace ID in context */
func setTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

/* getSpanID gets span ID from context */
func getSpanID(ctx context.Context) *string {
	if id, ok := ctx.Value(spanIDKey).(string); ok {
		return &id
	}
	return nil
}

/* setSpanID sets span ID in context */
func setSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

/* getTrace gets trace from context */
func getTrace(ctx context.Context) *Trace {
	if trace, ok := ctx.Value(traceKey).(*Trace); ok {
		return trace
	}
	return nil
}

/* setTrace sets trace in context */
func setTrace(ctx context.Context, trace *Trace) context.Context {
	return context.WithValue(ctx, traceKey, trace)
}
