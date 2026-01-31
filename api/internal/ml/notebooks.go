package ml

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Notebook represents a notebook */
type Notebook struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	WorkspaceID     *uuid.UUID `json:"workspace_id,omitempty"`
	OwnerID         string     `json:"owner_id"`
	DefaultLanguage string     `json:"default_language"`
	WorkflowID      *uuid.UUID `json:"workflow_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

/* NotebookCell represents a notebook cell */
type NotebookCell struct {
	ID           uuid.UUID  `json:"id"`
	NotebookID   uuid.UUID  `json:"notebook_id"`
	Position     int        `json:"position"`
	CellType     string     `json:"cell_type"`
	Content      string     `json:"content"`
	Output       *string    `json:"output,omitempty"`
	RunID        *uuid.UUID `json:"run_id,omitempty"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

/* NotebookRun represents a notebook execution run */
type NotebookRun struct {
	ID                  uuid.UUID  `json:"id"`
	NotebookID          uuid.UUID  `json:"notebook_id"`
	WorkflowExecutionID *uuid.UUID `json:"workflow_execution_id,omitempty"`
	TriggeredBy         string     `json:"triggered_by"`
	Status              string     `json:"status"`
	StartedAt           time.Time  `json:"started_at"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	ErrorMessage        *string    `json:"error_message,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

/* NotebookService provides notebook CRUD and run management */
type NotebookService struct {
	pool *pgxpool.Pool
}

/* NewNotebookService creates a new notebook service */
func NewNotebookService(pool *pgxpool.Pool) *NotebookService {
	return &NotebookService{pool: pool}
}

/* CreateNotebook creates a notebook */
func (s *NotebookService) CreateNotebook(ctx context.Context, name, description, ownerID string, workspaceID, workflowID *uuid.UUID, defaultLanguage string) (*Notebook, error) {
	if defaultLanguage == "" {
		defaultLanguage = "sql"
	}
	id := uuid.New()
	now := time.Now()
	query := `
		INSERT INTO neuronip.notebooks (id, name, description, workspace_id, owner_id, default_language, workflow_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, description, workspace_id, owner_id, default_language, workflow_id, created_at, updated_at
	`
	var n Notebook
	var desc *string
	err := s.pool.QueryRow(ctx, query, id, name, nullStr(description), workspaceID, ownerID, defaultLanguage, workflowID, now, now).Scan(
		&n.ID, &n.Name, &desc, &n.WorkspaceID, &n.OwnerID, &n.DefaultLanguage, &n.WorkflowID, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create notebook: %w", err)
	}
	n.Description = desc
	return &n, nil
}

/* GetNotebook returns a notebook by ID */
func (s *NotebookService) GetNotebook(ctx context.Context, id uuid.UUID) (*Notebook, error) {
	query := `SELECT id, name, description, workspace_id, owner_id, default_language, workflow_id, created_at, updated_at FROM neuronip.notebooks WHERE id = $1`
	var n Notebook
	var desc *string
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.Name, &desc, &n.WorkspaceID, &n.OwnerID, &n.DefaultLanguage, &n.WorkflowID, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get notebook: %w", err)
	}
	n.Description = desc
	return &n, nil
}

/* ListNotebooks lists notebooks for workspace or owner */
func (s *NotebookService) ListNotebooks(ctx context.Context, workspaceID *uuid.UUID, ownerID string) ([]Notebook, error) {
	query := `SELECT id, name, description, workspace_id, owner_id, default_language, workflow_id, created_at, updated_at FROM neuronip.notebooks WHERE 1=1`
	args := []interface{}{}
	argN := 1
	if workspaceID != nil {
		query += fmt.Sprintf(" AND workspace_id = $%d", argN)
		args = append(args, workspaceID)
		argN++
	}
	if ownerID != "" {
		query += fmt.Sprintf(" AND owner_id = $%d", argN)
		args = append(args, ownerID)
		argN++
	}
	query += " ORDER BY updated_at DESC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Notebook
	for rows.Next() {
		var n Notebook
		var desc *string
		if err := rows.Scan(&n.ID, &n.Name, &desc, &n.WorkspaceID, &n.OwnerID, &n.DefaultLanguage, &n.WorkflowID, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		n.Description = desc
		list = append(list, n)
	}
	return list, rows.Err()
}

/* AddCell adds a cell to a notebook */
func (s *NotebookService) AddCell(ctx context.Context, notebookID uuid.UUID, position int, cellType, content string) (*NotebookCell, error) {
	if cellType == "" {
		cellType = "sql"
	}
	id := uuid.New()
	now := time.Now()
	query := `
		INSERT INTO neuronip.notebook_cells (id, notebook_id, position, cell_type, content, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7)
		RETURNING id, notebook_id, position, cell_type, content, output, run_id, status, started_at, completed_at, error_message, created_at, updated_at
	`
	var c NotebookCell
	var output, errMsg *string
	var runID *uuid.UUID
	var startedAt, completedAt *time.Time
	err := s.pool.QueryRow(ctx, query, id, notebookID, position, cellType, content, now, now).Scan(
		&c.ID, &c.NotebookID, &c.Position, &c.CellType, &c.Content, &output, &runID, &c.Status, &startedAt, &completedAt, &errMsg, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add cell: %w", err)
	}
	c.Output = output
	c.RunID = runID
	c.StartedAt = startedAt
	c.CompletedAt = completedAt
	c.ErrorMessage = errMsg
	return &c, nil
}

/* ListCells returns cells for a notebook */
func (s *NotebookService) ListCells(ctx context.Context, notebookID uuid.UUID) ([]NotebookCell, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, notebook_id, position, cell_type, content, output, run_id, status, started_at, completed_at, error_message, created_at, updated_at
		FROM neuronip.notebook_cells WHERE notebook_id = $1 ORDER BY position
	`, notebookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []NotebookCell
	for rows.Next() {
		var c NotebookCell
		var output, errMsg *string
		var runID *uuid.UUID
		var startedAt, completedAt *time.Time
		if err := rows.Scan(&c.ID, &c.NotebookID, &c.Position, &c.CellType, &c.Content, &output, &runID, &c.Status, &startedAt, &completedAt, &errMsg, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.Output = output
		c.RunID = runID
		c.StartedAt = startedAt
		c.CompletedAt = completedAt
		c.ErrorMessage = errMsg
		list = append(list, c)
	}
	return list, rows.Err()
}

/* CreateRun creates a notebook run (execution record; actual execution can be via workflow) */
func (s *NotebookService) CreateRun(ctx context.Context, notebookID uuid.UUID, triggeredBy string, workflowExecutionID *uuid.UUID) (*NotebookRun, error) {
	id := uuid.New()
	now := time.Now()
	query := `
		INSERT INTO neuronip.notebook_runs (id, notebook_id, workflow_execution_id, triggered_by, status, started_at, created_at)
		VALUES ($1, $2, $3, $4, 'running', $5, $6)
		RETURNING id, notebook_id, workflow_execution_id, triggered_by, status, started_at, completed_at, error_message, created_at
	`
	var r NotebookRun
	var completedAt *time.Time
	var errMsg *string
	err := s.pool.QueryRow(ctx, query, id, notebookID, workflowExecutionID, triggeredBy, now, now).Scan(
		&r.ID, &r.NotebookID, &r.WorkflowExecutionID, &r.TriggeredBy, &r.Status, &r.StartedAt, &completedAt, &errMsg, &r.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create run: %w", err)
	}
	r.CompletedAt = completedAt
	r.ErrorMessage = errMsg
	return &r, nil
}

/* ListRuns returns runs for a notebook */
func (s *NotebookService) ListRuns(ctx context.Context, notebookID uuid.UUID) ([]NotebookRun, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, notebook_id, workflow_execution_id, triggered_by, status, started_at, completed_at, error_message, created_at
		FROM neuronip.notebook_runs WHERE notebook_id = $1 ORDER BY started_at DESC
	`, notebookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []NotebookRun
	for rows.Next() {
		var r NotebookRun
		var completedAt *time.Time
		var errMsg *string
		if err := rows.Scan(&r.ID, &r.NotebookID, &r.WorkflowExecutionID, &r.TriggeredBy, &r.Status, &r.StartedAt, &completedAt, &errMsg, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.CompletedAt = completedAt
		r.ErrorMessage = errMsg
		list = append(list, r)
	}
	return list, rows.Err()
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
