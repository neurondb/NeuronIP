package tracing

import (
	"context"
	"fmt"
	"time"
)

/* Span represents a tracing span */
type Span struct {
	TraceID    string
	SpanID     string
	ParentID   string
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Tags       map[string]string
	Logs       []SpanLog
	Status     SpanStatus
	Error      error
}

/* SpanStatus represents span status */
type SpanStatus string

const (
	SpanStatusOK    SpanStatus = "ok"
	SpanStatusError SpanStatus = "error"
)

/* SpanLog represents a span log entry */
type SpanLog struct {
	Timestamp time.Time
	Fields    map[string]interface{}
}

/* StartSpan starts a new span */
func StartSpan(ctx context.Context, name string) (context.Context, *Span) {
	traceID := getTraceIDFromContext(ctx)
	if traceID == "" {
		traceID = generateTraceID()
		ctx = setTraceIDInContext(ctx, traceID)
	}

	spanID := generateSpanID()
	parentID := getSpanIDFromContext(ctx)

	span := &Span{
		TraceID:   traceID,
		SpanID:    spanID,
		ParentID:  parentID,
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Logs:      make([]SpanLog, 0),
		Status:    SpanStatusOK,
	}

	ctx = withSpanInContext(ctx, span)
	return ctx, span
}

/* EndSpan ends a span */
func EndSpan(span *Span) {
	if span == nil {
		return
	}
	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)
}

/* SetTag sets a tag on a span */
func SetTag(span *Span, key, value string) {
	if span == nil {
		return
	}
	span.Tags[key] = value
}

/* AddLog adds a log entry to a span */
func AddLog(span *Span, fields map[string]interface{}) {
	if span == nil {
		return
	}
	span.Logs = append(span.Logs, SpanLog{
		Timestamp: time.Now(),
		Fields:    fields,
	})
}

/* SetError sets an error on a span */
func SetError(span *Span, err error) {
	if span == nil {
		return
	}
	span.Error = err
	span.Status = SpanStatusError
	if err != nil {
		SetTag(span, "error", "true")
		SetTag(span, "error.message", err.Error())
	}
}

/* SpanFromContext retrieves a span from context */
func SpanFromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value(spanContextKey{}).(*Span); ok {
		return span
	}
	return nil
}

type spanContextKey struct{}

/* withSpanInContext adds a span to context */
func withSpanInContext(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanContextKey{}, span)
}

/* getTraceIDFromContext retrieves trace ID from context */
func getTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

/* setTraceIDInContext adds trace ID to context */
func setTraceIDInContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

/* generateTraceID generates a new trace ID */
func generateTraceID() string {
	return fmt.Sprintf("trace-%d", time.Now().UnixNano())
}

/* generateSpanID generates a new span ID */
func generateSpanID() string {
	return fmt.Sprintf("span-%d", time.Now().UnixNano())
}

/* getSpanIDFromContext retrieves span ID from context */
func getSpanIDFromContext(ctx context.Context) string {
	span := SpanFromContext(ctx)
	if span != nil {
		return span.SpanID
	}
	return ""
}
