package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* ComplianceAnalyticsService provides enhanced compliance analytics */
type ComplianceAnalyticsService struct {
	pool *pgxpool.Pool
}

/* NewComplianceAnalyticsService creates a new compliance analytics service */
func NewComplianceAnalyticsService(pool *pgxpool.Pool) *ComplianceAnalyticsService {
	return &ComplianceAnalyticsService{pool: pool}
}

/* ComplianceDashboard represents compliance dashboard data */
type ComplianceDashboard struct {
	Summary          ComplianceSummary       `json:"summary"`
	Trends           []ComplianceTrend       `json:"trends,omitempty"`
	TopViolations    []ComplianceViolation   `json:"top_violations,omitempty"`
	PolicyCompliance []PolicyCompliance      `json:"policy_compliance,omitempty"`
	RiskAreas        []RiskArea              `json:"risk_areas,omitempty"`
	GeneratedAt      time.Time               `json:"generated_at"`
}

/* ComplianceSummary represents compliance summary metrics */
type ComplianceSummary struct {
	TotalPolicies      int     `json:"total_policies"`
	ActivePolicies     int     `json:"active_policies"`
	CompliantPolicies  int     `json:"compliant_policies"`
	NonCompliantPolicies int   `json:"non_compliant_policies"`
	ComplianceRate     float64 `json:"compliance_rate"` // Percentage
	TotalViolations    int     `json:"total_violations"`
	CriticalViolations int     `json:"critical_violations"`
	AnomaliesDetected  int     `json:"anomalies_detected"`
}

/* ComplianceTrend represents compliance trends over time */
type ComplianceTrend struct {
	Date           time.Time `json:"date"`
	ComplianceRate float64   `json:"compliance_rate"`
	Violations     int       `json:"violations"`
	Anomalies      int       `json:"anomalies"`
}

/* ComplianceViolation represents a compliance violation */
type ComplianceViolation struct {
	ID            string                 `json:"id"`
	PolicyID      string                 `json:"policy_id"`
	PolicyName    string                 `json:"policy_name"`
	ResourceID    string                 `json:"resource_id"`
	ResourceType  string                 `json:"resource_type"`
	Severity      string                 `json:"severity"`
	Description   string                 `json:"description"`
	DetectedAt    time.Time              `json:"detected_at"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* PolicyCompliance represents policy compliance status */
type PolicyCompliance struct {
	PolicyID       string    `json:"policy_id"`
	PolicyName     string    `json:"policy_name"`
	ComplianceRate float64   `json:"compliance_rate"`
	LastChecked    time.Time `json:"last_checked"`
	Status         string    `json:"status"` // "compliant", "non_compliant", "warning"
	ViolationCount int       `json:"violation_count"`
}

/* RiskArea represents a high-risk area */
type RiskArea struct {
	Area          string    `json:"area"`
	RiskScore     float64   `json:"risk_score"` // 0.0 to 1.0
	ViolationCount int      `json:"violation_count"`
	LastIncident  time.Time `json:"last_incident"`
	Description   string    `json:"description"`
}

/* GetComplianceDashboard retrieves comprehensive compliance dashboard data */
func (s *ComplianceAnalyticsService) GetComplianceDashboard(ctx context.Context,
	startTime, endTime time.Time) (*ComplianceDashboard, error) {

	if endTime.IsZero() {
		endTime = time.Now()
	}
	if startTime.IsZero() {
		startTime = endTime.AddDate(0, 0, -30) // Last 30 days by default
	}

	dashboard := &ComplianceDashboard{
		GeneratedAt: time.Now(),
	}

	// Get summary
	summary, err := s.getComplianceSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}
	dashboard.Summary = *summary

	// Get trends
	trends, err := s.getComplianceTrends(ctx, startTime, endTime)
	if err == nil {
		dashboard.Trends = trends
	}

	// Get top violations
	violations, err := s.getTopViolations(ctx, 10)
	if err == nil {
		dashboard.TopViolations = violations
	}

	// Get policy compliance
	policyCompliance, err := s.getPolicyCompliance(ctx)
	if err == nil {
		dashboard.PolicyCompliance = policyCompliance
	}

	// Get risk areas
	riskAreas, err := s.getRiskAreas(ctx)
	if err == nil {
		dashboard.RiskAreas = riskAreas
	}

	return dashboard, nil
}

/* getComplianceSummary retrieves compliance summary */
func (s *ComplianceAnalyticsService) getComplianceSummary(ctx context.Context) (*ComplianceSummary, error) {
	var summary ComplianceSummary

	// Get policy counts
	err := s.pool.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE enabled = true) as active,
			COUNT(*) FILTER (WHERE status = 'compliant') as compliant,
			COUNT(*) FILTER (WHERE status = 'non_compliant') as non_compliant
		FROM neuronip.compliance_policies`).Scan(
		&summary.TotalPolicies, &summary.ActivePolicies,
		&summary.CompliantPolicies, &summary.NonCompliantPolicies,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get policy counts: %w", err)
	}

	// Calculate compliance rate
	if summary.ActivePolicies > 0 {
		summary.ComplianceRate = float64(summary.CompliantPolicies) / float64(summary.ActivePolicies) * 100
	}

	// Get violation counts
	s.pool.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE severity = 'critical') as critical
		FROM neuronip.compliance_violations
		WHERE resolved_at IS NULL`).Scan(
		&summary.TotalViolations, &summary.CriticalViolations,
	)

	// Get anomaly count
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM neuronip.anomaly_detections
		WHERE detected_at > NOW() - INTERVAL '7 days'`).Scan(&summary.AnomaliesDetected)

	return &summary, nil
}

/* getComplianceTrends retrieves compliance trends */
func (s *ComplianceAnalyticsService) getComplianceTrends(ctx context.Context,
	startTime, endTime time.Time) ([]ComplianceTrend, error) {

	query := `
		SELECT 
			DATE(detected_at) as date,
			COUNT(*) as violations,
			COUNT(*) FILTER (WHERE EXISTS (
				SELECT 1 FROM neuronip.compliance_policies cp
				WHERE cp.id::text = cv.policy_id
			)) as policy_violations
		FROM neuronip.compliance_violations cv
		WHERE detected_at >= $1 AND detected_at <= $2
		GROUP BY DATE(detected_at)
		ORDER BY date ASC`

	rows, err := s.pool.Query(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get trends: %w", err)
	}
	defer rows.Close()

	var trends []ComplianceTrend
	for rows.Next() {
		var trend ComplianceTrend
		var date time.Time
		var violations int

		err := rows.Scan(&date, &violations, &trend.Anomalies)
		if err != nil {
			continue
		}

		trend.Date = date
		trend.Violations = violations

		// Calculate compliance rate for this day
		trend.ComplianceRate = 100.0 // Default, would calculate based on policies

		trends = append(trends, trend)
	}

	return trends, nil
}

/* getTopViolations retrieves top violations */
func (s *ComplianceAnalyticsService) getTopViolations(ctx context.Context, limit int) ([]ComplianceViolation, error) {
	if limit == 0 {
		limit = 10
	}

	query := `
		SELECT cv.id, cv.policy_id, cp.name as policy_name,
		       cv.resource_id, cv.resource_type, cv.severity,
		       cv.description, cv.detected_at, cv.resolved_at, cv.metadata
		FROM neuronip.compliance_violations cv
		LEFT JOIN neuronip.compliance_policies cp ON cp.id::text = cv.policy_id
		WHERE cv.resolved_at IS NULL
		ORDER BY 
			CASE cv.severity
				WHEN 'critical' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				ELSE 4
			END,
			cv.detected_at DESC
		LIMIT $1`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get violations: %w", err)
	}
	defer rows.Close()

	var violations []ComplianceViolation
	for rows.Next() {
		var violation ComplianceViolation
		var metadataJSON []byte
		var resolvedAt interface{}

		err := rows.Scan(&violation.ID, &violation.PolicyID, &violation.PolicyName,
			&violation.ResourceID, &violation.ResourceType, &violation.Severity,
			&violation.Description, &violation.DetectedAt, &resolvedAt, &metadataJSON)
		if err != nil {
			continue
		}

		if resolvedAt != nil {
			if t, ok := resolvedAt.(time.Time); ok {
				violation.ResolvedAt = &t
			}
		}
		json.Unmarshal(metadataJSON, &violation.Metadata)

		violations = append(violations, violation)
	}

	return violations, nil
}

/* getPolicyCompliance retrieves policy compliance status */
func (s *ComplianceAnalyticsService) getPolicyCompliance(ctx context.Context) ([]PolicyCompliance, error) {
	query := `
		SELECT 
			cp.id as policy_id,
			cp.name as policy_name,
			cp.last_checked,
			cp.status,
			COUNT(cv.id) as violation_count,
			CASE 
				WHEN COUNT(cv.id) = 0 THEN 100.0
				ELSE (1.0 - (COUNT(cv.id)::float / NULLIF(COUNT(*) FILTER (WHERE cv.detected_at > cp.last_checked - INTERVAL '30 days'), 0))) * 100.0
			END as compliance_rate
		FROM neuronip.compliance_policies cp
		LEFT JOIN neuronip.compliance_violations cv ON cv.policy_id = cp.id::text AND cv.resolved_at IS NULL
		WHERE cp.enabled = true
		GROUP BY cp.id, cp.name, cp.last_checked, cp.status
		ORDER BY violation_count DESC, cp.name ASC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy compliance: %w", err)
	}
	defer rows.Close()

	var policies []PolicyCompliance
	for rows.Next() {
		var policy PolicyCompliance

		err := rows.Scan(&policy.PolicyID, &policy.PolicyName, &policy.LastChecked,
			&policy.Status, &policy.ViolationCount, &policy.ComplianceRate)
		if err != nil {
			continue
		}

		policies = append(policies, policy)
	}

	return policies, nil
}

/* getRiskAreas identifies high-risk areas */
func (s *ComplianceAnalyticsService) getRiskAreas(ctx context.Context) ([]RiskArea, error) {
	// Analyze violations by resource type and area
	query := `
		SELECT 
			resource_type as area,
			COUNT(*) as violation_count,
			MAX(detected_at) as last_incident
		FROM neuronip.compliance_violations
		WHERE resolved_at IS NULL
		AND detected_at > NOW() - INTERVAL '90 days'
		GROUP BY resource_type
		HAVING COUNT(*) > 0
		ORDER BY violation_count DESC
		LIMIT 10`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk areas: %w", err)
	}
	defer rows.Close()

	var riskAreas []RiskArea
	for rows.Next() {
		var area RiskArea

		err := rows.Scan(&area.Area, &area.ViolationCount, &area.LastIncident)
		if err != nil {
			continue
		}

		// Calculate risk score (simplified)
		area.RiskScore = calculateRiskScore(area.ViolationCount, area.LastIncident)
		area.Description = fmt.Sprintf("High violation count (%d) in %s", area.ViolationCount, area.Area)

		riskAreas = append(riskAreas, area)
	}

	return riskAreas, nil
}

/* calculateRiskScore calculates risk score based on violations and recency */
func calculateRiskScore(violationCount int, lastIncident time.Time) float64 {
	// Base score from violation count
	countScore := float64(violationCount) * 0.1
	if countScore > 0.7 {
		countScore = 0.7
	}

	// Recency factor (more recent = higher risk)
	daysSince := time.Since(lastIncident).Hours() / 24
	recencyScore := 0.3 * (1.0 - (daysSince / 90.0)) // Decay over 90 days
	if recencyScore < 0 {
		recencyScore = 0
	}

	return countScore + recencyScore
}
