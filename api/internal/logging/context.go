package logging

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	userIDKey       contextKey = "user_id"
	operationKey    contextKey = "operation"
)

var (
	correlationIDGenerator = &sync.Mutex{}
	correlationCounter     int64
)

/* GetCorrelationID retrieves correlation ID from context */
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

/* SetCorrelationID sets correlation ID in context */
func SetCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

/* GenerateCorrelationID generates a new correlation ID */
func GenerateCorrelationID() string {
	correlationIDGenerator.Lock()
	defer correlationIDGenerator.Unlock()
	correlationCounter++
	return fmt.Sprintf("corr-%d-%d", time.Now().UnixNano(), correlationCounter)
}

/* GetUserID retrieves user ID from context */
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

/* SetUserID sets user ID in context */
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

/* GetOperation retrieves operation name from context */
func GetOperation(ctx context.Context) string {
	if op, ok := ctx.Value(operationKey).(string); ok {
		return op
	}
	return ""
}

/* SetOperation sets operation name in context */
func SetOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, operationKey, operation)
}

/* WithCorrelationID adds correlation ID to logger */
func (l *Logger) WithCorrelationID(correlationID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("correlation_id", correlationID),
	}
}

/* WithUserID adds user ID to logger */
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("user_id", userID),
	}
}

/* WithOperation adds operation to logger */
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		Logger: l.Logger.With("operation", operation),
	}
}

/* WithFullContext extracts all context values and adds them to logger */
func (l *Logger) WithFullContext(ctx context.Context) *Logger {
	logger := l.Logger

	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		logger = logger.With("correlation_id", correlationID)
	}

	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With("request_id", requestID)
	}

	if userID := GetUserID(ctx); userID != "" {
		logger = logger.With("user_id", userID)
	}

	if operation := GetOperation(ctx); operation != "" {
		logger = logger.With("operation", operation)
	}

	return &Logger{Logger: logger}
}
