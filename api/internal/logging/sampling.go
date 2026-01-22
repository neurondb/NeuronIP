package logging

import (
	"context"
	"sync/atomic"
	"time"
)

/* Sampler determines whether to log a message based on sampling rate */
type Sampler struct {
	sampleRate float64 // 0.0 to 1.0, where 1.0 means log everything
	counter    int64
}

/* NewSampler creates a new sampler with the given sample rate */
func NewSampler(sampleRate float64) *Sampler {
	if sampleRate < 0.0 {
		sampleRate = 0.0
	}
	if sampleRate > 1.0 {
		sampleRate = 1.0
	}
	return &Sampler{
		sampleRate: sampleRate,
	}
}

/* ShouldSample determines if a message should be logged */
func (s *Sampler) ShouldSample() bool {
	if s.sampleRate >= 1.0 {
		return true
	}
	if s.sampleRate <= 0.0 {
		return false
	}

	// Use atomic counter for thread-safe sampling
	counter := atomic.AddInt64(&s.counter, 1)
	// Simple modulo-based sampling
	return float64(counter%100) < s.sampleRate*100
}

/* LogLevelEscalation handles log level escalation for errors */
type LogLevelEscalation struct {
	errorThreshold int
	errorCount     int64
	windowStart    time.Time
	windowDuration time.Duration
}

/* NewLogLevelEscalation creates a new log level escalation */
func NewLogLevelEscalation(threshold int, windowDuration time.Duration) *LogLevelEscalation {
	return &LogLevelEscalation{
		errorThreshold: threshold,
		windowDuration: windowDuration,
		windowStart:    time.Now(),
	}
}

/* RecordError records an error and determines if escalation is needed */
func (lle *LogLevelEscalation) RecordError() bool {
	now := time.Now()
	
	// Reset window if needed
	if now.Sub(lle.windowStart) >= lle.windowDuration {
		lle.errorCount = 0
		lle.windowStart = now
	}

	lle.errorCount++
	return lle.errorCount >= int64(lle.errorThreshold)
}

/* ShouldEscalate determines if log level should be escalated */
func (lle *LogLevelEscalation) ShouldEscalate() bool {
	return lle.errorCount >= int64(lle.errorThreshold)
}

/* StructuredLogSchema defines a structured log schema */
type StructuredLogSchema struct {
	Timestamp     time.Time              `json:"timestamp"`
	Level         string                 `json:"level"`
	Message       string                 `json:"message"`
	RequestID     string                 `json:"request_id,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	Operation     string                 `json:"operation,omitempty"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Stack         []string               `json:"stack,omitempty"`
}

/* ToStructuredLog converts context and message to structured log */
func ToStructuredLog(ctx context.Context, level, message string, err error, fields map[string]interface{}) *StructuredLogSchema {
	schema := &StructuredLogSchema{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    make(map[string]interface{}),
	}

	if ctx != nil {
		schema.RequestID = GetRequestID(ctx)
		schema.CorrelationID = GetCorrelationID(ctx)
		schema.UserID = GetUserID(ctx)
		schema.Operation = GetOperation(ctx)
	}

	if err != nil {
		schema.Error = err.Error()
	}

	if fields != nil {
		for k, v := range fields {
			schema.Fields[k] = v
		}
	}

	return schema
}
