package handlers

import (
	"context"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/auth"
)

/* getUserIDFromContext extracts user ID from request context */
func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := auth.GetUserIDFromContext(ctx); ok {
		return userID
	}
	return ""
}

/* parseTime parses a time string in RFC3339 format, returns zero time if invalid */
func parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
