package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	name := flag.String("name", "Development Test Key", "API key name")
	rateLimit := flag.Int("rate-limit", 10000, "Rate limit per hour")
	flag.Parse()

	// Generate a random API key
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating key: %v\n", err)
		os.Exit(1)
	}
	key := "test-key-" + hex.EncodeToString(keyBytes)
	prefix := key[:8]

	// Hash with SHA256
	hasher := sha256.New()
	hasher.Write([]byte(key))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	// Bcrypt the SHA256 hash
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(keyHash), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error hashing key: %v\n", err)
		os.Exit(1)
	}

	// Insert into database
	sql := fmt.Sprintf(
		`INSERT INTO neuronip.api_keys (key_hash, key_prefix, name, rate_limit, created_at)
		VALUES ('%s', '%s', '%s', %d, NOW());`,
		strings.ReplaceAll(string(bcryptHash), "'", "''"),
		prefix,
		strings.ReplaceAll(*name, "'", "''"),
		*rateLimit,
	)

	cmd := exec.Command("psql", "neuronip", "-c", sql)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting into database: %v\n", err)
		fmt.Fprintf(os.Stderr, "Output: %s\n", output)
		os.Exit(1)
	}

	fmt.Println("✅ API Key created successfully!")
	fmt.Println()
	fmt.Println("📋 Use this token in your frontend:")
	fmt.Printf("   %s\n", key)
	fmt.Println()
	fmt.Println("🔧 Set it in browser console (F12 → Console):")
	fmt.Printf("   localStorage.setItem('api_token', '%s')\n", key)
	fmt.Println()
	fmt.Println("Then refresh your warehouse page!")
}
