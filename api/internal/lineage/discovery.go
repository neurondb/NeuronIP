package lineage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	SourceType  string                 `json:"source_type"` // "query_log", "sql_parser", "api_call", "etl_job"
	Pattern     map[string]interface{} `json:"pattern"`     // Pattern matching rules
	Enabled     bool                   `json:"enabled"`
	LastRunAt   *time.Time             `json:"last_run_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
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

/* discoverFromQueryLogs discovers lineage from query logs by parsing SQL to extract source/target tables */
func (s *DiscoveryService) discoverFromQueryLogs(ctx context.Context, rule DiscoveryRule) ([]DiscoveredLineage, error) {
	query := `
		SELECT id, query_text, created_at
		FROM neuronip.warehouse_queries
		WHERE query_text ILIKE '%SELECT%INTO%'
		   OR query_text ILIKE '%CREATE TABLE%AS SELECT%'
		   OR query_text ILIKE '%INSERT%INTO%SELECT%'
		ORDER BY created_at DESC
		LIMIT 100`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discovered []DiscoveredLineage
	for rows.Next() {
		var queryID uuid.UUID
		var queryText string
		var createdAt time.Time
		if err := rows.Scan(&queryID, &queryText, &createdAt); err != nil {
			continue
		}

		sourceTables, targetTable := s.parseSQLForLineage(queryText)
		if targetTable == "" || len(sourceTables) == 0 {
			continue
		}

		targetNodeID, err := s.ensureLineageNode(ctx, targetTable, "target")
		if err != nil {
			continue
		}

		for _, src := range sourceTables {
			sourceNodeID, err := s.ensureLineageNode(ctx, src, "source")
			if err != nil {
				continue
			}
			discovered = append(discovered, DiscoveredLineage{
				SourceNodeID: sourceNodeID,
				TargetNodeID: targetNodeID,
				EdgeType:     "reads",
				Confidence:   0.9,
				Evidence: map[string]interface{}{
					"query_id":   queryID.String(),
					"query_text": queryText,
					"created_at": createdAt.Format(time.RFC3339),
				},
			})
		}
	}

	return discovered, nil
}

/* discoverFromSQL discovers lineage by parsing SQL statements stored in audit logs */
func (s *DiscoveryService) discoverFromSQL(ctx context.Context, rule DiscoveryRule) ([]DiscoveredLineage, error) {
	query := `
		SELECT id, details->>'query' as query_text, created_at
		FROM neuronip.audit_logs
		WHERE action_type = 'query'
		AND details->>'query' IS NOT NULL
		AND created_at > NOW() - INTERVAL '7 days'
		ORDER BY created_at DESC
		LIMIT 200`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return []DiscoveredLineage{}, nil
	}
	defer rows.Close()

	var discovered []DiscoveredLineage
	for rows.Next() {
		var auditID uuid.UUID
		var queryText *string
		var createdAt time.Time
		if err := rows.Scan(&auditID, &queryText, &createdAt); err != nil || queryText == nil {
			continue
		}

		sourceTables, targetTable := s.parseSQLForLineage(*queryText)
		if targetTable == "" || len(sourceTables) == 0 {
			continue
		}

		targetNodeID, err := s.ensureLineageNode(ctx, targetTable, "target")
		if err != nil {
			continue
		}

		for _, src := range sourceTables {
			sourceNodeID, err := s.ensureLineageNode(ctx, src, "source")
			if err != nil {
				continue
			}
			discovered = append(discovered, DiscoveredLineage{
				SourceNodeID: sourceNodeID,
				TargetNodeID: targetNodeID,
				EdgeType:     "reads",
				Confidence:   0.85,
				Evidence: map[string]interface{}{
					"audit_id":   auditID.String(),
					"query_text": *queryText,
					"created_at": createdAt.Format(time.RFC3339),
				},
			})
		}
	}

	return discovered, nil
}

/* parseSQLForLineage parses SQL to extract source and target table names */
func (s *DiscoveryService) parseSQLForLineage(queryText string) (sourceTables []string, targetTable string) {
	lowerQuery := strings.ToLower(queryText)

	// CREATE TABLE ... AS SELECT
	if strings.Contains(lowerQuery, "create table") && strings.Contains(lowerQuery, "as select") {
		targetTable = extractTableAfterKeyword(queryText, "create table")
		sourceTables = extractTablesAfterKeyword(queryText, "from")
		return
	}

	// SELECT ... INTO
	if strings.Contains(lowerQuery, "into") && strings.Contains(lowerQuery, "select") {
		targetTable = extractTableAfterKeyword(queryText, "into")
		sourceTables = extractTablesAfterKeyword(queryText, "from")
		return
	}

	// INSERT INTO ... SELECT
	if strings.Contains(lowerQuery, "insert into") && strings.Contains(lowerQuery, "select") {
		targetTable = extractTableAfterKeyword(queryText, "insert into")
		sourceTables = extractTablesAfterKeyword(queryText, "from")
		return
	}

	return nil, ""
}

/* extractTableAfterKeyword extracts a single table name after a keyword */
func extractTableAfterKeyword(query, keyword string) string {
	lowerQuery := strings.ToLower(query)
	idx := strings.Index(lowerQuery, strings.ToLower(keyword))
	if idx == -1 {
		return ""
	}
	after := strings.TrimSpace(query[idx+len(keyword):])
	// Take the first word (table name)
	parts := strings.Fields(after)
	if len(parts) == 0 {
		return ""
	}
	tableName := parts[0]
	tableName = strings.Trim(tableName, "`\"[]")
	// Remove trailing parenthesis/comma
	tableName = strings.TrimRight(tableName, "(,;")
	return tableName
}

/* extractTablesAfterKeyword extracts table names after FROM/JOIN clauses */
func extractTablesAfterKeyword(query, keyword string) []string {
	lowerQuery := strings.ToLower(query)
	idx := strings.Index(lowerQuery, strings.ToLower(keyword))
	if idx == -1 {
		return nil
	}
	after := strings.TrimSpace(query[idx+len(keyword):])

	// Find end of FROM clause (WHERE, GROUP BY, ORDER BY, LIMIT, ;, or end)
	endKeywords := []string{"where", "group by", "order by", "limit", "having", ";"}
	endIdx := len(after)
	for _, ek := range endKeywords {
		if pos := strings.Index(strings.ToLower(after), ek); pos != -1 && pos < endIdx {
			endIdx = pos
		}
	}
	fromClause := after[:endIdx]

	// Split by JOIN and comma to get individual tables
	fromClause = strings.ReplaceAll(strings.ToLower(fromClause), " join ", ",")
	fromClause = strings.ReplaceAll(fromClause, " left ", " ")
	fromClause = strings.ReplaceAll(fromClause, " right ", " ")
	fromClause = strings.ReplaceAll(fromClause, " inner ", " ")
	fromClause = strings.ReplaceAll(fromClause, " outer ", " ")
	fromClause = strings.ReplaceAll(fromClause, " on ", ",")

	tableParts := strings.Split(fromClause, ",")
	var tables []string
	for _, part := range tableParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Take first word as table name, skip aliases
		words := strings.Fields(part)
		if len(words) == 0 {
			continue
		}
		tableName := words[0]
		tableName = strings.Trim(tableName, "`\"[]")
		tableName = strings.TrimRight(tableName, "(,;")
		if tableName != "" && !strings.HasPrefix(tableName, "(") {
			tables = append(tables, tableName)
		}
	}
	return tables
}

/* ensureLineageNode creates or retrieves a lineage node for the given table name */
func (s *DiscoveryService) ensureLineageNode(ctx context.Context, nodeName, nodeType string) (uuid.UUID, error) {
	var nodeID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM neuronip.lineage_nodes WHERE node_name = $1 AND node_type = $2
	`, nodeName, nodeType).Scan(&nodeID)
	if err == nil {
		return nodeID, nil
	}

	nodeID = uuid.New()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO neuronip.lineage_nodes (id, node_type, node_name, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT DO NOTHING
	`, nodeID, nodeType, nodeName)
	if err != nil {
		return uuid.Nil, err
	}
	return nodeID, nil
}

/* discoverFromETL discovers lineage from ETL job definitions by parsing config for source/target tables */
func (s *DiscoveryService) discoverFromETL(ctx context.Context, rule DiscoveryRule) ([]DiscoveredLineage, error) {
	query := `
		SELECT id, config, created_at
		FROM neuronip.ingestion_jobs
		WHERE status = 'completed'
		AND created_at > NOW() - INTERVAL '30 days'
		ORDER BY created_at DESC
		LIMIT 100`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discovered []DiscoveredLineage
	for rows.Next() {
		var jobID uuid.UUID
		var configJSON []byte
		var createdAt time.Time
		if err := rows.Scan(&jobID, &configJSON, &createdAt); err != nil {
			continue
		}

		var config map[string]interface{}
		if err := json.Unmarshal(configJSON, &config); err != nil {
			continue
		}

		// Extract source tables from config
		var sourceTables []string
		if tables, ok := config["tables"].([]interface{}); ok {
			for _, t := range tables {
				if tStr, ok := t.(string); ok {
					sourceTables = append(sourceTables, tStr)
				}
			}
		}
		if src, ok := config["source_table"].(string); ok {
			sourceTables = append(sourceTables, src)
		}

		// Extract target table from config
		targetTable := ""
		if tgt, ok := config["target_table"].(string); ok {
			targetTable = tgt
		} else if tgt, ok := config["destination"].(string); ok {
			targetTable = tgt
		}

		if targetTable == "" || len(sourceTables) == 0 {
			continue
		}

		targetNodeID, err := s.ensureLineageNode(ctx, targetTable, "target")
		if err != nil {
			continue
		}

		for _, src := range sourceTables {
			sourceNodeID, err := s.ensureLineageNode(ctx, src, "source")
			if err != nil {
				continue
			}
			discovered = append(discovered, DiscoveredLineage{
				SourceNodeID: sourceNodeID,
				TargetNodeID: targetNodeID,
				EdgeType:     "transforms",
				Confidence:   0.95,
				Evidence: map[string]interface{}{
					"job_id":     jobID.String(),
					"created_at": createdAt.Format(time.RFC3339),
				},
			})
		}
	}

	return discovered, nil
}

/* discoverFromAPICalls discovers lineage from API call logs. Creates or reuses lineage nodes for each resource and links them from a synthetic api_access source node. */
func (s *DiscoveryService) discoverFromAPICalls(ctx context.Context, rule DiscoveryRule) ([]DiscoveredLineage, error) {
	sourceNodeID, err := s.ensureAPIAccessNode(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, action_type, resource_type, resource_id, created_at
		FROM neuronip.audit_logs
		WHERE action_type IN ('data_access', 'query', 'workflow_execution')
		AND created_at > NOW() - INTERVAL '7 days'
		ORDER BY created_at DESC
		LIMIT 200`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return []DiscoveredLineage{}, nil
	}
	defer rows.Close()

	var discovered []DiscoveredLineage
	seen := make(map[string]struct{})
	for rows.Next() {
		var id uuid.UUID
		var actionType, resourceType, resourceID string
		var createdAt time.Time
		if err := rows.Scan(&id, &actionType, &resourceType, &resourceID, &createdAt); err != nil {
			continue
		}
		if resourceID == "" {
			continue
		}
		key := resourceType + ":" + resourceID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		targetNodeID, err := s.ensureResourceLineageNode(ctx, resourceType, resourceID)
		if err != nil {
			continue
		}
		edgeType := "reads"
		if actionType == "workflow_execution" {
			edgeType = "depends_on"
		}
		discovered = append(discovered, DiscoveredLineage{
			RuleID:       rule.ID,
			SourceNodeID: sourceNodeID,
			TargetNodeID: targetNodeID,
			EdgeType:     edgeType,
			Confidence:   0.6,
			Evidence:     map[string]interface{}{"audit_id": id.String(), "action_type": actionType, "resource_type": resourceType, "resource_id": resourceID},
			Verified:     false,
			CreatedAt:    time.Now(),
		})
	}
	return discovered, nil
}

func (s *DiscoveryService) ensureAPIAccessNode(ctx context.Context) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM neuronip.lineage_nodes
		WHERE node_type = 'source' AND node_name = 'api_access' AND metadata->>'discovery' = 'api_call'
		LIMIT 1`).Scan(&id)
	if err == nil {
		return id, nil
	}
	id = uuid.New()
	meta := map[string]interface{}{"discovery": "api_call"}
	metaJSON, _ := json.Marshal(meta)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO neuronip.lineage_nodes (id, node_type, node_name, schema_info, metadata, created_at)
		VALUES ($1, 'source', 'api_access', '{}', $2, NOW())`, id, metaJSON)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *DiscoveryService) ensureResourceLineageNode(ctx context.Context, resourceType, resourceID string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM neuronip.lineage_nodes
		WHERE metadata->>'resource_id' = $1 AND metadata->>'resource_type' = $2
		LIMIT 1`, resourceID, resourceType).Scan(&id)
	if err == nil {
		return id, nil
	}
	id = uuid.New()
	nodeType := "table"
	if resourceType == "workflow" || resourceType == "query" {
		nodeType = "transformation"
	}
	meta := map[string]interface{}{"resource_id": resourceID, "resource_type": resourceType}
	metaJSON, _ := json.Marshal(meta)
	nodeName := resourceID
	if resourceType != "" {
		nodeName = resourceType + ":" + resourceID
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO neuronip.lineage_nodes (id, node_type, node_name, schema_info, metadata, created_at)
		VALUES ($1, $2, $3, '{}', $4, NOW())`, id, nodeType, nodeName, metaJSON)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
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
