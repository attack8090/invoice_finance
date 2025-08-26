# Invoice Finance Platform - Startup Script
# This script starts all microservices using Docker Compose

Write-Host "🚀 Starting Invoice Finance Platform..." -ForegroundColor Green
Write-Host "This will start all microservices including:" -ForegroundColor Yellow
Write-Host "  • Frontend (React) - Port 3000" -ForegroundColor Cyan
Write-Host "  • Backend API (Go) - Port 8080" -ForegroundColor Cyan
Write-Host "  • User Management Service (Go) - Port 8081" -ForegroundColor Cyan
Write-Host "  • Financing Workflow Service (Go) - Port 8082" -ForegroundColor Cyan
Write-Host "  • Credit Scoring Service (Python) - Port 8083" -ForegroundColor Cyan
Write-Host "  • Blockchain Ledger Service (Go) - Port 8084" -ForegroundColor Cyan
Write-Host "  • Payment Service (Python) - Port 8085" -ForegroundColor Cyan
Write-Host "  • Notification Service (Python) - Port 8086" -ForegroundColor Cyan
Write-Host "  • Bank Integration Service (Go) - Port 8087" -ForegroundColor Cyan
Write-Host "  • Integration External Data Service (Python) - Port 8088" -ForegroundColor Cyan
Write-Host "  • Document Management Service (Python) - Port 8089" -ForegroundColor Cyan
Write-Host "  • OCR Service (Python) - Port 8090" -ForegroundColor Cyan
Write-Host "  • AI/ML Service (Python) - Port 5000" -ForegroundColor Cyan
Write-Host "  • Blockchain Node (Hardhat) - Port 8545" -ForegroundColor Cyan
Write-Host "  • PostgreSQL Database - Port 5432" -ForegroundColor Cyan
Write-Host "  • Redis Cache - Port 6379" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
try {
    docker --version | Out-Null
    Write-Host "✅ Docker is available" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}

# Check if docker-compose is available
try {
    docker-compose --version | Out-Null
    Write-Host "✅ Docker Compose is available" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker Compose is not available. Please install Docker Compose." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🔧 Building and starting all services..." -ForegroundColor Yellow

# Make the database script executable
if (Test-Path "./scripts/create-multiple-databases.sh") {
    Write-Host "✅ Database initialization script found" -ForegroundColor Green
} else {
    Write-Host "⚠️  Database initialization script not found" -ForegroundColor Yellow
}

# Start services with docker-compose
try {
    Write-Host "Starting with Docker Compose..." -ForegroundColor Blue
    docker-compose up --build -d
    
    Write-Host ""
    Write-Host "🎉 All services are starting up!" -ForegroundColor Green
    Write-Host "Please wait a few moments for all services to initialize..." -ForegroundColor Yellow
    
    Start-Sleep -Seconds 5
    
    Write-Host ""
    Write-Host "📊 Service Status:" -ForegroundColor Cyan
    docker-compose ps
    
    Write-Host ""
    Write-Host "🌐 Access Points:" -ForegroundColor Green
    Write-Host "  Frontend:        http://localhost:3000" -ForegroundColor Cyan
    Write-Host "  Main API:        http://localhost:8080" -ForegroundColor Cyan
    Write-Host "  API Docs:        http://localhost:8080/docs" -ForegroundColor Cyan
    Write-Host "  "
    Write-Host "Microservices Health Checks:" -ForegroundColor Green
    Write-Host "  User Management:    http://localhost:8081/health" -ForegroundColor Gray
    Write-Host "  Financing Workflow: http://localhost:8082/health" -ForegroundColor Gray
    Write-Host "  Credit Scoring:     http://localhost:8083/health" -ForegroundColor Gray
    Write-Host "  Blockchain Ledger:  http://localhost:8084/health" -ForegroundColor Gray
    Write-Host "  Payment Service:    http://localhost:8085/health" -ForegroundColor Gray
    Write-Host "  Notification:       http://localhost:8086/health" -ForegroundColor Gray
    Write-Host "  Bank Integration:   http://localhost:8087/health" -ForegroundColor Gray
    Write-Host "  External Data:      http://localhost:8088/health" -ForegroundColor Gray
    Write-Host "  Document Mgmt:      http://localhost:8089/health" -ForegroundColor Gray
    Write-Host "  OCR Service:        http://localhost:8090/health" -ForegroundColor Gray
    Write-Host "  AI/ML Service:      http://localhost:5000/health" -ForegroundColor Gray
    Write-Host ""
    Write-Host "🛠️  Useful Commands:" -ForegroundColor Magenta
    Write-Host "  View logs:       docker-compose logs -f [service-name]" -ForegroundColor Gray
    Write-Host "  Stop all:        docker-compose down" -ForegroundColor Gray
    Write-Host "  Restart service: docker-compose restart [service-name]" -ForegroundColor Gray
    Write-Host "  Check status:    docker-compose ps" -ForegroundColor Gray
    
} catch {
    Write-Host "❌ Failed to start services: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✨ Invoice Finance Platform is now running!" -ForegroundColor Green
Write-Host "Check the health endpoints above to verify all services are operational." -ForegroundColor Yellow
