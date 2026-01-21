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

/* ConsentRecord represents a consent record */
type ConsentRecord struct {
	ID              uuid.UUID              `json:"id"`
	SubjectID       string                 `json:"subject_id"`
	SubjectEmail    string                 `json:"subject_email"`
	ConsentType     string                 `json:"consent_type"` // "marketing", "analytics", "data_sharing", "processing"
	Purpose         string                 `json:"purpose"`
	Consented       bool                   `json:"consented"`
	ConsentedAt     *time.Time             `json:"consented_at,omitempty"`
	WithdrawnAt     *time.Time             `json:"withdrawn_at,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	Version         int                    `json:"version"`
	ConsentMethod   string                 `json:"consent_method"` // "explicit", "implicit", "opt_in", "opt_out"
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

/* ConsentService provides consent management functionality */
type ConsentService struct {
	pool *pgxpool.Pool
}

/* NewConsentService creates a new consent service */
func NewConsentService(pool *pgxpool.Pool) *ConsentService {
	return &ConsentService{pool: pool}
}

/* RecordConsent records a consent */
func (s *ConsentService) RecordConsent(ctx context.Context, consent ConsentRecord) (*ConsentRecord, error) {
	consent.ID = uuid.New()
	consent.CreatedAt = time.Now()
	consent.UpdatedAt = time.Now()

	if consent.Consented {
		now := time.Now()
		consent.ConsentedAt = &now
	}

	metadataJSON, _ := json.Marshal(consent.Metadata)

	var consentedAt, withdrawnAt, expiresAt sql.NullTime
	if consent.ConsentedAt != nil {
		consentedAt = sql.NullTime{Time: *consent.ConsentedAt, Valid: true}
	}
	if consent.WithdrawnAt != nil {
		withdrawnAt = sql.NullTime{Time: *consent.WithdrawnAt, Valid: true}
	}
	if consent.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *consent.ExpiresAt, Valid: true}
	}

	query := `
		INSERT INTO neuronip.consent_records
		(id, subject_id, subject_email, consent_type, purpose, consented, consented_at,
		 withdrawn_at, expires_at, version, consent_method, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		consent.ID, consent.SubjectID, consent.SubjectEmail, consent.ConsentType,
		consent.Purpose, consent.Consented, consentedAt, withdrawnAt, expiresAt,
		consent.Version, consent.ConsentMethod, metadataJSON, consent.CreatedAt, consent.UpdatedAt,
	).Scan(&consent.ID, &consent.CreatedAt, &consent.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to record consent: %w", err)
	}

	return &consent, nil
}

/* WithdrawConsent withdraws a consent */
func (s *ConsentService) WithdrawConsent(ctx context.Context, subjectID string, consentType string, purpose string) error {
	withdrawnAt := time.Now()

	query := `
		UPDATE neuronip.consent_records
		SET consented = false, withdrawn_at = $1, updated_at = NOW()
		WHERE subject_id = $2 AND consent_type = $3 AND purpose = $4
		  AND consented = true AND (expires_at IS NULL OR expires_at > NOW())`

	result, err := s.pool.Exec(ctx, query, withdrawnAt, subjectID, consentType, purpose)
	if err != nil {
		return fmt.Errorf("failed to withdraw consent: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no active consent found to withdraw")
	}

	return nil
}

/* CheckConsent checks if a subject has consented */
func (s *ConsentService) CheckConsent(ctx context.Context, subjectID string, consentType string, purpose string) (bool, error) {
	var consented bool

	query := `
		SELECT consented
		FROM neuronip.consent_records
		WHERE subject_id = $1 AND consent_type = $2 AND purpose = $3
		  AND (expires_at IS NULL OR expires_at > NOW())
		  AND withdrawn_at IS NULL
		ORDER BY version DESC, created_at DESC
		LIMIT 1`

	err := s.pool.QueryRow(ctx, query, subjectID, consentType, purpose).Scan(&consented)
	if err == sql.ErrNoRows {
		return false, nil // No consent record means no consent
	}
	if err != nil {
		return false, fmt.Errorf("failed to check consent: %w", err)
	}

	return consented, nil
}

/* GetSubjectConsents gets all consents for a subject */
func (s *ConsentService) GetSubjectConsents(ctx context.Context, subjectID string) ([]ConsentRecord, error) {
	query := `
		SELECT id, subject_id, subject_email, consent_type, purpose, consented,
		       consented_at, withdrawn_at, expires_at, version, consent_method, metadata,
		       created_at, updated_at
		FROM neuronip.consent_records
		WHERE subject_id = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, subjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get consents: %w", err)
	}
	defer rows.Close()

	consents := []ConsentRecord{}
	for rows.Next() {
		var consent ConsentRecord
		var consentedAt, withdrawnAt, expiresAt sql.NullTime
		var metadataJSON []byte

		err := rows.Scan(
			&consent.ID, &consent.SubjectID, &consent.SubjectEmail, &consent.ConsentType,
			&consent.Purpose, &consent.Consented, &consentedAt, &withdrawnAt, &expiresAt,
			&consent.Version, &consent.ConsentMethod, &metadataJSON,
			&consent.CreatedAt, &consent.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if consentedAt.Valid {
			consent.ConsentedAt = &consentedAt.Time
		}
		if withdrawnAt.Valid {
			consent.WithdrawnAt = &withdrawnAt.Time
		}
		if expiresAt.Valid {
			consent.ExpiresAt = &expiresAt.Time
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &consent.Metadata)
		}

		consents = append(consents, consent)
	}

	return consents, nil
}

/* CleanupExpiredConsents removes expired consents */
func (s *ConsentService) CleanupExpiredConsents(ctx context.Context) error {
	query := `
		UPDATE neuronip.consent_records
		SET consented = false, updated_at = NOW()
		WHERE expires_at < NOW() AND consented = true AND withdrawn_at IS NULL`

	_, err := s.pool.Exec(ctx, query)
	return err
}
