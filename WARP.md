# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

This is an AI-Enabled Invoice Financing Platform - a comprehensive SME fintech marketplace that leverages blockchain technology, AI/ML, and modern web technologies to facilitate invoice financing between SMEs and investors.

**Key Technologies:**
- **Backend**: Go 1.21+ with Gin framework, PostgreSQL, Redis
- **Frontend**: React 19+ with TypeScript, Material-UI, Web3.js
- **Blockchain**: Ethereum/Solidity with Hardhat, OpenZeppelin contracts
- **AI/ML**: Python FastAPI service with scikit-learn, XGBoost, LightGBM, OCR capabilities
- **Infrastructure**: Docker Compose for orchestration

## Quick Start Commands

### Full Stack Development
```bash
# Start all services (recommended for development)
docker-compose up -d

# View logs for specific services
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f ai-service
```

### Backend Development (Go)
```bash
cd backend

# Install dependencies
go mod download

# Run the server directly
go run cmd/server/main.go

# Run tests
go test ./...

# Build binary
go build -o bin/server cmd/server/main.go
```

### Frontend Development (React)
```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm start

# Run tests
npm test

# Run tests in watch mode
npm test -- --watch

# Build for production
npm run build
```

### Blockchain Development (Hardhat)
```bash
cd blockchain

# Install dependencies
npm install

# Start local Ethereum node
npx hardhat node

# Deploy contracts to local network
npx hardhat run scripts/deploy.js --network localhost

# Run contract tests
npx hardhat test

# Compile contracts
npx hardhat compile
```

### AI Service Development (Python)
```bash
cd ai-service

# Create and activate virtual environment
python -m venv venv
# Windows: venv\Scripts\activate
# Unix/MacOS: source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run the service
python main.py

# Run with Gunicorn for production-like testing
gunicorn main:app --workers 4 --worker-class uvicorn.workers.UvicornWorker --bind 0.0.0.0:5000

# Run tests
pytest

# Code formatting
black .
isort .
flake8
```

### Microservice Development

#### User Management Service (Go)
```bash
cd user-management-service

# Install dependencies
go mod download

# Run the service
go run main.go

# Run tests
go test ./...

# Build binary
go build -o bin/user-service main.go
```

#### Document Management Service (Python)
```bash
cd document-management-service

# Create and activate virtual environment
python -m venv venv
source venv/bin/activate  # Unix/MacOS
# venv\Scripts\activate  # Windows

# Install dependencies
pip install -r requirements.txt

# Run the service
python main.py

# Run tests
pytest

# Code formatting and linting
black .
isort .
flake8
```

#### OCR Service (Python)
```bash
cd ocr-service

# Setup virtual environment
python -m venv venv
source venv/bin/activate

# Install dependencies (including Tesseract)
pip install -r requirements.txt

# Configure Gemini API key
export GEMINI_API_KEY=your-gemini-api-key

# Run the service
python main.py

# Test OCR extraction
curl -X POST "http://localhost:8084/api/v1/ocr/extract" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@sample-invoice.pdf" \
  -F "document_type=invoice"
```

#### Credit Scoring Service (Python)
```bash
cd credit-scoring-service

# Setup environment
python -m venv venv
source venv/bin/activate

# Install ML dependencies
pip install -r requirements.txt

# Configure IDD Core integration
export IDD_CORE_API_KEY=your-idd-core-key
export IDD_CORE_ENDPOINT=https://api.iddcore.com/v5.1

# Run the service
python main.py

# Test credit scoring
curl -X POST "http://localhost:8085/api/v1/credit-score/calculate" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user123","scoring_model":"hybrid"}'
```

#### Financing Workflow Service (Go)
```bash
cd financing-workflow-service

# Install dependencies
go mod download

# Run the service
go run main.go

# Run tests with coverage
go test -cover ./...

# Test workflow creation
curl -X POST "http://localhost:8082/api/v1/financing/requests" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"invoice_id":"inv123","amount":50000}'
```

## Architecture Overview

### Complete Microservices Architecture
The platform follows a comprehensive microservices architecture with Epic compliance and advanced features:

1. **User Management Service (Go)**: at `localhost:8081`
   - KYC/AML compliance with Epic-1 standards
   - Multi-factor authentication (TOTP, SMS, Email)
   - Role-based access control (SME, Buyer, Bank, Admin)
   - Company onboarding and verification
   - Session management and audit logging
   - Compliance reporting and risk assessment

2. **Document Management Service (Python)**: at `localhost:8083`
   - AWS S3 integration with encrypted storage
   - Epic-2 compliance for document security
   - Virus scanning and file validation
   - Bulk upload and processing capabilities
   - Document versioning and audit trails
   - Integration with OCR service

3. **OCR Service (Python)**: at `localhost:8084`
   - Gemini AI integration for structured data extraction
   - Support for invoices, contracts, identity docs, bank statements
   - Multi-language OCR capabilities
   - Manual correction workflows
   - Table and signature extraction
   - Confidence scoring and validation

4. **Credit Scoring Service (Python)**: at `localhost:8085`
   - IDD Core 5.1 integration for enhanced risk assessment
   - Epic-5 compliance for credit scoring
   - Multiple scoring models (Traditional, Alternative, Hybrid, IDD Core)
   - Alternative data sources (social media, web presence, behavioral data)
   - Fraud detection and risk mitigation
   - Bulk scoring capabilities

5. **Financing Workflow Service (Go)**: at `localhost:8082`
   - Epic 4 and 5 compliance for end-to-end workflows
   - Invoice submission and buyer confirmation
   - Multi-party agreement signing with blockchain recording
   - Automated disbursement management
   - Dispute resolution and escalation
   - Integration with all other services

6. **Frontend (React)**: Web application at `localhost:3000`
   - Separate dashboards for SMEs, Buyers, Investors, Banks, and Administrators
   - Web3 integration for blockchain interactions
   - TypeScript for type safety, Material-UI for components
   - Real-time notifications and workflow tracking

7. **Blockchain (Hardhat)**: Ethereum contracts at `localhost:8545`
   - InvoiceNFT: Tokenizes invoices as NFTs for ownership representation
   - FinancingEscrow: Automated escrow system for secure transactions
   - Smart contract-based agreement recording
   - Deployed contracts are saved to `contract-addresses.json`

8. **Legacy AI Service (Python)**: ML/AI service at `localhost:5000`
   - Basic credit scoring algorithm
   - Risk assessment models
   - Fraud detection system
   - Invoice document verification with OCR
   - Market analysis and predictive analytics

9. **Database (PostgreSQL)**: at `localhost:5432`
   - Primary data storage with GORM for ORM
   - Database: `invoice_financing`, User: `admin`, Password: `password123` (development)
   - Shared across all Go services

10. **Cache (Redis)**: at `localhost:6379`
    - Session management and caching layer
    - Rate limiting data storage

### Key Code Patterns

#### Backend Structure (`backend/`)
- `cmd/server/main.go`: Application entry point
- `internal/`: Private application code
  - `api/`: HTTP handlers and server setup
  - `services/`: Business logic layer
  - `models/`: Data models and validation
  - `config/`: Configuration management
  - `database/`: Database initialization and migrations
- `middleware/`: Security, validation, rate limiting middleware
- `pkg/`: Public library code (JWT utilities)

#### Frontend Structure (`frontend/src/`)
- `App.tsx`: Main application component with routing
- `components/`: Reusable UI components
- `pages/`: Page-level components for different routes
- `hooks/`: Custom React hooks
- `services/`: API integration and utilities

#### Blockchain Structure (`blockchain/`)
- `contracts/`: Solidity smart contracts
- `scripts/deploy.js`: Contract deployment script
- `hardhat.config.js`: Hardhat configuration for local and testnet deployment

#### AI Service Structure (`ai-service/`)
- `main.py`: FastAPI application with all ML endpoints
- `config.py`: Comprehensive configuration with feature flags
- `models/ml_models.py`: ML model implementations
- `services/document_processor.py`: OCR and document processing

## Security Implementation

This platform implements comprehensive security measures (see `SECURITY.md` for details):

### Backend Security
- **Rate limiting**: Multi-tier (global: 60/min, auth: 5/min, uploads: 20/min, admin: 10/min)
- **Input validation**: Comprehensive validation middleware with business-specific validators
- **Security headers**: CORS, CSP, HSTS, X-Frame-Options, etc.
- **Authentication**: JWT with refresh token rotation, bcrypt password hashing

### Frontend Security
- **Client-side validation**: XSS prevention, input sanitization
- **Secure HTTP client**: Automatic token management, CSRF protection
- **Content Security Policy**: Enforced CSP headers

### Development vs Production Configuration
```bash
# Development
NODE_ENV=development
CORS_ALLOW_ALL=true
DEBUG=true

# Production
NODE_ENV=production
CORS_ALLOW_ALL=false
DEBUG=false
FORCE_HTTPS=true
```

## API Architecture

### Authentication Endpoints
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Token refresh

### Core Business Endpoints
- `GET /api/v1/invoices` - List user invoices
- `POST /api/v1/invoices` - Create new invoice
- `POST /api/v1/invoices/:id/verify` - Verify invoice
- `GET /api/v1/financing/requests` - List financing requests
- `POST /api/v1/financing/invest` - Make investment

### AI/ML Endpoints
- `POST /api/v1/ai/credit-score` - Calculate credit score
- `POST /api/v1/ai/risk-assessment` - Assess investment risk
- `POST /api/v1/ai/fraud-detection` - Detect fraud

### Blockchain Endpoints
- `POST /api/v1/blockchain/tokenize-invoice` - Tokenize invoice as NFT
- `GET /api/v1/blockchain/transactions/:hash` - Get transaction details

## Environment Configuration

### Required Environment Variables

#### Backend (Go)
```bash
DATABASE_URL=postgres://admin:password123@localhost:5432/invoice_financing?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-change-in-production
ETHEREUM_RPC=http://localhost:8545
CONTRACT_ADDRESS=0x0000000000000000000000000000000000000000
AI_MODEL_ENDPOINT=http://localhost:5000/api/ml
```

#### Frontend (React)
```bash
REACT_APP_API_URL=http://localhost:8080/api/v1
REACT_APP_WEB3_NETWORK=http://localhost:8545
```

#### AI Service (Python)
```bash
DATABASE_URL=postgresql://admin:password123@localhost:5432/invoice_financing
HOST=0.0.0.0
PORT=5000
DEBUG=false
ENABLE_CREDIT_SCORING=true
ENABLE_RISK_ASSESSMENT=true
ENABLE_FRAUD_DETECTION=true
ENABLE_DOCUMENT_VERIFICATION=true
```

## Testing Strategy

### Running Tests
```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test

# Blockchain tests
cd blockchain && npx hardhat test

# AI service tests
cd ai-service && pytest
```

### Test Structure
- Backend: Go's built-in testing framework with unit and integration tests
- Frontend: Jest and React Testing Library for component and integration tests
- Blockchain: Hardhat testing framework for smart contract tests
- AI Service: pytest for ML model and API endpoint tests

## Business Domain Knowledge

### Key Business Flows

1. **SME Invoice Financing Request**:
   - SME uploads invoice → AI verifies authenticity → Risk assessment → Blockchain tokenization → Available for investor funding

2. **Investor Funding Process**:
   - Browse opportunities → AI risk analysis → Investment decision → Blockchain escrow → Automated payouts

3. **AI-Powered Features**:
   - Credit scoring based on company data, financial history, transaction patterns
   - Fraud detection using behavioral analysis and document verification
   - Risk assessment considering invoice amount, customer rating, payment terms
   - OCR-based invoice verification with field extraction

### Role-Based Access Control
- **SME**: Upload invoices, request financing, track payments, view analytics
- **Investor**: Browse opportunities, make investments, manage portfolio, view returns
- **Admin**: Platform management, user verification, system monitoring

## Deployment Notes

### Docker Compose Services
- `postgres`: PostgreSQL database with auto-init from migrations
- `redis`: Redis cache server
- `backend`: Go API server with hot reload in development
- `frontend`: React development server
- `hardhat`: Local Ethereum node for blockchain development
- `ai-service`: Python FastAPI ML service

### Production Considerations
- Use proper secrets management (not environment variables in docker-compose)
- Enable HTTPS with valid SSL certificates
- Configure proper CORS origins
- Set up monitoring and logging (Prometheus metrics are included)
- Use production-grade database hosting
- Implement proper backup strategies
- Enable rate limiting and security headers

## Common Development Tasks

### Adding New API Endpoints
1. Define route in `backend/internal/api/handlers.go`
2. Implement business logic in appropriate service in `backend/internal/services/`
3. Add corresponding frontend API call in `frontend/src/services/`
4. Update TypeScript types if needed

### Adding New AI/ML Features
1. Define endpoint in `ai-service/main.py`
2. Implement model logic in `ai-service/models/ml_models.py`
3. Add feature flag in `ai-service/config.py`
4. Create corresponding backend integration
5. Update frontend to consume new AI capabilities

### Smart Contract Changes
1. Modify contracts in `blockchain/contracts/`
2. Update deployment script in `blockchain/scripts/deploy.js`
3. Recompile and deploy: `npx hardhat run scripts/deploy.js --network localhost`
4. Update contract addresses in backend configuration
5. Update frontend Web3 integration if needed
