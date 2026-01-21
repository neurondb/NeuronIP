package integrations

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* EmailService provides email integration functionality */
type EmailService struct {
	pool *pgxpool.Pool
}

/* NewEmailService creates a new email integration service */
func NewEmailService(pool *pgxpool.Pool) *EmailService {
	return &EmailService{pool: pool}
}

/* EmailConfig represents email integration configuration */
type EmailConfig struct {
	SMTPHost     string                 `json:"smtp_host"`
	SMTPPort     int                    `json:"smtp_port"`
	SMTPUsername string                 `json:"smtp_username"`
	SMTPPassword string                 `json:"smtp_password"`
	FromEmail    string                 `json:"from_email"`
	FromName     string                 `json:"from_name,omitempty"`
	UseTLS       bool                   `json:"use_tls"`
	UseSSL       bool                   `json:"use_ssl"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

/* SendEmail sends an email */
func (s *EmailService) SendEmail(ctx context.Context, config EmailConfig, to []string, subject string, body string, isHTML bool) error {
	if config.SMTPHost == "" {
		return fmt.Errorf("smtp_host is required")
	}
	if config.SMTPPort == 0 {
		config.SMTPPort = 587
	}
	if config.FromEmail == "" {
		return fmt.Errorf("from_email is required")
	}
	if len(to) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	// Build email message
	from := config.FromEmail
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail)
	}

	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", strings.Join(to, ", "))
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	if isHTML {
		message += "MIME-Version: 1.0\r\n"
		message += "Content-Type: text/html; charset=UTF-8\r\n"
	} else {
		message += "Content-Type: text/plain; charset=UTF-8\r\n"
	}
	message += "\r\n" + body

	// Setup authentication
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
	
	if config.UseSSL {
		// Use TLS connection
		tlsConfig := &tls.Config{
			ServerName: config.SMTPHost,
		}
		
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer conn.Close()
		
		client, err := smtp.NewClient(conn, config.SMTPHost)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		defer client.Close()
		
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		
		if err := client.Mail(config.FromEmail); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}
		
		for _, recipient := range to {
			if err := client.Rcpt(recipient); err != nil {
				return fmt.Errorf("failed to set recipient: %w", err)
			}
		}
		
		writer, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %w", err)
		}
		
		if _, err := writer.Write([]byte(message)); err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}
		
		if err := writer.Close(); err != nil {
			return fmt.Errorf("failed to close writer: %w", err)
		}
	} else {
		// Standard SMTP
		err := smtp.SendMail(addr, auth, config.FromEmail, to, []byte(message))
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}

	return nil
}

/* SendHTMLEmail sends an HTML email */
func (s *EmailService) SendHTMLEmail(ctx context.Context, config EmailConfig, to []string, subject string, htmlBody string) error {
	return s.SendEmail(ctx, config, to, subject, htmlBody, true)
}

/* SendPlainEmail sends a plain text email */
func (s *EmailService) SendPlainEmail(ctx context.Context, config EmailConfig, to []string, subject string, textBody string) error {
	return s.SendEmail(ctx, config, to, subject, textBody, false)
}

/* TestEmailConnection tests email connection */
func (s *EmailService) TestEmailConnection(ctx context.Context, config EmailConfig) error {
	return s.SendPlainEmail(ctx, config, []string{config.FromEmail}, "Test Email from NeuronIP", "This is a test email from NeuronIP integration.")
}
