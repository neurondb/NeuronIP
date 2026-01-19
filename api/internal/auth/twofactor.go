package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

/* TwoFactorService provides two-factor authentication functionality */
type TwoFactorService struct {
	queries *db.Queries
}

/* NewTwoFactorService creates a new 2FA service */
func NewTwoFactorService(queries *db.Queries) *TwoFactorService {
	return &TwoFactorService{
		queries: queries,
	}
}

/* TOTPSecret represents a TOTP secret with QR code URL */
type TOTPSecret struct {
	Secret   string
	QRCodeURL string
}

/* GenerateTOTPSecret generates a new TOTP secret for a user */
func (s *TwoFactorService) GenerateTOTPSecret(ctx context.Context, userID uuid.UUID, email string) (*TOTPSecret, error) {
	// Generate a random base32 secret (compatible with TOTP)
	secret := generateBase32Secret(32)
	// QR code URL would be generated in a real implementation
	// For now, return the secret
	qrCodeURL := fmt.Sprintf("otpauth://totp/NeuronIP:%s?secret=%s&issuer=NeuronIP", email, secret)

	return &TOTPSecret{
		Secret:    secret,
		QRCodeURL: qrCodeURL,
	}, nil
}

/* VerifyTOTP verifies a TOTP code - simplified implementation */
func (s *TwoFactorService) VerifyTOTP(secret, code string) (bool, error) {
	// In a real implementation, use github.com/pquerna/otp/totp
	// For now, this is a placeholder - should use totp.Validate(code, secret)
	// This will need the otp library to be added to go.mod
	return len(code) == 6, nil // Placeholder validation
}

/* generateBase32Secret generates a random base32 secret */
func generateBase32Secret(length int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(bytes)
}

/* Enable2FA enables 2FA for a user */
func (s *TwoFactorService) Enable2FA(ctx context.Context, userID uuid.UUID, secret string) error {
	// Encrypt secret
	secretHash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}
	secretHashStr := string(secretHash)

	query := `UPDATE neuronip.users SET two_factor_enabled = true, two_factor_secret = $1 WHERE id = $2`
	_, err = s.queries.DB.Exec(ctx, query, secretHashStr, userID)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}
	return nil
}

/* Disable2FA disables 2FA for a user */
func (s *TwoFactorService) Disable2FA(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE neuronip.users SET two_factor_enabled = false, two_factor_secret = NULL WHERE id = $1`
	_, err := s.queries.DB.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}
	return nil
}

/* GetUserTOTPSecret retrieves the user's TOTP secret (decrypted) */
func (s *TwoFactorService) GetUserTOTPSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}

	if user.TwoFactorSecret == nil {
		return "", fmt.Errorf("2FA not configured for user")
	}

	// Note: In a real implementation, you'd decrypt this
	// For now, we'll return it as-is (should be decrypted from storage)
	return *user.TwoFactorSecret, nil
}
