package knowledgegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* GraphDBService provides graph database integration (PostgreSQL-native) */
type GraphDBService struct {
	pool *pgxpool.Pool
}

/* NewGraphDBService creates a new graph database service */
func NewGraphDBService(pool *pgxpool.Pool) *GraphDBService {
	return &GraphDBService{pool: pool}
}

/* ExecuteCypherQuery executes a Cypher-like query using PostgreSQL recursive CTEs.
 * Supported patterns:
 *   - MATCH (n) RETURN n
 *   - MATCH (n) WHERE n.type = 'X' RETURN n
 *   - MATCH (a)-[r]->(b) RETURN a, r, b
 *   - MATCH (a)-[r*1..depth]->(b) RETURN a, r, b  (params: startId, depth, relationshipType)
 * Params: startId (UUID string), depth (int), relationshipType (string), entityType (string)
 */
func (gds *GraphDBService) ExecuteCypherQuery(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	query = strings.TrimSpace(query)
	qUpper := strings.ToUpper(query)

	// MATCH (n) RETURN n — list entities, optional WHERE by type
	if strings.Contains(qUpper, "MATCH") && strings.Contains(qUpper, "RETURN") {
		// Variable-length path: MATCH (a)-[r*1..3]->(b) or similar
		depthRe := regexp.MustCompile(`(?i)\[\s*\w*\s*\*\s*(\d+)\s*\.\.\s*(\d+)\s*\]`)
		if depthRe.MatchString(query) {
			return gds.traverseCypherStyle(ctx, query, params)
		}
		// (a)-[r]->(b) pattern
		if strings.Contains(query, "-") && strings.Contains(query, "->") {
			return gds.edgeQueryCypherStyle(ctx, query, params)
		}
		// (n) or (a) single node pattern
		return gds.nodeQueryCypherStyle(ctx, query, params)
	}

	return nil, fmt.Errorf("unsupported Cypher pattern: query must contain MATCH and RETURN")
}

/* traverseCypherStyle runs a variable-length traversal and returns rows as maps */
func (gds *GraphDBService) traverseCypherStyle(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	startID, _ := paramUUID(params, "startId")
	if startID == nil {
		startID, _ = paramUUID(params, "start_id")
	}
	depth := paramInt(params, "depth", 3)
	relType := paramString(params, "relationshipType")
	if relType == "" {
		relType = paramString(params, "relationship_type")
	}

	rows, err := gds.traverseRecursive(ctx, startID, depth, relType, "")
	if err != nil {
		return nil, err
	}
	return rows, nil
}

/* edgeQueryCypherStyle returns (a)-[r]->(b) style results */
func (gds *GraphDBService) edgeQueryCypherStyle(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	relFilter := paramString(params, "relationshipType")
	if relFilter == "" {
		relFilter = paramString(params, "relationship_type")
	}

	sql := `
		SELECT 
			e1.id AS source_id, e1.entity_name AS source_name, e1.entity_type_id AS source_type_id,
			el.id AS link_id, el.relationship_type, el.relationship_strength,
			e2.id AS target_id, e2.entity_name AS target_name, e2.entity_type_id AS target_type_id
		FROM neuronip.entity_links el
		JOIN neuronip.entities e1 ON e1.id = el.source_entity_id
		JOIN neuronip.entities e2 ON e2.id = el.target_entity_id
		WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if relFilter != "" {
		sql += fmt.Sprintf(" AND el.relationship_type = $%d", argNum)
		args = append(args, relFilter)
		argNum++
	}
	sql += " LIMIT 500"

	rows, err := gds.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var srcID, tgtID, linkID uuid.UUID
		var srcName, tgtName, relType string
		var strength float64
		var srcTypeID, tgtTypeID *uuid.UUID
		if err := rows.Scan(&srcID, &srcName, &srcTypeID, &linkID, &relType, &strength, &tgtID, &tgtName, &tgtTypeID); err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"a": map[string]interface{}{"id": srcID.String(), "entity_name": srcName, "entity_type_id": srcTypeID},
			"r": map[string]interface{}{"id": linkID.String(), "relationship_type": relType, "relationship_strength": strength},
			"b": map[string]interface{}{"id": tgtID.String(), "entity_name": tgtName, "entity_type_id": tgtTypeID},
		})
	}
	return result, nil
}

/* nodeQueryCypherStyle returns MATCH (n) style entity list */
func (gds *GraphDBService) nodeQueryCypherStyle(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	entityType := paramString(params, "entityType")
	if entityType == "" {
		entityType = paramString(params, "entity_type")
	}

	sql := `SELECT id, entity_name, entity_type_id, entity_value, description, metadata, confidence_score, created_at
			FROM neuronip.entities WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if entityType != "" {
		sql += fmt.Sprintf(" AND entity_type_id = (SELECT id FROM neuronip.entity_types WHERE type_name = $%d)", argNum)
		args = append(args, entityType)
		argNum++
	}
	sql += " LIMIT 500"

	rows, err := gds.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var name string
		var typeID *uuid.UUID
		var value, desc *string
		var meta json.RawMessage
		var confidence float64
		var createdAt interface{}
		if err := rows.Scan(&id, &name, &typeID, &value, &desc, &meta, &confidence, &createdAt); err != nil {
			continue
		}
		n := map[string]interface{}{"id": id.String(), "entity_name": name, "entity_type_id": typeID, "entity_value": value, "description": desc, "confidence_score": confidence}
		if len(meta) > 0 {
			var m map[string]interface{}
			_ = json.Unmarshal(meta, &m)
			n["metadata"] = m
		}
		result = append(result, map[string]interface{}{"n": n})
	}
	return result, nil
}

/* traverseRecursive returns all nodes and edges reachable from startID within depth using recursive CTE */
func (gds *GraphDBService) traverseRecursive(ctx context.Context, startID *uuid.UUID, maxDepth int, relationshipType, direction string) ([]map[string]interface{}, error) {
	if maxDepth < 1 {
		maxDepth = 3
	}
	if startID == nil {
		// No start: return all edges up to limit
		return gds.edgeQueryCypherStyle(ctx, "MATCH (a)-[r]->(b) RETURN a,r,b", map[string]interface{}{"relationshipType": relationshipType})
	}

	// Single recursive CTE: (entity_id, depth, path, is_source)
	// We use entity_links and recurse; direction can be "out", "in", "both"
	sql := `
	WITH RECURSIVE traverse AS (
		SELECT 
			$1::uuid AS entity_id,
			0 AS depth,
			ARRAY[$1::uuid] AS path,
			$1::uuid AS from_entity_id,
			$1::uuid AS to_entity_id,
			NULL::uuid AS link_id,
			NULL::text AS relationship_type,
			NULL::float8 AS relationship_strength
		UNION ALL
		SELECT 
			CASE WHEN el.source_entity_id = t.entity_id THEN el.target_entity_id ELSE el.source_entity_id END,
			t.depth + 1,
			t.path || CASE WHEN el.source_entity_id = t.entity_id THEN el.target_entity_id ELSE el.source_entity_id END,
			el.source_entity_id,
			el.target_entity_id,
			el.id,
			el.relationship_type,
			el.relationship_strength
		FROM traverse t
		JOIN neuronip.entity_links el ON (
			(el.source_entity_id = t.entity_id OR el.target_entity_id = t.entity_id)
			AND t.depth < $2
			AND (array_length(t.path, 1) IS NULL OR NOT (CASE WHEN el.source_entity_id = t.entity_id THEN el.target_entity_id ELSE el.source_entity_id END = ANY(t.path)))
		)
		WHERE (CASE WHEN $3::text = '' THEN true ELSE el.relationship_type = $3 END)
	)
	SELECT entity_id, depth, from_entity_id, to_entity_id, link_id, relationship_type, relationship_strength
	FROM traverse
	WHERE depth > 0
	ORDER BY depth, entity_id
	LIMIT 1000`
	rows, err := gds.pool.Query(ctx, sql, startID, maxDepth, relationshipType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	seen := make(map[string]bool)
	for rows.Next() {
		var entityID, fromID, toID uuid.UUID
		var linkID *uuid.UUID
		var relType *string
		var strength *float64
		var depth int
		if err := rows.Scan(&entityID, &depth, &fromID, &toID, &linkID, &relType, &strength); err != nil {
			continue
		}
		key := fromID.String() + "|" + toID.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		r := map[string]interface{}{
			"from_entity_id": fromID.String(),
			"to_entity_id":   toID.String(),
			"depth":          depth,
		}
		if linkID != nil {
			r["link_id"] = linkID.String()
		}
		if relType != nil {
			r["relationship_type"] = *relType
		}
		if strength != nil {
			r["relationship_strength"] = *strength
		}
		result = append(result, r)
	}
	return result, nil
}

/* ExecuteGremlinQuery executes a Gremlin-like traversal using PostgreSQL.
 * Supported: g.V(startId).out() / .in() / .both() / .out(relationshipType) / .limit(n)
 */
func (gds *GraphDBService) ExecuteGremlinQuery(ctx context.Context, query string) ([]map[string]interface{}, error) {
	query = strings.TrimSpace(query)
	// Very simple parser: g.V(...).out() or g.V().out('rel')
	// Pattern: g.V( <uuid> ).out( optional 'label' ) or .in() / .both()
	var startID *uuid.UUID
	if idx := strings.Index(query, "g.V("); idx >= 0 {
		rest := query[idx+4:]
		closeIdx := strings.Index(rest, ")")
		if closeIdx > 0 {
			arg := strings.TrimSpace(rest[:closeIdx])
			if id, err := uuid.Parse(arg); err == nil {
				startID = &id
			}
		}
	}
	direction := "both"
	relType := ""
	if strings.Contains(query, ".out(") {
		direction = "out"
		relType = extractStringArg(query, ".out(")
	} else if strings.Contains(query, ".in(") {
		direction = "in"
		relType = extractStringArg(query, ".in(")
	} else if strings.Contains(query, ".out()") {
		direction = "out"
	} else if strings.Contains(query, ".in()") {
		direction = "in"
	} else if strings.Contains(query, ".both()") || strings.Contains(query, ".both(") {
		direction = "both"
		relType = extractStringArg(query, ".both(")
	}
	limit := 100
	if l := extractLimit(query); l > 0 {
		limit = l
	}

	if startID == nil {
		// Return all edges if no start vertex
		return gds.edgeQueryCypherStyle(ctx, "", map[string]interface{}{"relationshipType": relType})
	}

	var sql string
	var args []interface{}
	if direction == "out" {
		sql = `SELECT el.id, el.source_entity_id, el.target_entity_id, el.relationship_type, el.relationship_strength,
			   e.id AS entity_id, e.entity_name, e.entity_type_id, e.entity_value
			   FROM neuronip.entity_links el
			   JOIN neuronip.entities e ON e.id = el.target_entity_id
			   WHERE el.source_entity_id = $1`
		args = []interface{}{startID}
		if relType != "" {
			sql += " AND el.relationship_type = $2"
			args = append(args, relType)
		}
		sql += fmt.Sprintf(" LIMIT %d", limit)
	} else if direction == "in" {
		sql = `SELECT el.id, el.source_entity_id, el.target_entity_id, el.relationship_type, el.relationship_strength,
			   e.id AS entity_id, e.entity_name, e.entity_type_id, e.entity_value
			   FROM neuronip.entity_links el
			   JOIN neuronip.entities e ON e.id = el.source_entity_id
			   WHERE el.target_entity_id = $1`
		args = []interface{}{startID}
		if relType != "" {
			sql += " AND el.relationship_type = $2"
			args = append(args, relType)
		}
		sql += fmt.Sprintf(" LIMIT %d", limit)
	} else {
		sql = `SELECT el.id, el.source_entity_id, el.target_entity_id, el.relationship_type, el.relationship_strength,
			   e.id AS entity_id, e.entity_name, e.entity_type_id, e.entity_value
			   FROM neuronip.entity_links el
			   JOIN neuronip.entities e ON (e.id = el.source_entity_id OR e.id = el.target_entity_id) AND (e.id != $1)
			   WHERE el.source_entity_id = $1 OR el.target_entity_id = $1`
		args = []interface{}{startID}
		if relType != "" {
			sql += " AND el.relationship_type = $2"
			args = append(args, relType)
		}
		sql += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := gds.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var linkID, srcID, tgtID uuid.UUID
		var relType string
		var strength float64
		var entityID uuid.UUID
		var name string
		var typeID *uuid.UUID
		var value *string
		if err := rows.Scan(&linkID, &srcID, &tgtID, &relType, &strength, &entityID, &name, &typeID, &value); err != nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":                    entityID.String(),
			"entity_name":           name,
			"entity_type_id":        typeID,
			"entity_value":          value,
			"link_id":               linkID.String(),
			"relationship_type":     relType,
			"relationship_strength": strength,
		})
	}
	return result, nil
}

func paramUUID(params map[string]interface{}, key string) (*uuid.UUID, bool) {
	if params == nil {
		return nil, false
	}
	v, ok := params[key]
	if !ok {
		return nil, false
	}
	switch t := v.(type) {
	case string:
		id, err := uuid.Parse(t)
		if err != nil {
			return nil, false
		}
		return &id, true
	case uuid.UUID:
		return &t, true
	default:
		return nil, false
	}
}

func paramInt(params map[string]interface{}, key string, defaultVal int) int {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(t)
		return n
	default:
		return defaultVal
	}
}

func paramString(params map[string]interface{}, key string) string {
	if params == nil {
		return ""
	}
	v, ok := params[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func extractStringArg(query, prefix string) string {
	i := strings.Index(query, prefix)
	if i < 0 {
		return ""
	}
	start := i + len(prefix)
	rest := query[start:]
	// 'label' or "label"
	for _, quote := range []rune{'\'', '"'} {
		if strings.HasPrefix(rest, string(quote)) {
			end := strings.Index(rest[1:], string(quote))
			if end >= 0 {
				return rest[1 : end+1]
			}
		}
	}
	return ""
}

func extractLimit(query string) int {
	re := regexp.MustCompile(`\.limit\s*\(\s*(\d+)\s*\)`)
	m := re.FindStringSubmatch(strings.ToLower(query))
	if len(m) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}
