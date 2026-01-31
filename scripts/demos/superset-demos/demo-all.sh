#!/usr/bin/env bash
# NeuronIP full-stack demo: warehouse, workload, data products, notebooks, decision dashboards, ITSM, templates, AI assistant
set -e
API_BASE="${API_BASE:-http://localhost:8082/api/v1}"
API_KEY="${NEURONIP_API_KEY:-test-key-82f13cedd19abec5bdd9ffad70f3f774}"

echo "=== NeuronIP full-stack demo ==="

echo "--- 1. Warehouse & elastic analytics ---"
echo "1.1 Warehouse query"
curl -s -X POST "$API_BASE/warehouse/query" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"query":"What are the top 5 products by revenue?"}' | head -c 200
echo ""
echo "1.2 List workload queues"
curl -s -X GET "$API_BASE/warehouse/workload/queues" -H "Authorization: Bearer $API_KEY" | head -c 200
echo ""
echo "1.3 Create workload queue"
curl -s -X POST "$API_BASE/warehouse/workload/queues" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"name":"analytics","max_concurrency":10,"query_timeout_seconds":300}' | head -c 200
echo ""
echo "1.4 Create data product"
curl -s -X POST "$API_BASE/data-products" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"name":"Sales Summary","owner_id":"demo","version":"1.0.0","visibility":"private"}' | head -c 200
echo ""
echo "1.5 Cache stats"
curl -s -X GET "$API_BASE/warehouse/cache/stats" -H "Authorization: Bearer $API_KEY" | head -c 200
echo ""

echo "--- 2. Notebooks ---"
NB=$(curl -s -X POST "$API_BASE/notebooks" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"name":"Sales Analysis","owner_id":"demo","default_language":"sql"}')
echo "2.1 Create notebook: $NB" | head -c 200
echo ""
NB_ID=$(echo "$NB" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$NB_ID" ]; then
  echo "2.2 Add cell"
  curl -s -X POST "$API_BASE/notebooks/$NB_ID/cells" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"position":0,"cell_type":"sql","content":"SELECT 1"}' | head -c 200
  echo ""
  echo "2.3 Create run"
  curl -s -X POST "$API_BASE/notebooks/$NB_ID/runs" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"triggered_by":"demo"}' | head -c 200
  echo ""
fi

echo "--- 3. Decision dashboards & governance ---"
D=$(curl -s -X POST "$API_BASE/decision-dashboards" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"name":"Revenue Review","owner_id":"demo","layout":[],"visibility":"private"}')
echo "3.1 Create decision dashboard: $D" | head -c 200
echo ""
D_ID=$(echo "$D" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$D_ID" ]; then
  echo "3.2 Record dashboard run"
  curl -s -X POST "$API_BASE/decision-dashboards/$D_ID/runs" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"triggered_by":"demo","snapshot":{},"status":"completed"}' | head -c 200
  echo ""
fi
echo "3.3 Lineage graph"
curl -s -X GET "$API_BASE/lineage/graph" -H "Authorization: Bearer $API_KEY" | head -c 200
echo ""
echo "3.4 RLS policies"
curl -s -X GET "$API_BASE/governance/rls/policies" -H "Authorization: Bearer $API_KEY" | head -c 200
echo ""

echo "--- 4. ITSM (incidents, changes, runbooks) ---"
echo "4.1 Create incident"
curl -s -X POST "$API_BASE/itsm/incidents" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"title":"Login failure","description":"Users cannot login","priority":"high","requester_id":"demo"}' | head -c 200
echo ""
echo "4.2 Create change"
curl -s -X POST "$API_BASE/itsm/changes" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"title":"Deploy v2","description":"Deploy new version","change_type":"normal","requester_id":"demo"}' | head -c 200
echo ""
echo "4.3 List runbooks"
curl -s -X GET "$API_BASE/itsm/runbooks" -H "Authorization: Bearer $API_KEY" | head -c 200
echo ""

echo "--- 5. Knowledge workspace & AI ---"
echo "5.1 List page templates"
curl -s -X GET "$API_BASE/notion-ui/templates/pages" -H "Authorization: Bearer $API_KEY" | head -c 200
echo ""
echo "5.2 List database templates"
curl -s -X GET "$API_BASE/notion-ui/templates/databases" -H "Authorization: Bearer $API_KEY" | head -c 200
echo ""
echo "5.3 AI assistant (RAG)"
curl -s -X POST "$API_BASE/ai/assistant" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d '{"query":"What is NeuronIP?","limit":5}' | head -c 300
echo ""

echo "=== NeuronIP full-stack demo done ==="
