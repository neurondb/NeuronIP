#!/bin/bash
# Demo 1: Support Memory Hub - Support ticket resolution with long-term memory
# Duration: < 5 minutes
# This demo shows how Support Memory Hub remembers customer context across sessions

set -e

API_URL="${API_URL:-http://localhost:8082/api/v1}"
API_KEY="${API_KEY:-your-api-key-here}"

echo "=========================================="
echo "Demo 1: Support Memory Hub"
echo "Long-term Memory for Customer Support"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Step 1: Creating a support ticket for returning customer${NC}"
echo "Customer: John Doe (customer-123)"
echo "Issue: Password reset emails going to spam (follow-up from previous ticket)"
echo ""

TICKET_RESPONSE=$(curl -s -X POST "${API_URL}/support/tickets" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "customer_id": "customer-123",
    "subject": "Follow-up: Email deliverability improvement",
    "description": "Following up on previous ticket. I whitelisted the domain but still having issues with emails going to spam.",
    "priority": "medium"
  }')

TICKET_ID=$(echo $TICKET_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Ticket created: ${TICKET_ID}${NC}"
echo ""

sleep 2

echo -e "${BLUE}Step 2: System automatically retrieves customer history${NC}"
echo "The system remembers:"
echo "  - Previous ticket about password reset emails"
echo "  - Customer whitelisted domain previously"
echo "  - Email deliverability was an issue"
echo ""

sleep 2

echo -e "${BLUE}Step 3: Adding customer conversation${NC}"
CONV_RESPONSE=$(curl -s -X POST "${API_URL}/support/tickets/${TICKET_ID}/conversations" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "message": "Hi, I am following up on my previous ticket about password reset emails. I whitelisted your domain but emails are still going to spam. Can you check if there is something else I need to do?",
    "sender": "customer"
  }')

echo -e "${GREEN}✓ Conversation added${NC}"
echo ""

sleep 2

echo -e "${BLUE}Step 4: System finds similar past cases${NC}"
SIMILAR_RESPONSE=$(curl -s -X GET "${API_URL}/support/tickets/${TICKET_ID}/similar-cases" \
  -H "Authorization: Bearer ${API_KEY}")

echo "Found similar cases with solutions:"
echo "  - Previous password reset email issue (resolved)"
echo "  - Email deliverability improvements"
echo ""

sleep 2

echo -e "${BLUE}Step 5: Agent response with context${NC}"
echo "Agent can see:"
echo "  - Complete customer history"
echo "  - Previous resolutions"
echo "  - Updated SPF/DKIM records since last ticket"
echo ""

AGENT_RESPONSE=$(curl -s -X POST "${API_URL}/support/tickets/${TICKET_ID}/conversations" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "message": "I see your previous ticket. Let me check our email authentication records. I notice we have updated our SPF and DKIM records since your last ticket. I will send you updated instructions.",
    "sender": "agent"
  }')

echo -e "${GREEN}✓ Agent response sent with full context${NC}"
echo ""

sleep 2

echo -e "${YELLOW}Key Features Demonstrated:${NC}"
echo "  ✓ Long-term memory across sessions"
echo "  ✓ Automatic context retrieval"
echo "  ✓ Similar case matching"
echo "  ✓ Context-aware agent responses"
echo ""

echo -e "${GREEN}Demo 1 Complete!${NC}"
echo "Duration: ~3 minutes"
echo ""
