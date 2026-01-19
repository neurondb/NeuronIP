package parsers

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

/* CSVParser parses CSV files */
type CSVParser struct {
	hasHeader bool
	delimiter rune
}

/* NewCSVParser creates a new CSV parser */
func NewCSVParser(hasHeader bool, delimiter rune) *CSVParser {
	if delimiter == 0 {
		delimiter = ','
	}
	return &CSVParser{
		hasHeader: hasHeader,
		delimiter: delimiter,
	}
}

/* Parse parses a CSV file and returns rows and inferred schema */
func (p *CSVParser) Parse(reader io.Reader) ([]map[string]interface{}, *CSVSchema, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = p.delimiter
	
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) == 0 {
		return nil, nil, fmt.Errorf("CSV file is empty")
	}
	
	var headers []string
	var dataStart int
	
	if p.hasHeader {
		headers = records[0]
		dataStart = 1
	} else {
		// Generate column names
		headers = make([]string, len(records[0]))
		for i := range headers {
			headers[i] = fmt.Sprintf("column_%d", i+1)
		}
		dataStart = 0
	}
	
	// Infer schema from data
	schema := p.inferSchema(records[dataStart:], headers)
	
	// Parse rows
	rows := make([]map[string]interface{}, 0, len(records)-dataStart)
	for i := dataStart; i < len(records); i++ {
		row := make(map[string]interface{})
		for j, header := range headers {
			if j < len(records[i]) {
				value := strings.TrimSpace(records[i][j])
				row[header] = p.parseValue(value, schema.Columns[j].DataType)
			} else {
				row[header] = nil
			}
		}
		rows = append(rows, row)
	}
	
	return rows, schema, nil
}

/* inferSchema infers the schema from CSV data */
func (p *CSVParser) inferSchema(records [][]string, headers []string) *CSVSchema {
	if len(records) == 0 {
		return &CSVSchema{Columns: []ColumnInfo{}}
	}
	
	columns := make([]ColumnInfo, len(headers))
	
	// Sample first 100 rows for type inference
	sampleSize := 100
	if len(records) < sampleSize {
		sampleSize = len(records)
	}
	
	for i, header := range headers {
		column := ColumnInfo{
			Name: header,
			DataType: "text", // Default
		}
		
		// Try to infer type from sample
		nonEmptyCount := 0
		isInt := true
		isFloat := true
		isBool := true
		
		for j := 0; j < sampleSize; j++ {
			if i >= len(records[j]) {
				continue
			}
			
			value := strings.TrimSpace(records[j][i])
			if value == "" {
				continue
			}
			
			nonEmptyCount++
			
			// Check if integer
			if isInt {
				if _, err := strconv.ParseInt(value, 10, 64); err != nil {
					isInt = false
				}
			}
			
			// Check if float
			if isFloat {
				if _, err := strconv.ParseFloat(value, 64); err != nil {
					isFloat = false
				}
			}
			
			// Check if boolean
			if isBool {
				lower := strings.ToLower(value)
				if lower != "true" && lower != "false" && lower != "1" && lower != "0" && 
				   lower != "yes" && lower != "no" {
					isBool = false
				}
			}
		}
		
		// Determine type
		if nonEmptyCount == 0 {
			column.DataType = "text"
		} else if isBool && nonEmptyCount > 0 {
			column.DataType = "boolean"
		} else if isInt {
			column.DataType = "integer"
		} else if isFloat {
			column.DataType = "numeric"
		} else {
			column.DataType = "text"
		}
		
		columns[i] = column
	}
	
	return &CSVSchema{Columns: columns}
}

/* parseValue parses a string value according to the data type */
func (p *CSVParser) parseValue(value string, dataType string) interface{} {
	if value == "" {
		return nil
	}
	
	switch dataType {
	case "integer":
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
		return value
	case "numeric":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
		return value
	case "boolean":
		lower := strings.ToLower(value)
		return lower == "true" || lower == "1" || lower == "yes"
	default:
		return value
	}
}

/* CSVSchema represents the inferred schema of a CSV file */
type CSVSchema struct {
	Columns []ColumnInfo `json:"columns"`
}

/* ColumnInfo represents column information */
type ColumnInfo struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}
