package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides compliance and audit analytics functionality */
type Service struct {
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
}

/* NewService creates a new compliance service */
func NewService(pool *pgxpool.Pool, neurondbClient *neurondb.Client) *Service {
	return &Service{
		pool:           pool,
		neurondbClient: neurondbClient,
	}
}

/* ComplianceMatch represents a compliance match result */
type ComplianceMatch struct {
	PolicyID    uuid.UUID              `json:"policy_id"`
	PolicyName  string                 `json:"policy_name"`
	PolicyType  string                 `json:"policy_type"`
	MatchScore  float64                `json:"match_score"`
	MatchDetails map[string]interface{} `json:"match_details,omitempty"`
}

/* CheckCompliance checks compliance for an entity using policy matching */
func (s *Service) CheckCompliance(ctx context.Context, entityType string, entityID string, entityContent string) ([]ComplianceMatch, error) {
	// Get all enabled compliance policies
	query := `
		SELECT id, policy_name, policy_type, policy_text, embedding, rules
		FROM neuronip.compliance_policies
		WHERE enabled = true`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance policies: %w", err)
	}
	defer rows.Close()

	var matches []ComplianceMatch

	// Generate embedding for entity content
	var entityEmbedding string
	if entityContent != "" {
		var err error
		entityEmbedding, err = s.neurondbClient.GenerateEmbedding(ctx, entityContent, "sentence-transformers/all-MiniLM-L6-v2")
		if err != nil {
			// Continue without embedding-based matching if generation fails
			entityEmbedding = ""
		}
	}

	// Check each policy
	for rows.Next() {
		var policyID uuid.UUID
		var policyName, policyType, policyText string
		var policyEmbedding *string
		var rulesJSON json.RawMessage

		err := rows.Scan(&policyID, &policyName, &policyType, &policyText, &policyEmbedding, &rulesJSON)
		if err != nil {
			continue
		}

		var matchScore float64
		var matchDetails map[string]interface{}

		// Perform semantic matching if embeddings are available
		if entityEmbedding != "" && policyEmbedding != nil && *policyEmbedding != "" {
			similarityQuery := `SELECT 1 - ($1::vector <=> $2::vector) as similarity`
			err = s.pool.QueryRow(ctx, similarityQuery, entityEmbedding, *policyEmbedding).Scan(&matchScore)
			if err != nil {
				matchScore = 0
			}
		} else {
			// Fallback: simple text matching
			matchScore = s.simpleTextMatch(entityContent, policyText)
		}

		// Parse and check structured rules if available
		if rulesJSON != nil && len(rulesJSON) > 0 {
			var rules []map[string]interface{}
			if err := json.Unmarshal(rulesJSON, &rules); err == nil {
				ruleMatches := s.checkRules(rules, entityContent)
				// Combine semantic match with rule matches
				if len(ruleMatches) > 0 {
					ruleScore := float64(len(ruleMatches)) / float64(len(rules))
					matchScore = (matchScore + ruleScore) / 2.0
					matchDetails = map[string]interface{}{
						"rule_matches": ruleMatches,
					}
				}
			}
		}

		// Only include matches above threshold
		if matchScore >= 0.5 {
			matches = append(matches, ComplianceMatch{
				PolicyID:     policyID,
				PolicyName:   policyName,
				PolicyType:   policyType,
				MatchScore:   matchScore,
				MatchDetails: matchDetails,
			})

			// Store compliance match
			s.storeComplianceMatch(ctx, policyID, entityType, entityID, matchScore, matchDetails)
		}
	}

	return matches, nil
}

/* simpleTextMatch performs text-based matching using word overlap and similarity */
func (s *Service) simpleTextMatch(content, policy string) float64 {
	if len(content) == 0 || len(policy) == 0 {
		return 0.0
	}

	// Normalize text: lowercase, remove punctuation
	contentNormalized := normalizeText(content)
	policyNormalized := normalizeText(policy)

	// Split into words
	contentWords := strings.Fields(contentNormalized)
	policyWords := strings.Fields(policyNormalized)

	if len(contentWords) == 0 || len(policyWords) == 0 {
		return 0.0
	}

	// Calculate word overlap using Jaccard similarity
	contentSet := make(map[string]bool)
	for _, word := range contentWords {
		if len(word) > 2 { // Ignore very short words
			contentSet[word] = true
		}
	}

	policySet := make(map[string]bool)
	for _, word := range policyWords {
		if len(word) > 2 {
			policySet[word] = true
		}
	}

	// Calculate intersection and union
	intersection := 0
	for word := range contentSet {
		if policySet[word] {
			intersection++
		}
	}

	union := len(contentSet) + len(policySet) - intersection
	if union == 0 {
		return 0.0
	}

	jaccard := float64(intersection) / float64(union)

	// Also check for exact phrase matches (higher score)
	phraseScore := 0.0
	if strings.Contains(strings.ToLower(content), strings.ToLower(policy)) {
		phraseScore = 0.3
	}

	// Combine Jaccard and phrase scores
	return (jaccard * 0.7) + (phraseScore * 0.3)
}

/* normalizeText normalizes text for matching */
func normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	// Remove common punctuation
	text = strings.ReplaceAll(text, ".", " ")
	text = strings.ReplaceAll(text, ",", " ")
	text = strings.ReplaceAll(text, ";", " ")
	text = strings.ReplaceAll(text, ":", " ")
	text = strings.ReplaceAll(text, "!", " ")
	text = strings.ReplaceAll(text, "?", " ")
	text = strings.ReplaceAll(text, "(", " ")
	text = strings.ReplaceAll(text, ")", " ")
	// Collapse multiple spaces
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

/* checkRules checks structured compliance rules */
func (s *Service) checkRules(rules []map[string]interface{}, content string) []map[string]interface{} {
	var matches []map[string]interface{}
	for _, rule := range rules {
		if s.matchesRule(rule, content) {
			matches = append(matches, rule)
		}
	}
	return matches
}

/* matchesRule checks if content matches a compliance rule */
func (s *Service) matchesRule(rule map[string]interface{}, content string) bool {
	// Check pattern-based matching
	if pattern, ok := rule["pattern"].(string); ok && pattern != "" {
		patternType := "substring"
		if pt, ok := rule["pattern_type"].(string); ok {
			patternType = pt
		}

		switch patternType {
		case "regex":
			return matchesRegexPattern(content, pattern)
		case "keyword":
			return matchesKeywords(content, pattern)
		case "exact":
			return strings.Contains(strings.ToLower(content), strings.ToLower(pattern))
		default: // "substring"
			return containsPattern(content, pattern)
		}
	}

	// Check keyword-based matching
	if keywords, ok := rule["keywords"].([]interface{}); ok {
		return matchesKeywordList(content, keywords)
	}

	// Check numeric conditions
	if field, ok := rule["field"].(string); ok {
		if operator, ok := rule["operator"].(string); ok {
			if value, ok := rule["value"]; ok {
				return matchesNumericCondition(content, field, operator, value)
			}
		}
	}

	return false
}

/* containsPattern checks if content contains a pattern (substring match) */
func containsPattern(content, pattern string) bool {
	if len(pattern) == 0 || len(content) == 0 {
		return false
	}
	return strings.Contains(strings.ToLower(content), strings.ToLower(pattern))
}

/* matchesRegexPattern checks if content matches a regex pattern */
func matchesRegexPattern(content, pattern string) bool {
	re, err := regexp.Compile("(?i)" + pattern) // Case-insensitive
	if err != nil {
		// If regex is invalid, fall back to substring match
		return containsPattern(content, pattern)
	}
	return re.MatchString(content)
}

/* matchesKeywords checks if content contains keywords */
func matchesKeywords(content, keywords string) bool {
	keywordList := strings.Fields(keywords)
	contentLower := strings.ToLower(content)
	for _, keyword := range keywordList {
		if !strings.Contains(contentLower, strings.ToLower(keyword)) {
			return false
		}
	}
	return len(keywordList) > 0
}

/* matchesKeywordList checks if content contains all keywords from a list */
func matchesKeywordList(content string, keywords []interface{}) bool {
	contentLower := strings.ToLower(content)
	for _, kw := range keywords {
		if keyword, ok := kw.(string); ok {
			if !strings.Contains(contentLower, strings.ToLower(keyword)) {
				return false
			}
		}
	}
	return len(keywords) > 0
}

/* matchesNumericCondition checks numeric conditions by extracting numeric values from structured content */
func matchesNumericCondition(content, field, operator string, value interface{}) bool {
	// Try to parse content as JSON to extract structured field values
	var contentData map[string]interface{}
	if err := json.Unmarshal([]byte(content), &contentData); err == nil {
		// Content is JSON - extract field value
		if fieldValue, ok := contentData[field]; ok {
			return compareNumericValues(fieldValue, operator, value)
		}
	}
	
	// Fallback: try to find numeric patterns in text content
	// This is less precise but handles text-based content with embedded numbers
	re := regexp.MustCompile(fmt.Sprintf(`(?i)%s[:\s=]+([0-9]+\.?[0-9]*)`, regexp.QuoteMeta(field)))
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		if num, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return compareNumericValues(num, operator, value)
		}
	}
	
	return false
}

/* compareNumericValues compares two numeric values using the specified operator */
func compareNumericValues(left interface{}, operator string, right interface{}) bool {
	leftNum, leftOk := toFloat64(left)
	rightNum, rightOk := toFloat64(right)
	
	if !leftOk || !rightOk {
		return false
	}
	
	switch operator {
	case "gt", ">":
		return leftNum > rightNum
	case "gte", ">=":
		return leftNum >= rightNum
	case "lt", "<":
		return leftNum < rightNum
	case "lte", "<=":
		return leftNum <= rightNum
	case "eq", "==":
		return leftNum == rightNum
	case "ne", "!=":
		return leftNum != rightNum
	default:
		return false
	}
}

/* toFloat64 converts various numeric types to float64 */
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case string:
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

/* storeComplianceMatch stores a compliance match result */
func (s *Service) storeComplianceMatch(ctx context.Context, policyID uuid.UUID, entityType, entityID string, matchScore float64, details map[string]interface{}) error {
	detailsJSON, _ := json.Marshal(details)

	query := `
		INSERT INTO neuronip.compliance_matches 
		(policy_id, entity_type, entity_id, match_score, match_details, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', NOW())
		ON CONFLICT DO NOTHING`

	_, err := s.pool.Exec(ctx, query, policyID, entityType, entityID, matchScore, detailsJSON)
	return err
}
