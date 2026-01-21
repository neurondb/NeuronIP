#!/bin/bash
# Create a test API key for development

set -e

cd "$(dirname "$0")/.."

# Generate a test key
TEST_KEY="test-key-$(openssl rand -hex 16)"
KEY_PREFIX="${TEST_KEY:0:8}"

echo "🔑 Creating test API key..."
echo ""

# Use Go to generate the proper hash (SHA256 + bcrypt)
go run - << 'GOSCRIPT'
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <api-key>\n", os.Args[0])
		os.Exit(1)
	}
	
	key := os.Args[1]
	prefix := key[:8]
	
	// Hash with SHA256
	hasher := sha256.New()
	hasher.Write([]byte(key))
	keyHash := hex.EncodeToString(hasher.Sum(nil))
	
	// Bcrypt the SHA256 hash
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(keyHash), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("KEY=%s\n", key)
	fmt.Printf("PREFIX=%s\n", prefix)
	fmt.Printf("HASH=%s\n", string(bcryptHash))
}
GOSCRIPT "$TEST_KEY" > /tmp/key_info.txt

source /tmp/key_info.txt

# Insert into database
psql neuronip << SQL
INSERT INTO neuronip.api_keys (key_hash, key_prefix, name, rate_limit, created_at)
VALUES ('$HASH', '$PREFIX', 'Development Test Key', 10000, NOW())
ON CONFLICT (key_prefix) DO UPDATE 
SET key_hash = EXCLUDED.key_hash, name = EXCLUDED.name, rate_limit = EXCLUDED.rate_limit;
SQL

echo "✅ API Key created successfully!"
echo ""
echo "📋 Use this token in your frontend:"
echo "   $KEY"
echo ""
echo "🔧 Set it in browser console (F12):"
echo "   localStorage.setItem('api_token', '$KEY')"
echo ""
echo "Then refresh the page!"
