#!/usr/bin/env python3
"""
Create a test API key for development
This script generates a test API key and stores it in the database
"""

import sys
import hashlib
import bcrypt
import secrets
import subprocess
import os

def generate_api_key():
    """Generate a secure API key"""
    key = f"test-key-{secrets.token_hex(16)}"
    return key

def hash_api_key(key):
    """Hash the API key using SHA256 + bcrypt (matches Go implementation)"""
    # Step 1: SHA256 hash
    sha256_hash = hashlib.sha256(key.encode()).hexdigest()
    
    # Step 2: Bcrypt the SHA256 hash
    bcrypt_hash = bcrypt.hashpw(sha256_hash.encode(), bcrypt.gensalt()).decode()
    
    return bcrypt_hash

def create_api_key_in_db(key, key_prefix, bcrypt_hash):
    """Insert API key into database"""
    db_command = f"""
    INSERT INTO neuronip.api_keys (key_hash, key_prefix, name, rate_limit, created_at)
    VALUES ('{bcrypt_hash}', '{key_prefix}', 'Development Test Key', 10000, NOW())
    ON CONFLICT (key_prefix) DO UPDATE 
    SET key_hash = EXCLUDED.key_hash, name = EXCLUDED.name, rate_limit = EXCLUDED.rate_limit;
    """
    
    result = subprocess.run(
        ['psql', 'neuronip', '-c', db_command],
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        print(f"Error creating API key: {result.stderr}", file=sys.stderr)
        return False
    return True

def main():
    print("🔑 Creating test API key for development...")
    print()
    
    # Check if bcrypt is available
    try:
        import bcrypt
    except ImportError:
        print("❌ Error: bcrypt module not found")
        print("   Install it with: pip install bcrypt")
        sys.exit(1)
    
    # Generate key
    key = generate_api_key()
    key_prefix = key[:8]
    
    print(f"Generated API Key: {key}")
    print(f"Key Prefix: {key_prefix}")
    print()
    
    # Hash the key
    print("Hashing API key...")
    bcrypt_hash = hash_api_key(key)
    
    # Insert into database
    print("Inserting into database...")
    if create_api_key_in_db(key, key_prefix, bcrypt_hash):
        print()
        print("✅ API Key created successfully!")
        print()
        print("📋 Use this token in your frontend:")
        print(f"   {key}")
        print()
        print("🔧 Set it in browser console (F12 → Console):")
        print(f"   localStorage.setItem('api_token', '{key}')")
        print()
        print("Then refresh your warehouse page!")
        return 0
    else:
        print("❌ Failed to create API key in database")
        return 1

if __name__ == "__main__":
    sys.exit(main())
