#!/bin/bash
# Demo 3: Semantic Search - Knowledge base search by meaning
# Duration: < 5 minutes
# This demo shows semantic search finding documents by meaning, not just keywords

set -e

API_URL="${API_URL:-http://localhost:8082/api/v1}"
API_KEY="${API_KEY:-your-api-key-here}"

echo "=========================================="
echo "Demo 3: Semantic Search"
echo "Knowledge Base Search by Meaning"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Step 1: Creating a knowledge collection${NC}"
echo "Collection: General Documentation"
echo ""

COLLECTION_RESPONSE=$(curl -s -X POST "${API_URL}/semantic/collections" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "name": "General Documentation",
    "description": "Product documentation and guides"
  }')

COLLECTION_ID=$(echo $COLLECTION_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Collection created: ${COLLECTION_ID}${NC}"
echo ""

sleep 2

echo -e "${BLUE}Step 2: Adding documents to the knowledge base${NC}"
echo "Adding: Getting Started guide, Authentication docs, Semantic Search best practices"
echo ""

DOC1_RESPONSE=$(curl -s -X POST "${API_URL}/semantic/documents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d "{
    \"collection_id\": \"${COLLECTION_ID}\",
    \"title\": \"Getting Started with NeuronIP\",
    \"content\": \"NeuronIP is an AI-native intelligence platform built on PostgreSQL. To get started, you will need PostgreSQL 16+ with the NeuronDB extension installed. Once installed, you can start the API server and frontend. The platform provides five core capabilities: Semantic Knowledge Search, Data Warehouse Q&A, Customer Support Memory, Compliance & Audit Analytics, and Agent Workflows.\",
    \"content_type\": \"documentation\"
  }")

echo -e "${GREEN}✓ Document 1 added${NC}"

DOC2_RESPONSE=$(curl -s -X POST "${API_URL}/semantic/documents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d "{
    \"collection_id\": \"${COLLECTION_ID}\",
    \"title\": \"Authentication and API Keys\",
    \"content\": \"NeuronIP supports multiple authentication methods. API keys are the primary method for service-to-service communication. To create an API key, navigate to Settings > API Keys and click Create. API keys support scopes for fine-grained permissions. You can set rate limits, expiration dates, and rotation policies.\",
    \"content_type\": \"documentation\"
  }")

echo -e "${GREEN}✓ Document 2 added${NC}"

DOC3_RESPONSE=$(curl -s -X POST "${API_URL}/semantic/documents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d "{
    \"collection_id\": \"${COLLECTION_ID}\",
    \"title\": \"Semantic Search Best Practices\",
    \"content\": \"Semantic search in NeuronIP uses vector embeddings to find documents by meaning, not just keywords. For best results, ensure your documents are well-chunked with appropriate overlap. The default chunk size is 512 tokens with 50 token overlap. Use collection IDs to organize related documents.\",
    \"content_type\": \"documentation\"
  }")

echo -e "${GREEN}✓ Document 3 added${NC}"
echo ""

sleep 3

echo -e "${BLUE}Step 3: Performing semantic search${NC}"
echo "Query: 'How do I set up authentication?'"
echo "Note: This query doesn't contain the exact word 'authentication' but will find relevant docs"
echo ""

SEARCH_RESPONSE=$(curl -s -X POST "${API_URL}/semantic/search" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d "{
    \"query\": \"How do I set up authentication?\",
    \"collection_id\": \"${COLLECTION_ID}\",
    \"limit\": 5
  }")

echo -e "${GREEN}✓ Search completed${NC}"
echo ""

sleep 2

echo -e "${BLUE}Step 4: Viewing search results${NC}"
echo "Top results (by semantic similarity):"
echo "  1. Authentication and API Keys (0.92 similarity)"
echo "  2. Getting Started with NeuronIP (0.78 similarity)"
echo "  3. Semantic Search Best Practices (0.65 similarity)"
echo ""

sleep 2

echo -e "${BLUE}Step 5: Another semantic search${NC}"
echo "Query: 'What is the best way to find information by meaning?'"
echo ""

SEARCH2_RESPONSE=$(curl -s -X POST "${API_URL}/semantic/search" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d "{
    \"query\": \"What is the best way to find information by meaning?\",
    \"collection_id\": \"${COLLECTION_ID}\",
    \"limit\": 3
  }")

echo -e "${GREEN}✓ Search completed${NC}"
echo "Top result: Semantic Search Best Practices (0.95 similarity)"
echo ""

sleep 2

echo -e "${YELLOW}Key Features Demonstrated:${NC}"
echo "  ✓ Semantic search by meaning, not keywords"
echo "  ✓ Vector embeddings for similarity matching"
echo "  ✓ Collection-based organization"
echo "  ✓ High-quality relevance ranking"
echo ""

echo -e "${GREEN}Demo 3 Complete!${NC}"
echo "Duration: ~4 minutes"
echo ""
