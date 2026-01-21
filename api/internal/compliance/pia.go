package compliance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* PIARequest represents a Privacy Impact Assessment request */
type PIARequest struct {
	ID               uuid.UUID              `json:"id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	ProjectName      string                 `json:"project_name"`
	DataTypes        []string               `json:"data_types"`
	DataSubjects     []string               `json:"data_subjects"`
	ProcessingPurposes []string             `json:"processing_purposes"`
	RiskLevel        string                 `json:"risk_level"` // "low", "medium", "high", "critical"
	Status           string                 `json:"status"` // "draft", "submitted", "review", "approved", "rejected"
	SubmittedBy      string                 `json:"submitted_by"`
	ReviewedBy       *string                `json:"reviewed_by,omitempty"`
	ApprovedBy       *string                `json:"approved_by,omitempty"`
	AssessmentResults map[string]interface{} `json:"assessment_results,omitempty"`
	Recommendations  []string               `json:"recommendations,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	SubmittedAt      *time.Time             `json:"submitted_at,omitempty"`
	ReviewedAt       *time.Time             `json:"reviewed_at,omitempty"`
	ApprovedAt       *time.Time             `json:"approved_at,omitempty"`
}

/* PIAService provides Privacy Impact Assessment functionality */
type PIAService struct {
	pool *pgxpool.Pool
}

/* NewPIAService creates a new PIA service */
func NewPIAService(pool *pgxpool.Pool) *PIAService {
	return &PIAService{pool: pool}
}

/* CreatePIARequest creates a new PIA request */
func (s *PIAService) CreatePIARequest(ctx context.Context, req PIARequest) (*PIARequest, error) {
	req.ID = uuid.New()
	req.Status = "draft"
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	dataTypesJSON, _ := json.Marshal(req.DataTypes)
	dataSubjectsJSON, _ := json.Marshal(req.DataSubjects)
	processingPurposesJSON, _ := json.Marshal(req.ProcessingPurposes)
	assessmentResultsJSON, _ := json.Marshal(req.AssessmentResults)
	recommendationsJSON, _ := json.Marshal(req.Recommendations)
	metadataJSON, _ := json.Marshal(req.Metadata)

	var reviewedBy, approvedBy sql.NullString

	query := `
		INSERT INTO neuronip.pia_requests
		(id, title, description, project_name, data_types, data_subjects, processing_purposes,
		 risk_level, status, submitted_by, reviewed_by, approved_by, assessment_results,
		 recommendations, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		req.ID, req.Title, req.Description, req.ProjectName, dataTypesJSON,
		dataSubjectsJSON, processingPurposesJSON, req.RiskLevel, req.Status,
		req.SubmittedBy, reviewedBy, approvedBy, assessmentResultsJSON,
		recommendationsJSON, metadataJSON, req.CreatedAt, req.UpdatedAt,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create PIA request: %w", err)
	}

	// Perform automated risk assessment
	s.performRiskAssessment(ctx, &req)

	return &req, nil
}

/* performRiskAssessment performs automated risk assessment */
func (s *PIAService) performRiskAssessment(ctx context.Context, req *PIARequest) {
	riskScore := 0

	// Check data types - sensitive data increases risk
	for _, dataType := range req.DataTypes {
		switch dataType {
		case "health", "financial", "biometric":
			riskScore += 3
		case "location", "behavioral":
			riskScore += 2
		default:
			riskScore += 1
		}
	}

	// Check data subjects - vulnerable populations increase risk
	for _, subject := range req.DataSubjects {
		switch subject {
		case "children", "vulnerable_adults":
			riskScore += 3
		case "employees", "customers":
			riskScore += 1
		}
	}

	// Determine risk level
	if riskScore >= 10 {
		req.RiskLevel = "critical"
	} else if riskScore >= 7 {
		req.RiskLevel = "high"
	} else if riskScore >= 4 {
		req.RiskLevel = "medium"
	} else {
		req.RiskLevel = "low"
	}

	// Generate recommendations based on risk
	recommendations := []string{}
	if req.RiskLevel == "high" || req.RiskLevel == "critical" {
		recommendations = append(recommendations, "Implement data minimization principles")
		recommendations = append(recommendations, "Require explicit consent from data subjects")
		recommendations = append(recommendations, "Implement additional security controls")
		recommendations = append(recommendations, "Conduct regular security audits")
	}
	if req.RiskLevel == "critical" {
		recommendations = append(recommendations, "Require DPO approval")
		recommendations = append(recommendations, "Implement data protection impact assessment")
	}

	req.Recommendations = recommendations

	// Update assessment results
	req.AssessmentResults = map[string]interface{}{
		"risk_score": riskScore,
		"risk_factors": map[string]interface{}{
			"data_types_count": len(req.DataTypes),
			"data_subjects_count": len(req.DataSubjects),
		},
	}

	// Update in database
	assessmentJSON, _ := json.Marshal(req.AssessmentResults)
	recommendationsJSON, _ := json.Marshal(req.Recommendations)

	updateQuery := `
		UPDATE neuronip.pia_requests
		SET risk_level = $1, assessment_results = $2, recommendations = $3, updated_at = NOW()
		WHERE id = $4`

	s.pool.Exec(ctx, updateQuery, req.RiskLevel, assessmentJSON, recommendationsJSON, req.ID)
}

/* SubmitPIARequest submits a PIA request for review */
func (s *PIAService) SubmitPIARequest(ctx context.Context, requestID uuid.UUID) error {
	submittedAt := time.Now()

	query := `
		UPDATE neuronip.pia_requests
		SET status = 'submitted', submitted_at = $1, updated_at = NOW()
		WHERE id = $2`

	_, err := s.pool.Exec(ctx, query, submittedAt, requestID)
	return err
}

/* ReviewPIARequest reviews a PIA request */
func (s *PIAService) ReviewPIARequest(ctx context.Context, requestID uuid.UUID, reviewerID string, approved bool) error {
	reviewedAt := time.Now()
	status := "approved"
	if !approved {
		status = "rejected"
	}

	var approvedBy sql.NullString
	var approvedAt sql.NullTime
	if approved {
		approvedBy = sql.NullString{String: reviewerID, Valid: true}
		approvedAt = sql.NullTime{Time: reviewedAt, Valid: true}
	}

	query := `
		UPDATE neuronip.pia_requests
		SET status = $1, reviewed_by = $2, reviewed_at = $3, approved_by = $4, approved_at = $5, updated_at = NOW()
		WHERE id = $6`

	_, err := s.pool.Exec(ctx, query, status, reviewerID, reviewedAt, approvedBy, approvedAt, requestID)
	return err
}

/* GetPIARequest gets a PIA request by ID */
func (s *PIAService) GetPIARequest(ctx context.Context, requestID uuid.UUID) (*PIARequest, error) {
	var req PIARequest
	var reviewedBy, approvedBy sql.NullString
	var submittedAt, reviewedAt, approvedAt sql.NullTime
	var dataTypesJSON, dataSubjectsJSON, processingPurposesJSON []byte
	var assessmentResultsJSON, recommendationsJSON, metadataJSON []byte

	query := `
		SELECT id, title, description, project_name, data_types, data_subjects, processing_purposes,
		       risk_level, status, submitted_by, reviewed_by, approved_by, assessment_results,
		       recommendations, metadata, created_at, updated_at, submitted_at, reviewed_at, approved_at
		FROM neuronip.pia_requests
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, requestID).Scan(
		&req.ID, &req.Title, &req.Description, &req.ProjectName, &dataTypesJSON,
		&dataSubjectsJSON, &processingPurposesJSON, &req.RiskLevel, &req.Status,
		&req.SubmittedBy, &reviewedBy, &approvedBy, &assessmentResultsJSON,
		&recommendationsJSON, &metadataJSON, &req.CreatedAt, &req.UpdatedAt,
		&submittedAt, &reviewedAt, &approvedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("PIA request not found: %w", err)
	}

	if reviewedBy.Valid {
		req.ReviewedBy = &reviewedBy.String
	}
	if approvedBy.Valid {
		req.ApprovedBy = &approvedBy.String
	}
	if submittedAt.Valid {
		req.SubmittedAt = &submittedAt.Time
	}
	if reviewedAt.Valid {
		req.ReviewedAt = &reviewedAt.Time
	}
	if approvedAt.Valid {
		req.ApprovedAt = &approvedAt.Time
	}

	json.Unmarshal(dataTypesJSON, &req.DataTypes)
	json.Unmarshal(dataSubjectsJSON, &req.DataSubjects)
	json.Unmarshal(processingPurposesJSON, &req.ProcessingPurposes)
	if assessmentResultsJSON != nil {
		json.Unmarshal(assessmentResultsJSON, &req.AssessmentResults)
	}
	if recommendationsJSON != nil {
		json.Unmarshal(recommendationsJSON, &req.Recommendations)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &req.Metadata)
	}

	return &req, nil
}
