package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

/* OpenTelemetryTracer provides OpenTelemetry-based distributed tracing */
type OpenTelemetryTracer struct {
	tracer trace.Tracer
	enabled bool
}

/* NewOpenTelemetryTracer creates a new OpenTelemetry tracer */
func NewOpenTelemetryTracer(serviceName string, jaegerEndpoint string, enabled bool) (*OpenTelemetryTracer, error) {
	if !enabled {
		return &OpenTelemetryTracer{enabled: false}, nil
	}

	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
	)

	// Register as global tracer provider
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer(serviceName)

	return &OpenTelemetryTracer{
		tracer:  tracer,
		enabled: true,
	}, nil
}

/* StartSpan starts a new OpenTelemetry span */
func (ot *OpenTelemetryTracer) StartSpan(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !ot.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := ot.tracer.Start(ctx, operationName, opts...)
	return ctx, span
}

/* EndSpan ends a span */
func (ot *OpenTelemetryTracer) EndSpan(span trace.Span) {
	if !ot.enabled || !span.IsRecording() {
		return
	}
	span.End()
}

/* AddSpanAttributes adds attributes to a span */
func (ot *OpenTelemetryTracer) AddSpanAttributes(span trace.Span, attrs map[string]interface{}) {
	if !ot.enabled || !span.IsRecording() {
		return
	}

	for k, v := range attrs {
		span.SetAttributes(attribute.String(k, fmt.Sprintf("%v", v)))
	}
}

/* AddSpanEvent adds an event to a span */
func (ot *OpenTelemetryTracer) AddSpanEvent(span trace.Span, name string, attrs map[string]interface{}) {
	if !ot.enabled || !span.IsRecording() {
		return
	}

	eventAttrs := make([]attribute.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		eventAttrs = append(eventAttrs, attribute.String(k, fmt.Sprintf("%v", v)))
	}

	span.AddEvent(name, trace.WithAttributes(eventAttrs...))
}

/* RecordError records an error in a span */
func (ot *OpenTelemetryTracer) RecordError(span trace.Span, err error) {
	if !ot.enabled || !span.IsRecording() {
		return
	}
	span.RecordError(err)
}

/* SetSpanStatus sets the status of a span */
func (ot *OpenTelemetryTracer) SetSpanStatus(span trace.Span, code codes.Code, description string) {
	if !ot.enabled || !span.IsRecording() {
		return
	}
	span.SetStatus(code, description)
}

/* ExtractTraceContext extracts trace context from context */
func (ot *OpenTelemetryTracer) ExtractTraceContext(ctx context.Context) map[string]string {
	if !ot.enabled {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return nil
	}

	spanContext := span.SpanContext()
	return map[string]string{
		"trace_id": spanContext.TraceID().String(),
		"span_id":  spanContext.SpanID().String(),
	}
}

/* InjectTraceContext injects trace context into a map for propagation */
func (ot *OpenTelemetryTracer) InjectTraceContext(ctx context.Context, headers map[string]string) {
	if !ot.enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	spanContext := span.SpanContext()
	headers["trace_id"] = spanContext.TraceID().String()
	headers["span_id"] = spanContext.SpanID().String()
}

/* TraceWithSpan traces a function with a span */
func (ot *OpenTelemetryTracer) TraceWithSpan(ctx context.Context, operationName string, fn func(context.Context) error) error {
	ctx, span := ot.StartSpan(ctx, operationName)
	defer ot.EndSpan(span)

	err := fn(ctx)
	if err != nil {
		ot.RecordError(span, err)
		ot.SetSpanStatus(span, codes.Error, err.Error())
		return err
	}

	ot.SetSpanStatus(span, codes.Ok, "")
	return nil
}

/* TraceWithSpanAndDuration traces a function and records duration */
func (ot *OpenTelemetryTracer) TraceWithSpanAndDuration(ctx context.Context, operationName string, fn func(context.Context) error) error {
	start := time.Now()
	ctx, span := ot.StartSpan(ctx, operationName)
	defer func() {
		duration := time.Since(start)
		ot.AddSpanAttributes(span, map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
		})
		ot.EndSpan(span)
	}()

	err := fn(ctx)
	if err != nil {
		ot.RecordError(span, err)
		ot.SetSpanStatus(span, codes.Error, err.Error())
		return err
	}

	ot.SetSpanStatus(span, codes.Ok, "")
	return nil
}
