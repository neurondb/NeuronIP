package lineage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* DiscoveryService provides automatic lineage discovery functionality */
type DiscoveryService struct {
	pool *pgxpool.Pool
}

/* NewDiscoveryService creates a new discovery service */
func NewDiscoveryService(pool *pgxpool.Pool) *DiscoveryService {
	return &DiscoveryService{pool: pool}
}

/* DiscoveryRule represents a rule for automatic lineage discovery */
type DiscoveryRule struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	SourceType   string                 `json:"source_type"`   // "query_log", "sql_parser", "api_call", "etl_job"
	Pattern      map[string]interface{} `json:"pattern"`       // Pattern matching rules
	Enabled      bool                   `json:"enabled"`
	LastRunAt    *time.Time             `json:"last_run_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

/* DiscoveredLineage represents automatically discovered lineage */
type DiscoveredLineage struct {
	ID           uuid.UUID              `json:"id"`
	RuleID       uuid.UUID              `json:"rule_id"`
	SourceNodeID uuid.UUID              `json:"source_node_id"`
	TargetNodeID uuid.UUID              `json:"target_node_id"`
	EdgeType     string                 `json:"edge_type"`
	Confidence   float64                `json:"confidence"` // 0.0 to 1.0
	Evidence     map[string]interface{} `json:"evidence,omitempty"`
	Verified     bool                   `json:"verified"`
	CreatedAt    time.Time              `json:"created_at"`
}

/* CreateDiscoveryRule creates a new discovery rule */
func (s *DiscoveryService) CreateDiscoveryRule(ctx context.Context, rule DiscoveryRule) (*DiscoveryRule, error) {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	patternJSON, _ := json.Marshal(rule.Pattern)
	metadataJSON, _ := json.Marshal(rule.Metadata)

	query := `
		INSERT INTO neuronip.lineage_discovery_rules
		(id, name, description, source_type, pattern, enabled, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.SourceType,
		patternJSON, rule.Enabled, rule.CreatedAt, rule.UpdatedAt, metadataJSON,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create discovery rule: %w", err)
	}

	return &rule, nil
}

/* RunDiscovery runs automatic lineage discovery based on rules */
func (s *DiscoveryService) RunDiscovery(ctx context.Context, ruleID *uuid.UUID) ([]DiscoveredLineage, error) {
	var rules []DiscoveryRule

	if ruleID != nil {
		// Get specific rule
		query := `
			SELECT id, name, description, source_type, pattern, enabled, last_run_at, created_at, updated_at, metadata
			FROM neuronip.lineage_discovery_rules
			WHERE id = $1 AND enabled = true`

		var rule DiscoveryRule
		var patternJSON, metadataJSON []byte
		var lastRunAt interface{}

		err := s.pool.QueryRow(ctx, query, ruleID).Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.SourceType,
			&patternJSON, &rule.Enabled, &lastRunAt, &rule.CreatedAt, &rule.UpdatedAt, &metadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to get rule: %w", err)
		}

		json.Unmarshal(patternJSON, &rule.Pattern)
		json.Unmarshal(metadataJSON, &rule.Metadata)
		if lastRunAt != nil {
			if t, ok := lastRunAt.(time.Time); ok {
				rule.LastRunAt = &t
			}
		}

		rules = append(rules, rule)
	} else {
		// Get all enabled rules
		query := `
			SELECT id, name, description, source_type, pattern, enabled, last_run_at, created_at, updated_at, metadata
			FROM neuronip.lineage_discovery_rules
			WHERE enabled = true`

		rows, err := s.pool.Query(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to get rules: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var rule DiscoveryRule
			var patternJSON, metadataJSON []byte
			var lastRunAt interface{}

			err := rows.Scan(
				&rule.ID, &rule.Name, &rule.Description, &rule.SourceType,
				&patternJSON, &rule.Enabled, &lastRunAt, &rule.CreatedAt, &rule.UpdatedAt, &metadataJSON,
			)
			if err != nil {
				continue
			}

			json.Unmarshal(patternJSON, &rule.Pattern)
			json.Unmarshal(metadataJSON, &rule.Metadata)
			if lastRunAt != nil {
				if t, ok := lastRunAt.(time.Time); ok {
					rule.LastRunAt = &t
				}
			}

			rules = append(rules, rule)
		}
	}

	// Run discovery for each rule
	var discovered []DiscoveredLineage
	for _, rule := range rules {
		results, err := s.discoverByRule(ctx, rule)
		if err != nil {
			continue
		}
		discovered = append(discovered, results...)
	}

	// Update last_run_at for rules
	now := time.Now()
	for _, rule := range rules {
		s.pool.Exec(ctx,
			"UPDATE neuronip.lineage_discovery_rules SET last_run_at = $1, updated_at = $2 WHERE id = $3",
			now, now, rule.ID,
		)
	}

	return discovered, nil
}

/* discoverByRule performs discovery based on a specific rule */
func (s *DiscoveryService) discoverByRule(ctx context.Context, rule DiscoveryRule) ([]DiscoveredLineage, error) {
	var discovered []DiscoveredLineage

	switch rule.SourceType {
	case "query_log":
		// Analyze query logs to discover lineage
		results, err := s.discoverFromQueryLogs(ctx, rule)
		if err != nil {
			return nil, err
		}
		discovered = append(discovered, results...)

	case "sql_parser":
		// Parse SQL statements to discover lineage
		results, err := s.discoverFromSQL(ctx, rule)
		if err != nil {
			return nil, err
		}
		discovered = append(discovered, results...)

	case "etl_job":
		// Analyze ETL job definitions (dbt, Airflow, etc.)
		results, err := s.discoverFromETL(ctx, rule)
		if err != nil {
			return nil, err
		}
		discovered = append(discovered, results...)

	case "api_call":
		// Analyze API call logs to discover lineage
		results, err := s.discoverFromAPICalls(ctx, rule)
		if err != nil {
			return nil, err
		}
		discovered = append(discovered, results...)
	}

	// Save discovered lineage
	for i := range discovered {
		discovered[i].ID = uuid.New()
		discovered[i].RuleID = rule.ID
		discovered[i].CreatedAt = time.Now()
		discovered[i].Verified = false

		evidenceJSON, _ := json.Marshal(discovered[i].Evidence)

		_, err := s.pool.Exec(ctx, `
			INSERT INTO neuronip.discovered_lineage
			(id, rule_id, source_node_id, target_node_id, edge_type, confidence, evidence, verified, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			discovered[i].ID, discovered[i].RuleID, discovered[i].SourceNodeID,
			discovered[i].TargetNodeID, discovered[i].EdgeType, discovered[i].Confidence,
			evidenceJSON, discovered[i].Verified, discovered[i].CreatedAt,
		)
		if err != nil {
			continue
		}
	}

	return discovered, nil
}

/* discoverFromQueryLogs discovers lineage from query logs */
func (s *DiscoveryService) discoverFromQueryLogs(ctx context.Context, rule DiscoveryRule) ([]DiscoveredLineage, error) {
	// Query warehouse query history for SELECT ... INTO or CTE patterns
	query := `
		SELECT id, query_text, schema_id, created_at
		FROM neuronip.warehouse_queries
		WHERE query_text LIKE '%SELECT%INTO%' OR query_text LIKE '%CREATE TABLE%AS SELECT%'
		ORDER BY created_at DESC
		LIMIT 100`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discovered []DiscoveredLineage
	// Parse queries to extract source/target relationships
	// This would involve SQL parsing logic
	// For now, return empty - implementation would use SQL parser

	return discovered, nil
}

/* discoverFromSQL discovers lineage from SQL statements */
func (s *DiscoveryService) discoverFromSQL(ctx context.Context, rule DiscoveryRule) ([]DiscoveredLineage, error) {
	// SQL parsing logic would go here
	// For now, return empty
	return []DiscoveredLineage{}, nil
}

/* discoverFromETL discovers lineage from ETL job definitions */
func (s *DiscoveryService) discoverFromETL(ctx context.Context, rule DiscoveryRule) ([]DiscoveredLineage, error) {
	// Query ingestion jobs for ETL patterns
	query := `
		SELECT id, config, created_at
		FROM neuronip.ingestion_jobs
		WHERE config->>'connector_type' IN ('dbt', 'airflow', 'fivetran')
		ORDER BY created_at DESC
		LIMIT 100`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discovered []DiscoveredLineage
	// Parse ETL job configs to extract lineage
	// Implementation would parse dbt manifest, Airflow DAGs, etc.

	return discovered, nil
}

/* discoverFromAPICalls discovers lineage from API call logs */
func (s *DiscoveryService) discoverFromAPICalls(ctx context.Context, rule DiscoveryRule) ([]DiscoveredLineage, error) {
	// Query audit logs for API patterns
	// For now, return empty
	return []DiscoveredLineage{}, nil
}

/* VerifyDiscoveredLineage marks discovered lineage as verified and creates actual lineage edges */
func (s *DiscoveryService) VerifyDiscoveredLineage(ctx context.Context, discoveredID uuid.UUID) error {
	// Get discovered lineage
	var discovered DiscoveredLineage
	var evidenceJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT id, rule_id, source_node_id, target_node_id, edge_type, confidence, evidence, verified
		FROM neuronip.discovered_lineage
		WHERE id = $1`, discoveredID,
	).Scan(&discovered.ID, &discovered.RuleID, &discovered.SourceNodeID,
		&discovered.TargetNodeID, &discovered.EdgeType, &discovered.Confidence,
		&evidenceJSON, &discovered.Verified)

	if err != nil {
		return fmt.Errorf("failed to get discovered lineage: %w", err)
	}

	json.Unmarshal(evidenceJSON, &discovered.Evidence)

	// Create actual lineage edge
	edgeID := uuid.New()
	evidenceJSON, _ = json.Marshal(discovered.Evidence)

	_, err = s.pool.Exec(ctx, `
		INSERT INTO neuronip.lineage_edges
		(id, source_node_id, target_node_id, edge_type, transformation, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (source_node_id, target_node_id, edge_type) DO NOTHING`,
		edgeID, discovered.SourceNodeID, discovered.TargetNodeID,
		discovered.EdgeType, evidenceJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create lineage edge: %w", err)
	}

	// Mark as verified
	_, err = s.pool.Exec(ctx, `
		UPDATE neuronip.discovered_lineage
		SET verified = true
		WHERE id = $1`, discoveredID)

	return err
}
