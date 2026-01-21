#!/bin/bash
# Quick status check for NeuronIP services

echo "🔍 NeuronIP Service Status"
echo "=========================="
echo ""

# Check containers
echo "📦 Container Status:"
docker ps -a --filter "name=neuronip" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "No containers found"
echo ""

# Check if services are responding
echo "🌐 Service Health:"
echo -n "  Frontend (http://localhost:3001): "
if curl -s -o /dev/null -w "%{http_code}" http://localhost:3001 2>/dev/null | grep -q "200\|307"; then
    echo "✅ Running"
else
    echo "❌ Not responding"
fi

echo -n "  Backend API (http://localhost:8082/health): "
if curl -s http://localhost:8082/health 2>/dev/null | grep -q "ok\|healthy"; then
    echo "✅ Running"
else
    echo "❌ Not responding (may need database connection)"
fi
echo ""

# Show recent logs
echo "📋 Recent Logs:"
echo "  Backend API (last 3 lines):"
docker logs neuronip-api --tail 3 2>/dev/null | sed 's/^/    /' || echo "    Container not running"
echo ""
echo "  Frontend (last 3 lines):"
docker logs neuronip-frontend --tail 3 2>/dev/null | sed 's/^/    /' || echo "    Container not running"
echo ""

# Access information
echo "🔗 Access URLs:"
echo "  - Frontend UI: http://localhost:3001"
echo "  - Backend API: http://localhost:8082"
echo "  - API Health: http://localhost:8082/health"
echo ""
