package partners

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* PartnerService provides partner management functionality */
type PartnerService struct {
	pool *pgxpool.Pool
}

/* NewPartnerService creates a new partner service */
func NewPartnerService(pool *pgxpool.Pool) *PartnerService {
	return &PartnerService{pool: pool}
}

/* RegisterPartner registers a new partner */
func (ps *PartnerService) RegisterPartner(ctx context.Context, partner Partner) (*Partner, error) {
	partner.ID = uuid.New()
	partner.CreatedAt = time.Now()
	partner.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(partner.Metadata)

	query := `
		INSERT INTO neuronip.partners 
		(id, partner_name, partner_type, company_name, contact_email, website_url, logo_url,
		 description, certification_level, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, partner_name, partner_type, company_name, contact_email, website_url, logo_url,
		          description, certification_level, status, metadata, created_at, updated_at`

	var metadataJSONRaw json.RawMessage
	err := ps.pool.QueryRow(ctx, query,
		partner.ID, partner.PartnerName, partner.PartnerType, partner.CompanyName,
		partner.ContactEmail, partner.WebsiteURL, partner.LogoURL, partner.Description,
		partner.CertificationLevel, partner.Status, metadataJSON, partner.CreatedAt, partner.UpdatedAt,
	).Scan(
		&partner.ID, &partner.PartnerName, &partner.PartnerType, &partner.CompanyName,
		&partner.ContactEmail, &partner.WebsiteURL, &partner.LogoURL, &partner.Description,
		&partner.CertificationLevel, &partner.Status, &metadataJSONRaw, &partner.CreatedAt, &partner.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register partner: %w", err)
	}

	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &partner.Metadata)
	}

	return &partner, nil
}

/* Partner represents a partner */
type Partner struct {
	ID                uuid.UUID              `json:"id"`
	PartnerName       string                 `json:"partner_name"`
	PartnerType       string                 `json:"partner_type"`
	CompanyName       string                 `json:"company_name"`
	ContactEmail      string                 `json:"contact_email"`
	WebsiteURL        *string                `json:"website_url,omitempty"`
	LogoURL           *string                `json:"logo_url,omitempty"`
	Description       *string                `json:"description,omitempty"`
	CertificationLevel *string               `json:"certification_level,omitempty"`
	Status            string                 `json:"status"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}
