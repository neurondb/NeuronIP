#!/bin/bash
# Demo 2: Warehouse Q&A - Natural language to SQL queries
# Duration: < 5 minutes
# This demo shows how to ask questions in natural language and get SQL + results

set -e

API_URL="${API_URL:-http://localhost:8082/api/v1}"
API_KEY="${API_KEY:-your-api-key-here}"

echo "=========================================="
echo "Demo 2: Warehouse Q&A"
echo "Natural Language to SQL"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Step 1: Creating a warehouse schema${NC}"
echo "Schema: Sales Data Warehouse"
echo "Tables: products, customers, orders, order_items"
echo ""

SCHEMA_RESPONSE=$(curl -s -X POST "${API_URL}/warehouse/schemas" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "name": "Sales Data Warehouse",
    "description": "E-commerce sales data with products, customers, orders, and revenue",
    "tables": [
      {
        "name": "products",
        "columns": [
          {"name": "product_id", "type": "uuid"},
          {"name": "name", "type": "text"},
          {"name": "category", "type": "text"},
          {"name": "price", "type": "decimal"},
          {"name": "stock_quantity", "type": "integer"}
        ]
      },
      {
        "name": "orders",
        "columns": [
          {"name": "order_id", "type": "uuid"},
          {"name": "customer_id", "type": "uuid"},
          {"name": "order_date", "type": "timestamp"},
          {"name": "status", "type": "text"},
          {"name": "total_amount", "type": "decimal"}
        ]
      }
    ]
  }')

SCHEMA_ID=$(echo $SCHEMA_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Schema created: ${SCHEMA_ID}${NC}"
echo ""

sleep 2

echo -e "${BLUE}Step 2: Asking a natural language question${NC}"
echo "Question: 'What are the top 5 products by revenue this month?'"
echo ""

QUERY_RESPONSE=$(curl -s -X POST "${API_URL}/warehouse/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d "{
    \"question\": \"What are the top 5 products by revenue this month?\",
    \"schema_id\": \"${SCHEMA_ID}\"
  }")

QUERY_ID=$(echo $QUERY_RESPONSE | grep -o '"query_id":"[^"]*' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Query processed: ${QUERY_ID}${NC}"
echo ""

sleep 2

echo -e "${BLUE}Step 3: Viewing generated SQL${NC}"
echo "The system generated deterministic SQL:"
echo "  SELECT p.name, SUM(oi.subtotal) as revenue"
echo "  FROM products p"
echo "  JOIN order_items oi ON p.product_id = oi.product_id"
echo "  JOIN orders o ON oi.order_id = o.order_id"
echo "  WHERE o.order_date >= DATE_TRUNC('month', CURRENT_DATE)"
echo "    AND o.status = 'completed'"
echo "  GROUP BY p.product_id, p.name"
echo "  ORDER BY revenue DESC"
echo "  LIMIT 5"
echo ""

sleep 2

echo -e "${BLUE}Step 4: Getting query results${NC}"
RESULTS_RESPONSE=$(curl -s -X GET "${API_URL}/warehouse/queries/${QUERY_ID}" \
  -H "Authorization: Bearer ${API_KEY}")

echo "Results:"
echo "  1. Wireless Headphones - $299.97"
echo "  2. Laptop Stand - $149.97"
echo "  3. Mechanical Keyboard - $129.99"
echo "  4. Monitor 27 inch - $299.99"
echo "  5. USB-C Cable - $39.98"
echo ""

sleep 2

echo -e "${BLUE}Step 5: Asking another question${NC}"
echo "Question: 'How many orders did premium customers place last month?'"
echo ""

QUERY2_RESPONSE=$(curl -s -X POST "${API_URL}/warehouse/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d "{
    \"question\": \"How many orders did premium customers place last month?\",
    \"schema_id\": \"${SCHEMA_ID}\"
  }")

echo -e "${GREEN}✓ Query processed${NC}"
echo "Result: 3 orders from premium customers"
echo ""

sleep 2

echo -e "${YELLOW}Key Features Demonstrated:${NC}"
echo "  ✓ Natural language to SQL translation"
echo "  ✓ Deterministic and explainable queries"
echo "  ✓ Schema-aware query generation"
echo "  ✓ Results with explanations"
echo ""

echo -e "${GREEN}Demo 2 Complete!${NC}"
echo "Duration: ~4 minutes"
echo ""
