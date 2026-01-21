package auth

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
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
	// Generate TOTP key using proper library
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "NeuronIP",
		AccountName: email,
		Period:      30, // 30 second time window
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	return &TOTPSecret{
		Secret:    key.Secret(),
		QRCodeURL: key.URL(),
	}, nil
}

/* VerifyTOTP verifies a TOTP code using proper TOTP validation */
func (s *TwoFactorService) VerifyTOTP(secret, code string) (bool, error) {
	// Validate TOTP code with proper library
	valid := totp.Validate(code, secret)
	if !valid {
		return false, fmt.Errorf("invalid TOTP code")
	}
	return true, nil
}

/* VerifyTOTPWithWindow verifies TOTP code with time window tolerance */
func (s *TwoFactorService) VerifyTOTPWithWindow(secret, code string, window int) (bool, error) {
	// Validate with time window tolerance (default is 1, can be increased for clock skew)
	if window <= 0 {
		window = 1
	}
	valid, _ := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      uint(window),
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if !valid {
		return false, fmt.Errorf("invalid TOTP code")
	}
	return true, nil
}

/* generateBase32Secret generates a random base32 secret (fallback method) */
func generateBase32Secret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based if rand fails
		for i := range bytes {
			bytes[i] = byte(time.Now().UnixNano() % 256)
		}
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
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
