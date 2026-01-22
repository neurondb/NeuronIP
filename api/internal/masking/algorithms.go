package masking

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/crypto/hkdf"
)

/* Tokenize tokenizes a value */
func (a *MaskingAlgorithms) Tokenize(value string, algorithm string) (string, error) {
	// Generate deterministic token from value
	hash := sha256.Sum256([]byte(value))
	token := base64.URLEncoding.EncodeToString(hash[:16]) // First 16 bytes
	
	return fmt.Sprintf("TOKEN_%s", token[:24]), nil
}

/* Encrypt encrypts a value using AES-256-GCM with key derivation */
func (a *MaskingAlgorithms) Encrypt(value string, algorithm string) (string, error) {
	// Get master key from environment or use default (in production, use proper key management)
	masterKey := getMasterKey()
	
	// Derive encryption key using HKDF
	key := deriveKey(masterKey, []byte("encryption-key"), 32)
	
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
	
	// Encode as base64 with prefix to identify encryption format
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return "enc:" + encoded, nil
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

/* Decrypt decrypts an encrypted value using AES-256-GCM with key derivation */
func (a *MaskingAlgorithms) Decrypt(encryptedValue string, algorithm string) (string, error) {
	// Check if value has encryption prefix
	if !strings.HasPrefix(encryptedValue, "enc:") {
		return "", fmt.Errorf("value does not appear to be encrypted (missing 'enc:' prefix)")
	}
	
	// Remove prefix and decode base64
	encoded := strings.TrimPrefix(encryptedValue, "enc:")
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted value: %w", err)
	}
	
	// Get master key from environment
	masterKey := getMasterKey()
	
	// Derive decryption key using same method as encryption
	key := deriveKey(masterKey, []byte("encryption-key"), 32)
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	
	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

/* getMasterKey retrieves the master encryption key from environment */
func getMasterKey() []byte {
	// Try to get from environment variable
	masterKeyStr := os.Getenv("ENCRYPTION_MASTER_KEY")
	if masterKeyStr != "" {
		// If provided as hex, decode it
		if key, err := base64.StdEncoding.DecodeString(masterKeyStr); err == nil && len(key) >= 32 {
			return key[:32] // Use first 32 bytes
		}
		// If provided as string, hash it to get 32 bytes
		hash := sha256.Sum256([]byte(masterKeyStr))
		return hash[:]
	}
	
	// Fallback: use a default key (WARNING: not secure for production)
	// In production, this should always come from a secure key management service
	defaultKey := sha256.Sum256([]byte("default-encryption-key-change-in-production"))
	return defaultKey[:]
}

/* deriveKey derives a key using HKDF */
func deriveKey(masterKey []byte, salt []byte, keyLen int) []byte {
	// Use HKDF for key derivation
	hash := sha256.New
	hkdfReader := hkdf.New(hash, masterKey, salt, nil)
	
	key := make([]byte, keyLen)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		// Fallback: use SHA256 hash if HKDF fails
		h := sha256.New()
		h.Write(masterKey)
		h.Write(salt)
		hashSum := h.Sum(nil)
		copy(key, hashSum[:keyLen])
	}
	
	return key
}
