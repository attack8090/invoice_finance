# Invoice Financing Platform - Development Startup Script
# Run this script to start all services for development

Write-Host "=== Invoice Financing Platform Development Setup ===" -ForegroundColor Green
Write-Host ""

# Check if Docker is running
Write-Host "Checking Docker status..." -ForegroundColor Yellow
try {
    docker version | Out-Null
    Write-Host "✅ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    Write-Host "After starting Docker Desktop, run this script again." -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "Starting infrastructure services..." -ForegroundColor Yellow

# Start only database and cache services
try {
    docker-compose up -d postgres redis
    Write-Host "✅ PostgreSQL and Redis started" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Docker services failed to start. Continuing with manual setup..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=== Manual Development Setup Instructions ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "1. Backend (Go API):" -ForegroundColor Yellow
Write-Host "   cd backend"
Write-Host "   go mod download"
Write-Host "   go run cmd/server/main.go"
Write-Host ""

Write-Host "2. Frontend (React):" -ForegroundColor Yellow
Write-Host "   cd frontend"
Write-Host "   npm install"
Write-Host "   npm start"
Write-Host ""

Write-Host "3. Blockchain (Hardhat):" -ForegroundColor Yellow
Write-Host "   cd blockchain"
Write-Host "   npm install"
Write-Host "   npx hardhat node"
Write-Host ""

Write-Host "4. AI Service (Python):" -ForegroundColor Yellow
Write-Host "   cd ai-service"
Write-Host "   pip install -r requirements.txt"
Write-Host "   python main.py"
Write-Host ""

Write-Host "5. Access the application:" -ForegroundColor Yellow
Write-Host "   Frontend: http://localhost:3000"
Write-Host "   Backend API: http://localhost:8080"
Write-Host "   AI Service: http://localhost:5000"
Write-Host "   Blockchain: http://localhost:8545"
Write-Host ""

Write-Host "=== Next Steps ===" -ForegroundColor Green
Write-Host "1. Make sure PostgreSQL is installed and running locally"
Write-Host "2. Create database: createdb invoice_financing"
Write-Host "3. Start each service in separate terminal windows"
Write-Host "4. Open http://localhost:3000 in your browser"
Write-Host ""

# Check if services are accessible
Write-Host "Checking service availability..." -ForegroundColor Yellow

# Test database connection (if running)
try {
    $null = Test-NetConnection -ComputerName localhost -Port 5432 -InformationLevel Quiet
    Write-Host "✅ PostgreSQL (5432) is accessible" -ForegroundColor Green
} catch {
    Write-Host "❌ PostgreSQL (5432) not accessible - install PostgreSQL locally" -ForegroundColor Red
}

# Test Redis connection (if running)
try {
    $null = Test-NetConnection -ComputerName localhost -Port 6379 -InformationLevel Quiet
    Write-Host "✅ Redis (6379) is accessible" -ForegroundColor Green
} catch {
    Write-Host "❌ Redis (6379) not accessible - will start with Docker or install locally" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Setup complete! Follow the manual setup instructions above." -ForegroundColor Green
