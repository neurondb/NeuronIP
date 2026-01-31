package tenancy

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* IsolationService provides tenant resource isolation functionality */
type IsolationService struct {
	pool *pgxpool.Pool
}

/* NewIsolationService creates a new isolation service */
func NewIsolationService(pool *pgxpool.Pool) *IsolationService {
	return &IsolationService{pool: pool}
}

/* TenantResourceQuota represents a resource quota for a tenant */
type TenantResourceQuota struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	ResourceType string                 `json:"resource_type"`
	QuotaLimit   float64                `json:"quota_limit"`
	CurrentUsage float64                `json:"current_usage"`
	PeriodStart  time.Time              `json:"period_start"`
	PeriodEnd    *time.Time             `json:"period_end,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

/* CheckQuota checks if a tenant has available quota for a resource */
func (is *IsolationService) CheckQuota(ctx context.Context, tenantID, resourceType string, requestedAmount float64) (bool, error) {
	quota, err := is.GetQuota(ctx, tenantID, resourceType)
	if err != nil {
		return false, fmt.Errorf("failed to get quota: %w", err)
	}
	
	if quota == nil {
		// No quota set, allow
		return true, nil
	}
	
	available := quota.QuotaLimit - quota.CurrentUsage
	return available >= requestedAmount, nil
}

/* ReserveQuota reserves quota for a tenant */
func (is *IsolationService) ReserveQuota(ctx context.Context, tenantID, resourceType string, amount float64) error {
	quota, err := is.GetQuota(ctx, tenantID, resourceType)
	if err != nil {
		return fmt.Errorf("failed to get quota: %w", err)
	}
	
	if quota == nil {
		// Create unlimited quota (or set default)
		return nil
	}
	
	// Check if quota is available
	available := quota.QuotaLimit - quota.CurrentUsage
	if available < amount {
		return fmt.Errorf("insufficient quota: requested %.2f, available %.2f", amount, available)
	}
	
	// Update usage
	query := `
		UPDATE neuronip.tenant_resource_quotas
		SET current_usage = current_usage + $1, updated_at = NOW()
		WHERE tenant_id = $2 AND resource_type = $3
	`
	
	_, err = is.pool.Exec(ctx, query, amount, tenantID, resourceType)
	return err
}

/* ReleaseQuota releases reserved quota */
func (is *IsolationService) ReleaseQuota(ctx context.Context, tenantID, resourceType string, amount float64) error {
	query := `
		UPDATE neuronip.tenant_resource_quotas
		SET current_usage = GREATEST(0, current_usage - $1), updated_at = NOW()
		WHERE tenant_id = $2 AND resource_type = $3
	`
	
	_, err := is.pool.Exec(ctx, query, amount, tenantID, resourceType)
	return err
}

/* GetQuota retrieves quota for a tenant and resource type */
func (is *IsolationService) GetQuota(ctx context.Context, tenantID, resourceType string) (*TenantResourceQuota, error) {
	query := `
		SELECT id, tenant_id, resource_type, quota_limit, current_usage, period_start, period_end, created_at, updated_at
		FROM neuronip.tenant_resource_quotas
		WHERE tenant_id = $1 AND resource_type = $2
		ORDER BY period_start DESC
		LIMIT 1
	`
	
	var quota TenantResourceQuota
	var periodEnd *time.Time
	
	err := is.pool.QueryRow(ctx, query, tenantID, resourceType).Scan(
		&quota.ID, &quota.TenantID, &quota.ResourceType, &quota.QuotaLimit, &quota.CurrentUsage,
		&quota.PeriodStart, &periodEnd, &quota.CreatedAt, &quota.UpdatedAt,
	)
	if err != nil {
		// No quota found
		return nil, nil
	}
	
	quota.PeriodEnd = periodEnd
	return &quota, nil
}

/* SetQuota sets quota for a tenant */
func (is *IsolationService) SetQuota(ctx context.Context, tenantID, resourceType string, quotaLimit float64, periodStart time.Time, periodEnd *time.Time) error {
	quotaID := uuid.New()
	now := time.Now()
	
	query := `
		INSERT INTO neuronip.tenant_resource_quotas 
		(id, tenant_id, resource_type, quota_limit, current_usage, period_start, period_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 0, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, resource_type, period_start) 
		DO UPDATE SET 
			quota_limit = EXCLUDED.quota_limit,
			updated_at = EXCLUDED.updated_at
	`
	
	_, err := is.pool.Exec(ctx, query, quotaID, tenantID, resourceType, quotaLimit, periodStart, periodEnd, now, now)
	return err
}

/* GetTenantQuotas retrieves all quotas for a tenant */
func (is *IsolationService) GetTenantQuotas(ctx context.Context, tenantID string) ([]TenantResourceQuota, error) {
	query := `
		SELECT id, tenant_id, resource_type, quota_limit, current_usage, period_start, period_end, created_at, updated_at
		FROM neuronip.tenant_resource_quotas
		WHERE tenant_id = $1
		ORDER BY resource_type, period_start DESC
	`
	
	rows, err := is.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotas: %w", err)
	}
	defer rows.Close()
	
	var quotas []TenantResourceQuota
	for rows.Next() {
		var quota TenantResourceQuota
		var periodEnd *time.Time
		
		err := rows.Scan(
			&quota.ID, &quota.TenantID, &quota.ResourceType, &quota.QuotaLimit, &quota.CurrentUsage,
			&quota.PeriodStart, &periodEnd, &quota.CreatedAt, &quota.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		quota.PeriodEnd = periodEnd
		quotas = append(quotas, quota)
	}
	
	return quotas, nil
}

/* IsolateTenantData ensures data isolation for a tenant */
func (is *IsolationService) IsolateTenantData(ctx context.Context, tenantID string) context.Context {
	// Set tenant context for RLS policies
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	return ctx
}

/* GetTenantFromContext extracts tenant ID from context */
func (is *IsolationService) GetTenantFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	return tenantID, ok
}
