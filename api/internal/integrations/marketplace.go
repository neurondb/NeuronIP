package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* MarketplaceService provides integration marketplace functionality */
type MarketplaceService struct {
	pool *pgxpool.Pool
}

/* NewMarketplaceService creates a new marketplace service */
func NewMarketplaceService(pool *pgxpool.Pool) *MarketplaceService {
	return &MarketplaceService{pool: pool}
}

/* ListIntegrations lists available integrations */
func (ms *MarketplaceService) ListIntegrations(ctx context.Context, category string) ([]MarketplaceIntegration, error) {
	query := `
		SELECT id, integration_name, category, description, version, rating_average,
		       rating_count, install_count, status, created_at
		FROM neuronip.integration_marketplace
		WHERE status = 'published'`

	args := []interface{}{}
	if category != "" {
		query += " AND category = $1"
		args = append(args, category)
	}

	query += " ORDER BY rating_average DESC, install_count DESC"

	rows, err := ms.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}
	defer rows.Close()

	var integrations []MarketplaceIntegration
	for rows.Next() {
		var integration MarketplaceIntegration
		err := rows.Scan(
			&integration.ID, &integration.Name, &integration.Category, &integration.Description,
			&integration.Version, &integration.RatingAverage, &integration.RatingCount,
			&integration.InstallCount, &integration.Status, &integration.CreatedAt,
		)
		if err != nil {
			continue
		}
		integrations = append(integrations, integration)
	}

	return integrations, nil
}

/* InstallIntegration installs an integration */
func (ms *MarketplaceService) InstallIntegration(ctx context.Context, integrationID uuid.UUID, userID string, config map[string]interface{}) error {
	installationID := uuid.New()
	configJSON, _ := json.Marshal(config)

	query := `
		INSERT INTO neuronip.integration_installations 
		(id, integration_id, user_id, config, status, installed_at)
		VALUES ($1, $2, $3, $4, 'active', NOW())`

	_, err := ms.pool.Exec(ctx, query, installationID, integrationID, userID, configJSON)
	if err != nil {
		return fmt.Errorf("failed to install integration: %w", err)
	}

	// Update install count
	updateQuery := `
		UPDATE neuronip.integration_marketplace
		SET install_count = install_count + 1
		WHERE id = $1`

	ms.pool.Exec(ctx, updateQuery, integrationID)

	return nil
}

/* MarketplaceIntegration represents a marketplace integration */
type MarketplaceIntegration struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	Description   string    `json:"description"`
	Version       string    `json:"version"`
	RatingAverage float64   `json:"rating_average"`
	RatingCount   int       `json:"rating_count"`
	InstallCount  int       `json:"install_count"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
