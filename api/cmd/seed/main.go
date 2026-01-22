package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/config"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

var (
	seedType = flag.String("type", "demo", "Seed type: demo, test, minimal")
	clear    = flag.Bool("clear", false, "Clear existing data before seeding")
)

func main() {
	flag.Parse()

	cfg := config.Load()
	pool, err := db.NewPool(nil, cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	seeder := NewSeeder(pool.Pool)

	if *clear {
		fmt.Println("Clearing existing data...")
		if err := seeder.ClearData(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clear data: %v\n", err)
			os.Exit(1)
		}
	}

	switch *seedType {
	case "demo":
		fmt.Println("Seeding demo data...")
		if err := seeder.SeedDemo(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to seed demo data: %v\n", err)
			os.Exit(1)
		}
	case "test":
		fmt.Println("Seeding test data...")
		if err := seeder.SeedTest(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to seed test data: %v\n", err)
			os.Exit(1)
		}
	case "minimal":
		fmt.Println("Seeding minimal data...")
		if err := seeder.SeedMinimal(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to seed minimal data: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown seed type: %s\n", *seedType)
		os.Exit(1)
	}

	fmt.Println("Seeding completed successfully")
}

/* Seeder provides data seeding functionality */
type Seeder struct {
	pool *pgxpool.Pool
}

/* NewSeeder creates a new seeder */
func NewSeeder(pool *pgxpool.Pool) *Seeder {
	return &Seeder{pool: pool}
}

/* ClearData clears existing seed data */
func (s *Seeder) ClearData() error {
	// Clear seed data tables (be careful not to delete production data)
	tables := []string{
		"knowledge_documents",
		"support_tickets",
		"warehouse_schemas",
		"data_sources",
	}

	for _, table := range tables {
		query := fmt.Sprintf(`DELETE FROM neuronip.%s WHERE metadata->>'seeded' = 'true'`, table)
		_, err := s.pool.Exec(nil, query)
		if err != nil {
			// Log but continue
			fmt.Printf("Warning: Failed to clear %s: %v\n", table, err)
		}
	}

	return nil
}

/* SeedDemo seeds demo data */
func (s *Seeder) SeedDemo() error {
	fmt.Println("Creating demo users...")
	if err := s.seedUsers(); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	fmt.Println("Creating demo API keys...")
	if err := s.seedAPIKeys(); err != nil {
		return fmt.Errorf("failed to seed API keys: %w", err)
	}

	fmt.Println("Creating demo support tickets...")
	if err := s.seedSupportTickets(); err != nil {
		return fmt.Errorf("failed to seed support tickets: %w", err)
	}

	fmt.Println("Creating demo knowledge base...")
	if err := s.seedKnowledgeBase(); err != nil {
		return fmt.Errorf("failed to seed knowledge base: %w", err)
	}

	fmt.Println("Creating demo warehouse schemas...")
	if err := s.seedWarehouseSchemas(); err != nil {
		return fmt.Errorf("failed to seed warehouse schemas: %w", err)
	}

	fmt.Println("Creating demo saved searches...")
	if err := s.seedSavedSearches(); err != nil {
		return fmt.Errorf("failed to seed saved searches: %w", err)
	}

	fmt.Println("Creating demo workflows...")
	if err := s.seedWorkflows(); err != nil {
		return fmt.Errorf("failed to seed workflows: %w", err)
	}

	fmt.Println("Creating demo metrics...")
	if err := s.seedMetrics(); err != nil {
		return fmt.Errorf("failed to seed metrics: %w", err)
	}

	return nil
}

func (s *Seeder) seedUsers() error {
	// Default password for all seeded users: "demo123" (change in production!)
	// This password hash corresponds to "demo123" using bcrypt with cost 10
	defaultPasswordHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	
	users := []struct {
		email    string
		name     string
		password string
		role     string
	}{
		{"demo@example.com", "Demo User", defaultPasswordHash, "admin"},
		{"john@example.com", "John Doe", defaultPasswordHash, "user"},
		{"jane@example.com", "Jane Smith", defaultPasswordHash, "user"},
		{"admin@example.com", "Admin User", defaultPasswordHash, "admin"},
	}

	for _, u := range users {
		query := `
			INSERT INTO neuronip.users (email, password_hash, name, role, metadata)
			VALUES ($1, $2, $3, $4, '{"seeded": true}'::jsonb)
			ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, role = EXCLUDED.role
		`
		_, err := s.pool.Exec(nil, query, u.email, u.password, u.name, u.role)
		if err != nil {
			fmt.Printf("Warning: Failed to insert user %s: %v\n", u.email, err)
		}
	}
	return nil
}

func (s *Seeder) seedAPIKeys() error {
	query := `
		INSERT INTO neuronip.api_keys (name, key_hash, rate_limit_per_hour, metadata)
		VALUES 
			('Demo API Key', '$2a$10$demo', 1000, '{"seeded": true}'::jsonb),
			('Development Key', '$2a$10$dev', 500, '{"seeded": true}'::jsonb)
		ON CONFLICT DO NOTHING
	`
	_, err := s.pool.Exec(nil, query)
	if err != nil {
		fmt.Printf("Warning: Failed to insert API keys: %v\n", err)
	}
	return nil
}

func (s *Seeder) seedSupportTickets() error {
	// Load support tickets from JSON file
	supportTicketsFile := "../../../examples/demos/support-tickets-demo.json"
	data, err := os.ReadFile(supportTicketsFile)
	if err != nil {
		fmt.Printf("Warning: Could not read support tickets file: %v\n", err)
		return nil
	}

	var ticketsData map[string]interface{}
	if err := json.Unmarshal(data, &ticketsData); err != nil {
		return err
	}

	tickets, ok := ticketsData["tickets"].([]interface{})
	if !ok {
		return nil
	}

	for _, ticketRaw := range tickets {
		ticket, ok := ticketRaw.(map[string]interface{})
		if !ok {
			continue
		}

		query := `
			INSERT INTO neuronip.support_tickets (
				id, customer_id, subject, description, priority, status, metadata
			) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
			ON CONFLICT (id) DO NOTHING
		`
		id := fmt.Sprintf("%v", ticket["id"])
		customerID := fmt.Sprintf("%v", ticket["customer_id"])
		subject := fmt.Sprintf("%v", ticket["subject"])
		description := fmt.Sprintf("%v", ticket["description"])
		priority := fmt.Sprintf("%v", ticket["priority"])
		status := "open"
		if s, ok := ticket["status"].(string); ok {
			status = s
		}

		metadata := map[string]interface{}{
			"seeded": true,
			"customer_email": ticket["customer_email"],
			"customer_name":  ticket["customer_name"],
		}
		metadataJSON, _ := json.Marshal(metadata)

		_, err := s.pool.Exec(nil, query, id, customerID, subject, description, priority, status, string(metadataJSON))
		if err != nil {
			fmt.Printf("Warning: Failed to insert ticket %s: %v\n", id, err)
		}
	}
	return nil
}

func (s *Seeder) seedKnowledgeBase() error {
	knowledgeBaseFile := "../../../examples/demos/knowledge-base-demo.json"
	data, err := os.ReadFile(knowledgeBaseFile)
	if err != nil {
		fmt.Printf("Warning: Could not read knowledge base file: %v\n", err)
		return nil
	}

	var kbData map[string]interface{}
	if err := json.Unmarshal(data, &kbData); err != nil {
		return err
	}

	collections, ok := kbData["collections"].([]interface{})
	if !ok {
		return nil
	}

	for _, collectionRaw := range collections {
		collection, ok := collectionRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Create collection
		collectionID := fmt.Sprintf("%v", collection["id"])
		collectionName := fmt.Sprintf("%v", collection["name"])

		collectionQuery := `
			INSERT INTO neuronip.semantic_collections (id, name, metadata)
			VALUES ($1, $2, '{"seeded": true}'::jsonb)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := s.pool.Exec(nil, collectionQuery, collectionID, collectionName)
		if err != nil {
			fmt.Printf("Warning: Failed to insert collection %s: %v\n", collectionID, err)
			continue
		}

		// Insert documents
		documents, ok := collection["documents"].([]interface{})
		if !ok {
			continue
		}

		for _, docRaw := range documents {
			doc, ok := docRaw.(map[string]interface{})
			if !ok {
				continue
			}

			docID := fmt.Sprintf("%v", doc["id"])
			title := fmt.Sprintf("%v", doc["title"])
			content := fmt.Sprintf("%v", doc["content"])

			docQuery := `
				INSERT INTO neuronip.knowledge_documents (
					id, collection_id, title, content, content_type, metadata
				) VALUES ($1, $2, $3, $4, $5, $6::jsonb)
				ON CONFLICT (id) DO NOTHING
			`
			contentType := "documentation"
			if ct, ok := doc["content_type"].(string); ok {
				contentType = ct
			}

			metadata := map[string]interface{}{"seeded": true}
			if tags, ok := doc["tags"].([]interface{}); ok {
				metadata["tags"] = tags
			}
			metadataJSON, _ := json.Marshal(metadata)

			_, err := s.pool.Exec(nil, docQuery, docID, collectionID, title, content, contentType, string(metadataJSON))
			if err != nil {
				fmt.Printf("Warning: Failed to insert document %s: %v\n", docID, err)
			}
		}
	}
	return nil
}

func (s *Seeder) seedWarehouseSchemas() error {
	warehouseSalesFile := "../../../examples/demos/warehouse-sales-demo.json"
	data, err := os.ReadFile(warehouseSalesFile)
	if err != nil {
		fmt.Printf("Warning: Could not read warehouse file: %v\n", err)
		return nil
	}

	var warehouseData map[string]interface{}
	if err := json.Unmarshal(data, &warehouseData); err != nil {
		return err
	}

	schemaData, ok := warehouseData["schema"].(map[string]interface{})
	if !ok {
		return nil
	}

	schemaID := fmt.Sprintf("%v", schemaData["id"])
	schemaName := fmt.Sprintf("%v", schemaData["name"])
	schemaDesc := fmt.Sprintf("%v", schemaData["description"])

	schemaJSON, _ := json.Marshal(warehouseData["schema"])

	query := `
		INSERT INTO neuronip.warehouse_schemas (id, name, description, schema_definition, metadata)
		VALUES ($1, $2, $3, $4::jsonb, '{"seeded": true}'::jsonb)
		ON CONFLICT (id) DO NOTHING
	`
	_, err = s.pool.Exec(nil, query, schemaID, schemaName, schemaDesc, string(schemaJSON))
	if err != nil {
		fmt.Printf("Warning: Failed to insert warehouse schema: %v\n", err)
	}

	return nil
}

func (s *Seeder) seedSavedSearches() error {
	searches := []struct {
		id          string
		name        string
		description string
		query       string
		isPublic    bool
	}{
		{"search-001", "Top Products by Revenue", "Find products with highest revenue", "SELECT name, SUM(total_amount) as revenue FROM orders JOIN order_items ON orders.order_id = order_items.order_id GROUP BY name ORDER BY revenue DESC LIMIT 10", true},
		{"search-002", "Active Customers", "List all active customers", "SELECT * FROM customers WHERE registration_date > NOW() - INTERVAL '30 days'", false},
		{"search-003", "Monthly Sales Summary", "Monthly sales breakdown", "SELECT DATE_TRUNC('month', order_date) as month, SUM(total_amount) as total FROM orders GROUP BY month ORDER BY month DESC", true},
	}

	for _, search := range searches {
		query := `
			INSERT INTO neuronip.saved_searches (
				id, name, description, query, is_public, metadata
			) VALUES ($1, $2, $3, $4, $5, '{"seeded": true}'::jsonb)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := s.pool.Exec(nil, query, search.id, search.name, search.description, search.query, search.isPublic)
		if err != nil {
			fmt.Printf("Warning: Failed to insert saved search %s: %v\n", search.id, err)
		}
	}
	return nil
}

func (s *Seeder) seedWorkflows() error {
	query := `
		INSERT INTO neuronip.workflows (id, name, description, definition, metadata)
		VALUES (
			'workflow-001',
			'Data Processing Pipeline',
			'Processes incoming data and generates reports',
			'{"nodes": [{"type": "agent", "name": "Process Data"}], "edges": []}'::jsonb,
			'{"seeded": true}'::jsonb
		)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := s.pool.Exec(nil, query)
	if err != nil {
		fmt.Printf("Warning: Failed to insert workflow: %v\n", err)
	}
	return nil
}

func (s *Seeder) seedMetrics() error {
	metrics := []struct {
		id          string
		name        string
		description string
		definition  string
	}{
		{"metric-001", "Total Revenue", "Sum of all order amounts", `{"type": "sum", "field": "total_amount", "table": "orders"}`},
		{"metric-002", "Customer Count", "Number of unique customers", `{"type": "count_distinct", "field": "customer_id", "table": "customers"}`},
		{"metric-003", "Average Order Value", "Average amount per order", `{"type": "avg", "field": "total_amount", "table": "orders"}`},
	}

	for _, metric := range metrics {
		query := `
			INSERT INTO neuronip.metrics (id, name, description, definition, metadata)
			VALUES ($1, $2, $3, $4::jsonb, '{"seeded": true}'::jsonb)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := s.pool.Exec(nil, query, metric.id, metric.name, metric.description, metric.definition)
		if err != nil {
			fmt.Printf("Warning: Failed to insert metric %s: %v\n", metric.id, err)
		}
	}
	return nil
}

/* SeedTest seeds test data */
func (s *Seeder) SeedTest() error {
	// Seed minimal test data
	return s.SeedMinimal()
}

/* SeedMinimal seeds minimal data */
func (s *Seeder) SeedMinimal() error {
	// Create a test user
	query := `
		INSERT INTO neuronip.users (email, password_hash, name, role)
		VALUES ('demo@example.com', '$2a$10$dummy', 'Demo User', 'admin')
		ON CONFLICT (email) DO NOTHING
	`
	_, err := s.pool.Exec(nil, query)
	return err
}
