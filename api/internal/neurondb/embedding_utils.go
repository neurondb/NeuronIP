package neurondb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

/* ParseEmbeddingString parses an embedding string (PostgreSQL vector format) to float64 slice */
func ParseEmbeddingString(embeddingStr string) ([]float64, error) {
	if embeddingStr == "" {
		return nil, fmt.Errorf("empty embedding string")
	}

	// Remove brackets if present
	embeddingStr = strings.TrimSpace(embeddingStr)
	embeddingStr = strings.TrimPrefix(embeddingStr, "[")
	embeddingStr = strings.TrimSuffix(embeddingStr, "]")

	// Split by comma
	parts := strings.Split(embeddingStr, ",")
	embedding := make([]float64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		val, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse embedding value '%s': %w", part, err)
		}
		embedding = append(embedding, val)
	}

	return embedding, nil
}

/* FormatEmbeddingToString formats a float64 slice to PostgreSQL vector string format */
func FormatEmbeddingToString(embedding []float64) string {
	if len(embedding) == 0 {
		return "[]"
	}

	parts := make([]string, len(embedding))
	for i, val := range embedding {
		parts[i] = strconv.FormatFloat(val, 'f', -1, 64)
	}

	return "[" + strings.Join(parts, ",") + "]"
}

/* ParseEmbeddingJSON parses an embedding from JSON format */
func ParseEmbeddingJSON(embeddingJSON []byte) ([]float64, error) {
	var embedding []float64
	err := json.Unmarshal(embeddingJSON, &embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding JSON: %w", err)
	}
	return embedding, nil
}
