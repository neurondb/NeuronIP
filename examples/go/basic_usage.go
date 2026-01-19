package main

import (
	"fmt"
	"log"

	"github.com/neurondb/NeuronIP/sdk/go"
)

func main() {
	// Initialize client
	client := neuronip.NewClient(
		"http://localhost:8082/api/v1",
		"your-api-key-here",
	)

	// Health check
	health, err := client.HealthCheck()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("API Status: %v\n", health["status"])

	// Semantic search
	results, err := client.SemanticSearch("machine learning algorithms", 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Search completed\n")

	// Warehouse query
	queryResult, err := client.WarehouseQuery("Show me total sales by region", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query executed: %v\n", queryResult["query_id"])

	// Create ingestion job
	job, err := client.CreateIngestionJob(
		"your-data-source-id",
		"sync",
		map[string]interface{}{
			"mode":   "full",
			"tables": []string{"users", "orders"},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created job: %v\n", job["id"])
}
