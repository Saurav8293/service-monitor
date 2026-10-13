# ==============================================================================
# Service Monitor & Incident Platform - End-to-End Simulation Demo
# ==============================================================================

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "   Service Monitor & Incident Platform - Live Demo        " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080"
$gatewayUrl = "https://localhost"
$headers = @{ "Content-Type" = "application/json" }

# Ignore self-signed cert warnings for demo
[System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}

# 1. Authenticate with Monitor Service to obtain JWT
Write-Host "`n[Step 1] Authenticating as admin to obtain JWT token..." -ForegroundColor Yellow
$loginBody = @{ username = "admin"; password = "admin123" } | ConvertTo-Json

try {
    $loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -Headers $headers -ErrorAction Stop
    $token = $loginResp.token
    Write-Host "✓ Successfully obtained JWT: $($token.Substring(0, 25))..." -ForegroundColor Green
} catch {
    Write-Host "✗ Login failed: $_" -ForegroundColor Red
    exit 1
}

$authHeaders = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer $token"
}

# 2. Register Payment & Order services
Write-Host "`n[Step 2] Registering target microservices (Payment & Order)..." -ForegroundColor Yellow

$paymentPayload = @{ name = "payment-service"; endpoint = "http://payment-service:8081/health" } | ConvertTo-Json
try {
    $payResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/services" -Method Post -Body $paymentPayload -Headers $authHeaders
    Write-Host "✓ Registered payment-service (ID: $($payResp.id))" -ForegroundColor Green
} catch {
    Write-Host "! payment-service might already be registered" -ForegroundColor Gray
}

$orderPayload = @{ name = "order-service"; endpoint = "http://order-service:8082/health" } | ConvertTo-Json
try {
    $ordResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/services" -Method Post -Body $orderPayload -Headers $authHeaders
    Write-Host "✓ Registered order-service (ID: $($ordResp.id))" -ForegroundColor Green
} catch {
    Write-Host "! order-service might already be registered" -ForegroundColor Gray
}

# 3. List services
Write-Host "`n[Step 3] Querying registered services..." -ForegroundColor Yellow
$services = Invoke-RestMethod -Uri "$baseUrl/api/v1/services" -Method Get -Headers $authHeaders
$services | Format-Table -Property id, name, endpoint, status

# 4. Trigger Outage Simulation on Payment Service
Write-Host "`n[Step 4] Simulating failure on payment-service (:8081)..." -ForegroundColor Red
try {
    Invoke-RestMethod -Uri "http://localhost:8081/simulate-outage?down=true" -Method Post
    Write-Host "✓ Simulated outage initiated! Waiting 7s for Worker Pool to detect DOWN & emit Kafka event..." -ForegroundColor Red
    Start-Sleep -Seconds 7
} catch {
    Write-Host "! Note: Payment service on localhost:8081 directly reachable: $_" -ForegroundColor Yellow
}

# 5. Check status after outage
Write-Host "`n[Step 5] Checking monitored service status after outage..." -ForegroundColor Yellow
$servicesAfterDown = Invoke-RestMethod -Uri "$baseUrl/api/v1/services" -Method Get -Headers $authHeaders
$servicesAfterDown | Format-Table -Property id, name, endpoint, status

# 6. Trigger Recovery Simulation
Write-Host "`n[Step 6] Simulating recovery on payment-service..." -ForegroundColor Green
try {
    Invoke-RestMethod -Uri "http://localhost:8081/simulate-outage?down=false" -Method Post
    Write-Host "✓ Recovery signal sent! Waiting 7s for Worker Pool to detect recovery & resolve incident..." -ForegroundColor Green
    Start-Sleep -Seconds 7
} catch {
    Write-Host "! Note: $_" -ForegroundColor Yellow
}

# 7. Final status check
Write-Host "`n[Step 7] Checking final monitored service status..." -ForegroundColor Yellow
$servicesFinal = Invoke-RestMethod -Uri "$baseUrl/api/v1/services" -Method Get -Headers $authHeaders
$servicesFinal | Format-Table -Property id, name, endpoint, status

Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host "Demo complete! Access your services:" -ForegroundColor Cyan
Write-Host "• Grafana Dashboard:  http://localhost:3000 (admin / admin)" -ForegroundColor White
Write-Host "• Prometheus Metrics: http://localhost:9090" -ForegroundColor White
Write-Host "• Nginx API Gateway:  https://localhost/api/v1/services" -ForegroundColor White
Write-Host "==========================================================" -ForegroundColor Cyan
