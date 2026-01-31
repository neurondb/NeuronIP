package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* BudgetService provides budget management functionality */
type BudgetService struct {
	pool *pgxpool.Pool
}

/* NewBudgetService creates a new budget service */
func NewBudgetService(pool *pgxpool.Pool) *BudgetService {
	return &BudgetService{pool: pool}
}

/* CreateBudget creates a budget */
func (bs *BudgetService) CreateBudget(ctx context.Context, budget Budget) (*Budget, error) {
	budget.ID = uuid.New()
	budget.CreatedAt = time.Now()
	budget.UpdatedAt = time.Now()

	allocationJSON, _ := json.Marshal(budget.Allocation)

	query := `
		INSERT INTO neuronip.budgets 
		(id, budget_name, user_id, period_start, period_end, limit_amount, currency,
		 allocation, alerts_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, budget_name, user_id, period_start, period_end, limit_amount, currency,
		          allocation, alerts_enabled, created_at, updated_at`

	var allocationJSONRaw json.RawMessage
	err := bs.pool.QueryRow(ctx, query,
		budget.ID, budget.Name, budget.UserID, budget.PeriodStart, budget.PeriodEnd,
		budget.LimitAmount, budget.Currency, allocationJSON, budget.AlertsEnabled,
		budget.CreatedAt, budget.UpdatedAt,
	).Scan(
		&budget.ID, &budget.Name, &budget.UserID, &budget.PeriodStart, &budget.PeriodEnd,
		&budget.LimitAmount, &budget.Currency, &allocationJSONRaw, &budget.AlertsEnabled,
		&budget.CreatedAt, &budget.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create budget: %w", err)
	}

	if allocationJSONRaw != nil {
		json.Unmarshal(allocationJSONRaw, &budget.Allocation)
	}

	return &budget, nil
}

/* CheckBudgetStatus checks if budget is exceeded */
func (bs *BudgetService) CheckBudgetStatus(ctx context.Context, budgetID uuid.UUID) (*BudgetStatus, error) {
	// Get budget
	budget, err := bs.GetBudget(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Calculate current spending
	query := `
		SELECT COALESCE(SUM(cost_amount), 0)
		FROM neuronip.cost_tracking
		WHERE user_id = $1
			AND billing_period_start >= $2
			AND billing_period_end <= $3`

	var currentSpending float64
	err = bs.pool.QueryRow(ctx, query, budget.UserID, budget.PeriodStart, budget.PeriodEnd).Scan(&currentSpending)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate spending: %w", err)
	}

	status := &BudgetStatus{
		BudgetID:       budgetID,
		LimitAmount:    budget.LimitAmount,
		CurrentSpending: currentSpending,
		Remaining:      budget.LimitAmount - currentSpending,
		PercentageUsed: (currentSpending / budget.LimitAmount) * 100,
		IsExceeded:     currentSpending >= budget.LimitAmount,
	}

	// Check alert thresholds
	if budget.AlertsEnabled {
		if status.PercentageUsed >= 90 {
			status.AlertLevel = "critical"
		} else if status.PercentageUsed >= 75 {
			status.AlertLevel = "warning"
		} else if status.PercentageUsed >= 50 {
			status.AlertLevel = "info"
		}
	}

	return status, nil
}

/* GetBudget retrieves a budget */
func (bs *BudgetService) GetBudget(ctx context.Context, budgetID uuid.UUID) (*Budget, error) {
	query := `
		SELECT id, budget_name, user_id, period_start, period_end, limit_amount, currency,
		       allocation, alerts_enabled, created_at, updated_at
		FROM neuronip.budgets
		WHERE id = $1`

	var budget Budget
	var allocationJSONRaw json.RawMessage

	err := bs.pool.QueryRow(ctx, query, budgetID).Scan(
		&budget.ID, &budget.Name, &budget.UserID, &budget.PeriodStart, &budget.PeriodEnd,
		&budget.LimitAmount, &budget.Currency, &allocationJSONRaw, &budget.AlertsEnabled,
		&budget.CreatedAt, &budget.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("budget not found: %w", err)
	}

	if allocationJSONRaw != nil {
		json.Unmarshal(allocationJSONRaw, &budget.Allocation)
	}

	return &budget, nil
}

/* Budget represents a budget */
type Budget struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	UserID        *string                `json:"user_id,omitempty"`
	PeriodStart   time.Time              `json:"period_start"`
	PeriodEnd     time.Time              `json:"period_end"`
	LimitAmount   float64                `json:"limit_amount"`
	Currency      string                 `json:"currency"`
	Allocation    map[string]float64     `json:"allocation,omitempty"` // By team/project
	AlertsEnabled bool                   `json:"alerts_enabled"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* BudgetStatus represents budget status */
type BudgetStatus struct {
	BudgetID        uuid.UUID `json:"budget_id"`
	LimitAmount     float64   `json:"limit_amount"`
	CurrentSpending float64   `json:"current_spending"`
	Remaining       float64   `json:"remaining"`
	PercentageUsed  float64   `json:"percentage_used"`
	IsExceeded      bool      `json:"is_exceeded"`
	AlertLevel      string    `json:"alert_level,omitempty"` // "info", "warning", "critical"
}
