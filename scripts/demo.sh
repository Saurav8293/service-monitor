#!/usr/bin/env bash
set -e

echo "=========================================================="
echo "   Service Monitor & Incident Platform - Live Demo        "
echo "=========================================================="

BASE_URL="http://localhost:8080"

echo -e "\n[Step 1] Authenticating as admin to obtain JWT token..."
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4)

echo "✓ Successfully obtained JWT: ${TOKEN:0:25}..."

echo -e "\n[Step 2] Registering target microservices (Payment & Order)..."
curl -s -X POST "$BASE_URL/api/v1/services" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"payment-service","endpoint":"http://payment-service:8081/health"}' || true

curl -s -X POST "$BASE_URL/api/v1/services" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"order-service","endpoint":"http://order-service:8082/health"}' || true

echo -e "\n[Step 3] Querying registered services..."
curl -s -X GET "$BASE_URL/api/v1/services" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n[Step 4] Simulating failure on payment-service (:8081)..."
curl -s -X POST "http://localhost:8081/simulate-outage?down=true"
echo "✓ Outage simulated. Sleeping 7s for Worker Pool and Kafka event..."
sleep 7

echo -e "\n[Step 5] Checking monitored service status after outage..."
curl -s -X GET "$BASE_URL/api/v1/services" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n[Step 6] Simulating recovery on payment-service..."
curl -s -X POST "http://localhost:8081/simulate-outage?down=false"
echo "✓ Recovery simulated. Sleeping 7s for resolution..."
sleep 7

echo -e "\n[Step 7] Checking final monitored service status..."
curl -s -X GET "$BASE_URL/api/v1/services" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n=========================================================="
echo "Demo complete! Access your services:"
echo "• Grafana Dashboard:  http://localhost:3000 (admin / admin)"
echo "• Prometheus Metrics: http://localhost:9090"
echo "• Nginx API Gateway:  https://localhost/api/v1/services"
echo "=========================================================="
