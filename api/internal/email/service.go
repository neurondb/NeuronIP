package email

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides email sending functionality */
type Service struct {
	pool      *pgxpool.Pool
	smtpHost  string
	smtpPort  string
	smtpUser  string
	smtpPass  string
	fromEmail string
	fromName  string
}

/* Config holds email service configuration */
type Config struct {
	SMTPHost  string
	SMTPPort  string
	SMTPUser  string
	SMTPPass  string
	FromEmail string
	FromName  string
}

/* NewService creates a new email service */
func NewService(pool *pgxpool.Pool, cfg Config) *Service {
	return &Service{
		pool:      pool,
		smtpHost:  cfg.SMTPHost,
		smtpPort:  cfg.SMTPPort,
		smtpUser:  cfg.SMTPUser,
		smtpPass:  cfg.SMTPPass,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
	}
}

/* SendVerificationEmail sends an email verification email */
func (s *Service) SendVerificationEmail(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	// Generate verification token
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Store verification token
	expiresAt := time.Now().Add(24 * time.Hour)
	query := `INSERT INTO neuronip.email_verifications (user_id, token, expires_at) 
	          VALUES ($1, $2, $3)`
	_, err = s.pool.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to store verification token: %w", err)
	}

	// Send email (simplified - in production, use proper email templates)
	verificationURL := fmt.Sprintf("https://example.com/verify-email?token=%s", token)
	subject := "Verify your email address"
	body := fmt.Sprintf("Please click the link to verify your email: %s", verificationURL)

	if err := s.sendEmail(email, subject, body); err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return token, nil
}

/* SendPasswordResetEmail sends a password reset email */
func (s *Service) SendPasswordResetEmail(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	// Generate reset token
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Store reset token
	expiresAt := time.Now().Add(1 * time.Hour)
	query := `INSERT INTO neuronip.password_resets (user_id, token, expires_at) 
	          VALUES ($1, $2, $3)`
	_, err = s.pool.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	// Send email
	resetURL := fmt.Sprintf("https://example.com/reset-password?token=%s", token)
	subject := "Reset your password"
	body := fmt.Sprintf("Click the link to reset your password: %s\n\nThis link expires in 1 hour.", resetURL)

	if err := s.sendEmail(email, subject, body); err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return token, nil
}

/* SendNotificationEmail sends a notification email */
func (s *Service) SendNotificationEmail(email, subject, message string) error {
	return s.sendEmail(email, subject, message)
}

/* sendEmail sends an email via SMTP */
func (s *Service) sendEmail(to, subject, body string) error {
	if s.smtpHost == "" {
		// Email not configured, skip sending (for development)
		return nil
	}

	msg := []byte(fmt.Sprintf("From: %s <%s>\r\n", s.fromName, s.fromEmail) +
		fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"\r\n" +
		body + "\r\n")

	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)

	return smtp.SendMail(addr, auth, s.fromEmail, []string{to}, msg)
}

/* generateToken generates a secure random token */
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

/* VerifyEmailToken verifies an email verification token */
func (s *Service) VerifyEmailToken(ctx context.Context, token string) (uuid.UUID, error) {
	var userID uuid.UUID
	var expiresAt time.Time

	query := `SELECT user_id, expires_at FROM neuronip.email_verifications WHERE token = $1`
	err := s.pool.QueryRow(ctx, query, token).Scan(&userID, &expiresAt)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid verification token")
	}

	if time.Now().After(expiresAt) {
		return uuid.Nil, fmt.Errorf("verification token expired")
	}

	// Mark email as verified
	updateQuery := `UPDATE neuronip.users SET email_verified = true WHERE id = $1`
	_, err = s.pool.Exec(ctx, updateQuery, userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to verify email: %w", err)
	}

	// Delete verification token
	deleteQuery := `DELETE FROM neuronip.email_verifications WHERE token = $1`
	s.pool.Exec(ctx, deleteQuery, token)

	return userID, nil
}

/* VerifyPasswordResetToken verifies a password reset token */
func (s *Service) VerifyPasswordResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	var userID uuid.UUID
	var expiresAt time.Time
	var used bool

	query := `SELECT user_id, expires_at, used FROM neuronip.password_resets WHERE token = $1`
	err := s.pool.QueryRow(ctx, query, token).Scan(&userID, &expiresAt, &used)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid reset token")
	}

	if used {
		return uuid.Nil, fmt.Errorf("reset token already used")
	}

	if time.Now().After(expiresAt) {
		return uuid.Nil, fmt.Errorf("reset token expired")
	}

	return userID, nil
}

/* MarkPasswordResetTokenUsed marks a password reset token as used */
func (s *Service) MarkPasswordResetTokenUsed(ctx context.Context, token string) error {
	query := `UPDATE neuronip.password_resets SET used = true WHERE token = $1`
	_, err := s.pool.Exec(ctx, query, token)
	return err
}
