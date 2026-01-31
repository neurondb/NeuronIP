package cdc

import (
	"time"
)

/* ChangeEvent represents a single change event from CDC */
type ChangeEvent struct {
	Table      string                 `json:"table"`
	Operation  string                 `json:"operation"` // "insert", "update", "delete"
	LSN        string                 `json:"lsn"`
	Timestamp  time.Time              `json:"timestamp"`
	OldData    map[string]interface{} `json:"old_data,omitempty"`
	NewData    map[string]interface{} `json:"new_data,omitempty"`
}
