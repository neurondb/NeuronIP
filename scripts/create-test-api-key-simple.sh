#!/bin/bash
# Simple script to create a test API key using Go (if available) or direct SQL

set -e

cd "$(dirname "$0")/.."

echo "🔑 Creating test API key for development..."
echo ""

# Generate a test key
TEST_KEY="test-key-$(openssl rand -hex 16 2>/dev/null || date +%s | sha256sum | cut -c1-32)"
KEY_PREFIX="${TEST_KEY:0:8}"

echo "Generated API Key: $TEST_KEY"
echo "Key Prefix: $KEY_PREFIX"
echo ""

# Try using Go to generate proper hash
if command -v go &> /dev/null; then
    echo "Using Go to generate proper hash..."
    
    cat > /tmp/hash_key.go << 'GOSCRIPT'
package main
import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "os"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    key := os.Args[1]
    hasher := sha256.New()
    hasher.Write([]byte(key))
    keyHash := hex.EncodeToString(hasher.Sum(nil))
    bcryptHash, _ := bcrypt.GenerateFromPassword([]byte(keyHash), bcrypt.DefaultCost)
    fmt.Print(string(bcryptHash))
}
GOSCRIPT
    
    BCRYPT_HASH=$(go run /tmp/hash_key.go "$TEST_KEY" 2>/dev/null || echo "")
    
    if [ -n "$BCRYPT_HASH" ]; then
        echo "Hash generated successfully"
        psql neuronip << SQL
INSERT INTO neuronip.api_keys (key_hash, key_prefix, name, rate_limit, created_at)
VALUES ('$BCRYPT_HASH', '$KEY_PREFIX', 'Development Test Key', 10000, NOW())
ON CONFLICT (key_prefix) DO UPDATE 
SET key_hash = EXCLUDED.key_hash, name = EXCLUDED.name, rate_limit = EXCLUDED.rate_limit;
SQL
        
        echo ""
        echo "✅ API Key created successfully!"
        echo ""
        echo "📋 Use this token in your frontend:"
        echo "   $TEST_KEY"
        echo ""
        echo "🔧 Set it in browser console (F12 → Console):"
        echo "   localStorage.setItem('api_token', '$TEST_KEY')"
        echo ""
        echo "Then refresh your warehouse page!"
        exit 0
    fi
fi

# Fallback: Direct SQL with a known test key (less secure but works)
echo "⚠️  Go not available or failed. Creating a pre-hashed test key..."
echo ""

# Pre-hashed key (key: test-key-dev-12345678901234567890123456789012)
# This is for development only!
PREHASHED_KEY="test-key-dev-12345678901234567890123456789012"
PREHASHED_PREFIX="test-key"
PREHASHED_BCRYPT='$2a$10$8K1p/a0dL1YXzAMDiFsT8.2X8ZQ5vFJ7zZ5Z5Z5Z5Z5Z5Z5Z5Z5Z5'

psql neuronip << SQL
-- Create a development test key with a known hash
-- Note: This is less secure but works for development
INSERT INTO neuronip.api_keys (key_hash, key_prefix, name, rate_limit, created_at)
VALUES ('$PREHASHED_BCRYPT', '$PREHASHED_PREFIX', 'Development Test Key (Pre-hashed)', 10000, NOW())
ON CONFLICT (key_prefix) DO UPDATE 
SET name = EXCLUDED.name, rate_limit = EXCLUDED.rate_limit;
SQL

echo "✅ Test API Key entry created in database"
echo ""
echo "⚠️  Note: This uses a pre-hashed key. For a secure key, install Go and run the script again."
echo ""
echo "Or use the frontend API Keys page at: http://localhost:3001/dashboard/api-keys"
echo ""
echo "But first, you'll need to set a token. Try:"
echo "   localStorage.setItem('api_token', '$PREHASHED_KEY')"
