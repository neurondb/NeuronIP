package collaboration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* CollaborationService provides collaboration features */
type CollaborationService struct {
	pool *pgxpool.Pool
}

/* NewCollaborationService creates a new collaboration service */
func NewCollaborationService(pool *pgxpool.Pool) *CollaborationService {
	return &CollaborationService{pool: pool}
}

/* SharedDashboard represents a shared dashboard */
type SharedDashboard struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	Description   *string                `json:"description,omitempty"`
	DashboardConfig map[string]interface{} `json:"dashboard_config"`
	CreatedBy     string                 `json:"created_by"`
	WorkspaceID   *uuid.UUID             `json:"workspace_id,omitempty"`
	IsPublic      bool                   `json:"is_public"`
	SharedWith    []string               `json:"shared_with,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* CreateSharedDashboard creates a shared dashboard */
func (s *CollaborationService) CreateSharedDashboard(ctx context.Context, name string, description *string, dashboardConfig map[string]interface{}, createdBy string, workspaceID *uuid.UUID, isPublic bool, sharedWith []string, tags []string) (*SharedDashboard, error) {
	id := uuid.New()
	configJSON, _ := json.Marshal(dashboardConfig)
	sharedWithJSON, _ := json.Marshal(sharedWith)
	tagsJSON, _ := json.Marshal(tags)

	query := `
		INSERT INTO neuronip.shared_dashboards 
		(id, name, description, dashboard_config, created_by, workspace_id, is_public, shared_with, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, name, description, dashboard_config, created_by, workspace_id, is_public, shared_with, tags, created_at, updated_at`

	var dashboard SharedDashboard
	var desc sql.NullString
	var sharedWithRaw, tagsRaw json.RawMessage

	err := s.pool.QueryRow(ctx, query, id, name, description, configJSON, createdBy, workspaceID, isPublic, sharedWithJSON, tagsJSON).Scan(
		&dashboard.ID, &dashboard.Name, &desc, &dashboard.DashboardConfig,
		&dashboard.CreatedBy, &dashboard.WorkspaceID, &dashboard.IsPublic,
		&sharedWithRaw, &tagsRaw, &dashboard.CreatedAt, &dashboard.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared dashboard: %w", err)
	}

	if desc.Valid {
		dashboard.Description = &desc.String
	}
	if sharedWithRaw != nil {
		json.Unmarshal(sharedWithRaw, &dashboard.SharedWith)
	}
	if tagsRaw != nil {
		json.Unmarshal(tagsRaw, &dashboard.Tags)
	}

	return &dashboard, nil
}

/* DashboardComment represents a comment on a dashboard */
type DashboardComment struct {
	ID              uuid.UUID  `json:"id"`
	DashboardID     uuid.UUID  `json:"dashboard_id"`
	UserID           string     `json:"user_id"`
	CommentText      string     `json:"comment_text"`
	ParentCommentID *uuid.UUID  `json:"parent_comment_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

/* AddDashboardComment adds a comment to a dashboard */
func (s *CollaborationService) AddDashboardComment(ctx context.Context, dashboardID uuid.UUID, userID string, commentText string, parentCommentID *uuid.UUID) (*DashboardComment, error) {
	id := uuid.New()

	query := `
		INSERT INTO neuronip.dashboard_comments 
		(id, dashboard_id, user_id, comment_text, parent_comment_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, dashboard_id, user_id, comment_text, parent_comment_id, created_at, updated_at`

	var comment DashboardComment
	var parentID sql.NullString

	err := s.pool.QueryRow(ctx, query, id, dashboardID, userID, commentText, parentCommentID).Scan(
		&comment.ID, &comment.DashboardID, &comment.UserID, &comment.CommentText,
		&parentID, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	if parentID.Valid {
		pID, _ := uuid.Parse(parentID.String)
		comment.ParentCommentID = &pID
	}

	return &comment, nil
}

/* AnswerCard represents a shared answer card */
type AnswerCard struct {
	ID          uuid.UUID              `json:"id"`
	Title       string                 `json:"title"`
	QueryText   string                 `json:"query_text"`
	QueryResult map[string]interface{} `json:"query_result"`
	Explanation *string                `json:"explanation,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	WorkspaceID *uuid.UUID             `json:"workspace_id,omitempty"`
	IsPublic    bool                   `json:"is_public"`
	SharedWith  []string               `json:"shared_with,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* CreateAnswerCard creates a shared answer card */
func (s *CollaborationService) CreateAnswerCard(ctx context.Context, title string, queryText string, queryResult map[string]interface{}, explanation *string, createdBy string, workspaceID *uuid.UUID, isPublic bool, sharedWith []string, tags []string) (*AnswerCard, error) {
	id := uuid.New()
	resultJSON, _ := json.Marshal(queryResult)
	sharedWithJSON, _ := json.Marshal(sharedWith)
	tagsJSON, _ := json.Marshal(tags)

	query := `
		INSERT INTO neuronip.answer_cards 
		(id, title, query_text, query_result, explanation, created_by, workspace_id, is_public, shared_with, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, title, query_text, query_result, explanation, created_by, workspace_id, is_public, shared_with, tags, created_at, updated_at`

	var card AnswerCard
	var expl sql.NullString
	var sharedWithRaw, tagsRaw json.RawMessage

	err := s.pool.QueryRow(ctx, query, id, title, queryText, resultJSON, explanation, createdBy, workspaceID, isPublic, sharedWithJSON, tagsJSON).Scan(
		&card.ID, &card.Title, &card.QueryText, &card.QueryResult,
		&expl, &card.CreatedBy, &card.WorkspaceID, &card.IsPublic,
		&sharedWithRaw, &tagsRaw, &card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create answer card: %w", err)
	}

	if expl.Valid {
		card.Explanation = &expl.String
	}
	if sharedWithRaw != nil {
		json.Unmarshal(sharedWithRaw, &card.SharedWith)
	}
	if tagsRaw != nil {
		json.Unmarshal(tagsRaw, &card.Tags)
	}

	return &card, nil
}

/* SavedQuestion represents a saved question */
type SavedQuestion struct {
	ID          uuid.UUID  `json:"id"`
	QuestionText string    `json:"question_text"`
	AnswerText   *string   `json:"answer_text,omitempty"`
	Explanation  *string   `json:"explanation,omitempty"`
	QueryUsed    *string   `json:"query_used,omitempty"`
	CreatedBy    string    `json:"created_by"`
	WorkspaceID  *uuid.UUID `json:"workspace_id,omitempty"`
	IsShared     bool      `json:"is_shared"`
	SharedWith   []string  `json:"shared_with,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

/* SaveQuestion saves a question */
func (s *CollaborationService) SaveQuestion(ctx context.Context, questionText string, answerText *string, explanation *string, queryUsed *string, createdBy string, workspaceID *uuid.UUID, isShared bool, sharedWith []string, tags []string) (*SavedQuestion, error) {
	id := uuid.New()
	sharedWithJSON, _ := json.Marshal(sharedWith)
	tagsJSON, _ := json.Marshal(tags)


	var saved SavedQuestion
	var answer, expl, queryUsedVal sql.NullString
	var sharedWithRaw, tagsRaw json.RawMessage

	querySQL := `
		INSERT INTO neuronip.saved_questions 
		(id, question_text, answer_text, explanation, query_used, created_by, workspace_id, is_shared, shared_with, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, question_text, answer_text, explanation, query_used, created_by, workspace_id, is_shared, shared_with, tags, created_at, updated_at`

	err := s.pool.QueryRow(ctx, querySQL, id, questionText, answerText, explanation, queryUsed, createdBy, workspaceID, isShared, sharedWithJSON, tagsJSON).Scan(
		&saved.ID, &saved.QuestionText, &answer, &expl, &queryUsedVal,
		&saved.CreatedBy, &saved.WorkspaceID, &saved.IsShared,
		&sharedWithRaw, &tagsRaw, &saved.CreatedAt, &saved.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save question: %w", err)
	}

	if answer.Valid {
		saved.AnswerText = &answer.String
	}
	if expl.Valid {
		saved.Explanation = &expl.String
	}
	if queryUsedVal.Valid {
		saved.QueryUsed = &queryUsedVal.String
	}
	if sharedWithRaw != nil {
		json.Unmarshal(sharedWithRaw, &saved.SharedWith)
	}
	if tagsRaw != nil {
		json.Unmarshal(tagsRaw, &saved.Tags)
	}

	return &saved, nil
}

/* GetSharedDashboards retrieves shared dashboards */
func (s *CollaborationService) GetSharedDashboards(ctx context.Context, workspaceID *uuid.UUID, userID *string, limit int) ([]SharedDashboard, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, name, description, dashboard_config, created_by, workspace_id, is_public, shared_with, tags, created_at, updated_at
		FROM neuronip.shared_dashboards
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if workspaceID != nil {
		query += fmt.Sprintf(" AND (workspace_id = $%d OR workspace_id IS NULL)", argIndex)
		args = append(args, *workspaceID)
		argIndex++
	}

	if userID != nil {
		query += fmt.Sprintf(" AND (created_by = $%d OR is_public = true OR $%d = ANY(shared_with))", argIndex, argIndex)
		args = append(args, *userID)
		argIndex++
	} else {
		query += " AND is_public = true"
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared dashboards: %w", err)
	}
	defer rows.Close()

	var dashboards []SharedDashboard
	for rows.Next() {
		var dashboard SharedDashboard
		var desc sql.NullString
		var configJSON, sharedWithRaw, tagsRaw json.RawMessage

		err := rows.Scan(
			&dashboard.ID, &dashboard.Name, &desc, &configJSON,
			&dashboard.CreatedBy, &dashboard.WorkspaceID, &dashboard.IsPublic,
			&sharedWithRaw, &tagsRaw, &dashboard.CreatedAt, &dashboard.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if desc.Valid {
			dashboard.Description = &desc.String
		}
		if configJSON != nil {
			json.Unmarshal(configJSON, &dashboard.DashboardConfig)
		}
		if sharedWithRaw != nil {
			json.Unmarshal(sharedWithRaw, &dashboard.SharedWith)
		}
		if tagsRaw != nil {
			json.Unmarshal(tagsRaw, &dashboard.Tags)
		}

		dashboards = append(dashboards, dashboard)
	}

	return dashboards, nil
}

/* GetDashboardComments retrieves comments for a dashboard */
func (s *CollaborationService) GetDashboardComments(ctx context.Context, dashboardID uuid.UUID) ([]DashboardComment, error) {
	query := `
		SELECT id, dashboard_id, user_id, comment_text, parent_comment_id, created_at, updated_at
		FROM neuronip.dashboard_comments
		WHERE dashboard_id = $1
		ORDER BY created_at ASC`

	rows, err := s.pool.Query(ctx, query, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []DashboardComment
	for rows.Next() {
		var comment DashboardComment
		var parentID sql.NullString

		err := rows.Scan(
			&comment.ID, &comment.DashboardID, &comment.UserID, &comment.CommentText,
			&parentID, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if parentID.Valid {
			pID, _ := uuid.Parse(parentID.String)
			comment.ParentCommentID = &pID
		}

		comments = append(comments, comment)
	}

	return comments, nil
}
