package logging

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
)

const requestIDKey contextKey = "request_id"

/* Logger wraps slog.Logger with additional utilities */
type Logger struct {
	*slog.Logger
}

/* Config holds logging configuration */
type Config struct {
	Level  string
	Format string
	Output string
}

/* FromConfig converts config.LoggingConfig to logging.Config */
func FromConfig(level, format, output string) Config {
	return Config{
		Level:  level,
		Format: format,
		Output: output,
	}
}

/* NewLogger creates a new structured logger */
func NewLogger(cfg Config) *Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
		AddSource: cfg.Level == "debug",
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time as ISO8601
			if a.Key == slog.TimeKey {
				return slog.String("time", a.Value.Time().Format(time.RFC3339Nano))
			}
			return a
		},
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

/* WithRequestID adds request ID to the logger */
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("request_id", requestID),
	}
}

/* WithContext extracts request ID from context and adds it to logger */
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if requestID := GetRequestID(ctx); requestID != "" {
		return l.WithRequestID(requestID)
	}
	return l
}

/* WithError adds an error to the logger */
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
	}
}

/* WithFields adds multiple fields to the logger */
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

/* LogError logs an error with context */
func (l *Logger) LogError(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	logger := l.WithContext(ctx).WithError(err)
	
	// Add stack trace in debug mode
	if l.Logger.Enabled(ctx, slog.LevelDebug) {
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			funcName := runtime.FuncForPC(pc).Name()
			logger = logger.WithFields(map[string]interface{}{
				"file": file,
				"line": line,
				"func": funcName,
			})
		}
	}

	// Add additional fields
	if len(fields) > 0 {
		for _, f := range fields {
			logger = logger.WithFields(f)
		}
	}

	logger.Error(msg)
}

/* GetRequestID retrieves request ID from context */
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

/* SetRequestID sets request ID in context */
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

/* GenerateRequestID generates a new request ID */
func GenerateRequestID() string {
	return uuid.New().String()
}

/* DefaultLogger returns the default logger instance */
var DefaultLogger *Logger

/* InitLogger initializes the default logger */
func InitLogger(level, format, output string) {
	cfg := FromConfig(level, format, output)
	DefaultLogger = NewLogger(cfg)
}

/* Helper functions for default logger */

/* Debug logs a debug message */
func Debug(msg string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Debug(msg, args...)
	}
}

/* Info logs an info message */
func Info(msg string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Info(msg, args...)
	}
}

/* Warn logs a warning message */
func Warn(msg string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Warn(msg, args...)
	}
}

/* Error logs an error message */
func Error(msg string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Error(msg, args...)
	}
}

/* ErrorContext logs an error with context */
func ErrorContext(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.LogError(ctx, msg, err, fields...)
	}
}
