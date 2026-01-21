package etl

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

/* FilterTransformation filters rows based on conditions */
type FilterTransformation struct{}

/* Transform filters data based on conditions */
func (f *FilterTransformation) Transform(ctx context.Context, data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error) {
	condition, ok := config["condition"].(string)
	if !ok {
		return data, nil
	}
	
	result := make([]map[string]interface{}, 0)
	for _, row := range data {
		if f.evaluateCondition(row, condition) {
			result = append(result, row)
		}
	}
	return result, nil
}

func (f *FilterTransformation) evaluateCondition(row map[string]interface{}, condition string) bool {
	parts := strings.Fields(condition)
	if len(parts) < 3 {
		return false
	}
	column := parts[0]
	operator := parts[1]
	valueStr := strings.Trim(strings.Join(parts[2:], " "), `"'`)
	rowValue, exists := row[column]
	if !exists {
		return false
	}
	return f.compareValues(rowValue, operator, valueStr)
}

func (f *FilterTransformation) compareValues(left interface{}, operator string, rightStr string) bool {
	right := f.convertValue(rightStr, left)
	switch operator {
	case "==", "=":
		return reflect.DeepEqual(left, right)
	case "!=", "<>":
		return !reflect.DeepEqual(left, right)
	case ">":
		return f.greaterThan(left, right)
	case ">=":
		return f.greaterThan(left, right) || reflect.DeepEqual(left, right)
	case "<":
		return f.lessThan(left, right)
	case "<=":
		return f.lessThan(left, right) || reflect.DeepEqual(left, right)
	}
	return false
}

func (f *FilterTransformation) convertValue(str string, target interface{}) interface{} {
	switch target.(type) {
	case int, int64:
		if val, err := strconv.ParseInt(str, 10, 64); err == nil {
			return val
		}
	case float64, float32:
		if val, err := strconv.ParseFloat(str, 64); err == nil {
			return val
		}
	case bool:
		if val, err := strconv.ParseBool(str); err == nil {
			return val
		}
	}
	return str
}

func (f *FilterTransformation) greaterThan(left, right interface{}) bool {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l > r
		}
	case int64:
		if r, ok := right.(int64); ok {
			return l > r
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l > r
		}
	}
	return false
}

func (f *FilterTransformation) lessThan(left, right interface{}) bool {
	switch l := left.(type) {
	case int:
		if r, ok := right.(int); ok {
			return l < r
		}
	case int64:
		if r, ok := right.(int64); ok {
			return l < r
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l < r
		}
	}
	return false
}

/* MapTransformation maps/transforms column values */
type MapTransformation struct{}

func (m *MapTransformation) Transform(ctx context.Context, data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error) {
	mappings, ok := config["mappings"].(map[string]interface{})
	if !ok {
		return data, nil
	}
	result := make([]map[string]interface{}, 0, len(data))
	for _, row := range data {
		newRow := make(map[string]interface{})
		for k, v := range row {
			newRow[k] = v
		}
		for newCol, mapping := range mappings {
			if mappingStr, ok := mapping.(string); ok {
				newRow[newCol] = m.evaluateExpression(row, mappingStr)
			}
		}
		result = append(result, newRow)
	}
	return result, nil
}

func (m *MapTransformation) evaluateExpression(row map[string]interface{}, expr string) interface{} {
	result := expr
	for key, value := range row {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

/* AggregateTransformation aggregates data */
type AggregateTransformation struct{}

func (a *AggregateTransformation) Transform(ctx context.Context, data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error) {
	groupBy, _ := config["group_by"].([]interface{})
	aggregations, _ := config["aggregations"].(map[string]interface{})
	if len(groupBy) == 0 {
		return a.aggregateAll(data, aggregations), nil
	}
	groups := make(map[string][]map[string]interface{})
	for _, row := range data {
		key := a.buildGroupKey(row, groupBy)
		groups[key] = append(groups[key], row)
	}
	result := make([]map[string]interface{}, 0, len(groups))
	for _, groupRows := range groups {
		aggRow := make(map[string]interface{})
		for _, col := range groupBy {
			if colStr, ok := col.(string); ok {
				if val, exists := groupRows[0][colStr]; exists {
					aggRow[colStr] = val
				}
			}
		}
		for col, funcName := range aggregations {
			if funcStr, ok := funcName.(string); ok {
				aggRow[col] = a.applyAggregation(groupRows, col, funcStr)
			}
		}
		result = append(result, aggRow)
	}
	return result, nil
}

func (a *AggregateTransformation) buildGroupKey(row map[string]interface{}, groupBy []interface{}) string {
	keys := make([]string, 0, len(groupBy))
	for _, col := range groupBy {
		if colStr, ok := col.(string); ok {
			if val, exists := row[colStr]; exists {
				keys = append(keys, fmt.Sprintf("%v", val))
			}
		}
	}
	return strings.Join(keys, "|")
}

func (a *AggregateTransformation) aggregateAll(data []map[string]interface{}, aggregations map[string]interface{}) []map[string]interface{} {
	result := make(map[string]interface{})
	for col, funcName := range aggregations {
		if funcStr, ok := funcName.(string); ok {
			result[col] = a.applyAggregation(data, col, funcStr)
		}
	}
	return []map[string]interface{}{result}
}

func (a *AggregateTransformation) applyAggregation(rows []map[string]interface{}, column string, funcName string) interface{} {
	values := make([]float64, 0)
	for _, row := range rows {
		if val, exists := row[column]; exists {
			if num, ok := a.toFloat64(val); ok {
				values = append(values, num)
			}
		}
	}
	if len(values) == 0 {
		return nil
	}
	switch funcName {
	case "sum":
		var sum float64
		for _, v := range values {
			sum += v
		}
		return sum
	case "avg", "mean":
		var sum float64
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values))
	case "min":
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min
	case "max":
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max
	case "count":
		return len(values)
	}
	return nil
}

func (a *AggregateTransformation) toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	}
	return 0, false
}

/* JoinTransformation joins two datasets */
type JoinTransformation struct{}

func (j *JoinTransformation) Transform(ctx context.Context, data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error) {
	// Get join configuration
	joinType, ok := config["join_type"].(string)
	if !ok {
		joinType = "inner" // Default to inner join
	}

	joinKey, ok := config["join_key"].(string)
	if !ok {
		return nil, fmt.Errorf("join_key is required for join transformation")
	}

	// Get second dataset from config
	secondDatasetRaw, ok := config["second_dataset"]
	if !ok {
		return nil, fmt.Errorf("second_dataset is required for join transformation")
	}

	// Convert second dataset to proper format
	secondDataset, ok := secondDatasetRaw.([]map[string]interface{})
	if !ok {
		// Try to convert from []interface{}
		if secondDatasetSlice, ok := secondDatasetRaw.([]interface{}); ok {
			secondDataset = make([]map[string]interface{}, 0, len(secondDatasetSlice))
			for _, item := range secondDatasetSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					secondDataset = append(secondDataset, itemMap)
				}
			}
		} else {
			return nil, fmt.Errorf("second_dataset must be an array of objects")
		}
	}

	// Create lookup map for second dataset
	lookupMap := make(map[interface{}][]map[string]interface{})
	for _, row := range secondDataset {
		keyValue := row[joinKey]
		if keyValue != nil {
			lookupMap[keyValue] = append(lookupMap[keyValue], row)
		}
	}

	// Perform join
	var result []map[string]interface{}

	switch strings.ToLower(joinType) {
	case "inner":
		// Inner join: only rows that match in both datasets
		for _, row := range data {
			keyValue := row[joinKey]
			if keyValue != nil {
				if matchingRows, found := lookupMap[keyValue]; found {
					for _, match := range matchingRows {
						joinedRow := make(map[string]interface{})
						// Copy all fields from first dataset
						for k, v := range row {
							joinedRow[k] = v
						}
						// Add fields from second dataset with prefix to avoid collisions
						prefix, _ := config["prefix"].(string)
						if prefix == "" {
							prefix = "right_"
						}
						for k, v := range match {
							if k == joinKey {
								continue // Skip join key as it's already in the row
							}
							joinedRow[prefix+k] = v
						}
						result = append(result, joinedRow)
					}
				}
			}
		}

	case "left":
		// Left join: all rows from first dataset, matched rows from second
		for _, row := range data {
			keyValue := row[joinKey]
			joinedRow := make(map[string]interface{})
			// Copy all fields from first dataset
			for k, v := range row {
				joinedRow[k] = v
			}

			if keyValue != nil {
				if matchingRows, found := lookupMap[keyValue]; found && len(matchingRows) > 0 {
					// Use first matching row
					match := matchingRows[0]
					prefix, _ := config["prefix"].(string)
					if prefix == "" {
						prefix = "right_"
					}
					for k, v := range match {
						if k == joinKey {
							continue
						}
						joinedRow[prefix+k] = v
					}
				}
			}
			result = append(result, joinedRow)
		}

	case "right":
		// Right join: all rows from second dataset, matched rows from first
		firstLookupMap := make(map[interface{}][]map[string]interface{})
		for _, row := range data {
			keyValue := row[joinKey]
			if keyValue != nil {
				firstLookupMap[keyValue] = append(firstLookupMap[keyValue], row)
			}
		}

		for _, row := range secondDataset {
			keyValue := row[joinKey]
			joinedRow := make(map[string]interface{})
			// Copy all fields from second dataset
			for k, v := range row {
				joinedRow[k] = v
			}

			if keyValue != nil {
				if matchingRows, found := firstLookupMap[keyValue]; found && len(matchingRows) > 0 {
					match := matchingRows[0]
					prefix, _ := config["prefix"].(string)
					if prefix == "" {
						prefix = "left_"
					}
					for k, v := range match {
						if k == joinKey {
							continue
						}
						joinedRow[prefix+k] = v
					}
				}
			}
			result = append(result, joinedRow)
		}

	case "outer", "full":
		// Full outer join: all rows from both datasets
		// First add all left rows with matches
		matchedKeys := make(map[interface{}]bool)
		for _, row := range data {
			keyValue := row[joinKey]
			joinedRow := make(map[string]interface{})
			for k, v := range row {
				joinedRow[k] = v
			}

			if keyValue != nil {
				matchedKeys[keyValue] = true
				if matchingRows, found := lookupMap[keyValue]; found && len(matchingRows) > 0 {
					match := matchingRows[0]
					prefix, _ := config["prefix"].(string)
					if prefix == "" {
						prefix = "right_"
					}
					for k, v := range match {
						if k != joinKey {
							joinedRow[prefix+k] = v
						}
					}
				}
			}
			result = append(result, joinedRow)
		}

		// Add unmatched rows from right
		prefix, _ := config["prefix"].(string)
		if prefix == "" {
			prefix = "left_"
		}
		for _, row := range secondDataset {
			keyValue := row[joinKey]
			if keyValue != nil && !matchedKeys[keyValue] {
				joinedRow := make(map[string]interface{})
				for k, v := range row {
					joinedRow[k] = v
				}
				result = append(result, joinedRow)
			}
		}

	default:
		return nil, fmt.Errorf("unsupported join type: %s (supported: inner, left, right, outer)", joinType)
	}

	return result, nil
}
