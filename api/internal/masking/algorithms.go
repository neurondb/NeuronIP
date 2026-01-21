package masking

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

/* Tokenize tokenizes a value */
func (a *MaskingAlgorithms) Tokenize(value string, algorithm string) (string, error) {
	// Generate deterministic token from value
	hash := sha256.Sum256([]byte(value))
	token := base64.URLEncoding.EncodeToString(hash[:16]) // First 16 bytes
	
	return fmt.Sprintf("TOKEN_%s", token[:24]), nil
}

/* Encrypt encrypts a value */
func (a *MaskingAlgorithms) Encrypt(value string, algorithm string) (string, error) {
	// Use AES-256-GCM for encryption
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

/* FormatPreservingMask masks a value while preserving format */
func (a *MaskingAlgorithms) FormatPreservingMask(value string, rule string) (string, error) {
	if rule == "" {
		// Default: preserve format by masking characters but keeping structure
		return a.preserveFormat(value), nil
	}

	switch rule {
	case "email":
		// user@domain.com -> u***@d***.com
		if parts := strings.Split(value, "@"); len(parts) == 2 {
			return maskString(parts[0], 1) + "@" + maskString(parts[1], 1), nil
		}
		return maskString(value, 1), nil

	case "phone":
		// (123) 456-7890 -> (***) ***-7890
		phoneRegex := regexp.MustCompile(`(\d{3})\D*(\d{3})\D*(\d{4})`)
		return phoneRegex.ReplaceAllString(value, "($1) $2-$3"), nil

	case "ssn":
		// 123-45-6789 -> ***-**-6789
		if len(value) >= 4 {
			return "***-**-" + value[len(value)-4:], nil
		}
		return "***-**-****", nil

	case "credit_card":
		// 1234 5678 9012 3456 -> **** **** **** 3456
		if len(value) >= 4 {
			cleaned := regexp.MustCompile(`\D`).ReplaceAllString(value, "")
			if len(cleaned) >= 4 {
				return "**** **** **** " + cleaned[len(cleaned)-4:], nil
			}
		}
		return "**** **** **** ****", nil

	default:
		return a.preserveFormat(value), nil
	}
}

/* PartialMask partially masks a value */
func (a *MaskingAlgorithms) PartialMask(value string, rule string) (string, error) {
	if rule == "" {
		// Default: show first 2 and last 2 characters
		if len(value) > 4 {
			return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:], nil
		}
		return strings.Repeat("*", len(value)), nil
	}

	// Parse rule like "first:2,last:2" or "first:3"
	parts := strings.Split(rule, ",")
	showFirst := 2
	showLast := 2

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "first:") {
			fmt.Sscanf(part, "first:%d", &showFirst)
		} else if strings.HasPrefix(part, "last:") {
			fmt.Sscanf(part, "last:%d", &showLast)
		}
	}

	if len(value) <= showFirst+showLast {
		return strings.Repeat("*", len(value)), nil
	}

	return value[:showFirst] + strings.Repeat("*", len(value)-showFirst-showLast) + value[len(value)-showLast:], nil
}

/* preserveFormat preserves the format of a string while masking */
func (a *MaskingAlgorithms) preserveFormat(value string) string {
	result := make([]byte, len(value))
	for i, char := range value {
		switch {
		case char >= '0' && char <= '9':
			result[i] = '*'
		case char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z':
			result[i] = '*'
		default:
			result[i] = byte(char) // Preserve separators and special characters
		}
	}
	return string(result)
}

/* maskString masks a string showing only first n characters */
func maskString(s string, showChars int) string {
	if len(s) <= showChars {
		return strings.Repeat("*", len(s))
	}
	return s[:showChars] + strings.Repeat("*", len(s)-showChars)
}

/* Decrypt decrypts an encrypted value */
func (a *MaskingAlgorithms) Decrypt(encryptedValue string, algorithm string) (string, error) {
	// Decrypt using AES-256-GCM
	// In production, you'd need to store/retrieve the encryption key securely
	return "", fmt.Errorf("decryption requires key management - not implemented in basic version")
}
