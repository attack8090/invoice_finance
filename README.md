# AI-Enabled Invoice Financing Platform

A comprehensive SME fintech marketplace that leverages blockchain technology, AI, and modern web technologies to facilitate invoice financing between SMEs and investors.

## Architecture

### Backend (Go)
- REST API built with Gin framework
- MongoDB NoSQL database
- JWT authentication
- AI-powered credit scoring and fraud detection
- Blockchain integration for transparency

### Frontend (React)
- Modern responsive web application
- TypeScript for type safety
- Material-UI/Tailwind CSS for styling
- Web3 integration for blockchain interactions
- Separate dashboards for SMEs, Investors, and Admins

### Blockchain (Ethereum/Solidity)
- Smart contracts for invoice tokenization
- Automated escrow and payment systems
- Immutable transaction records
- Decentralized verification

## Key Features

### For SMEs (Small/Medium Enterprises)
- Upload and verify invoices
- Request financing with AI-powered risk assessment
- Track financing status and payments
- Automated invoice validation using AI/OCR
- Real-time dashboard with analytics

### For Investors
- Browse available financing opportunities
- AI-powered investment recommendations
- Risk assessment and portfolio management
- Automated investment strategies
- ROI tracking and analytics

### AI-Enabled Features
- Credit scoring algorithm
- Fraud detection system
- Invoice authenticity verification
- Market trend analysis
- Automated risk assessment

### Blockchain Features
- Invoice tokenization as NFTs
- Smart contract-based escrow
- Immutable audit trail
- Decentralized verification
- Automated payments upon invoice settlement

## Current Working Environment

**Development System:**
- **Platform**: Windows 11
- **Shell**: PowerShell 5.1.22000.2538
- **Working Directory**: `C:\Users\pc\Documents\Github\invoice_finance`
- **Last Updated**: September 1, 2025

**Active Services & Versions:**
- **Go Backend**: Go 1.23.0 (toolchain 1.24.5) with Gin framework
- **React Frontend**: React 19.1.1 with TypeScript 4.9.5
- **AI Service**: Python FastAPI with scikit-learn, LightGBM, XGBoost
- **Database**: MongoDB 7 with multiple databases
- **Cache**: Redis 7-alpine
- **Blockchain**: Hardhat 3.0.1 with OpenZeppelin 5.4.0
- **Container Orchestration**: Docker Compose 3.8

## Tech Stack

- **Backend**: Go 1.23.0, Gin, MongoDB Driver, JWT Auth, CORS
- **Frontend**: React 19.1.1, TypeScript 4.9.5, Material-UI 7.3.1, Web3 4.16.0
- **AI/ML**: FastAPI 0.104.1, scikit-learn 1.3.2, LightGBM 4.1.0, XGBoost 2.0.1, Tesseract OCR
- **Blockchain**: Hardhat 3.0.1, Solidity, OpenZeppelin Contracts 5.4.0
- **Database**: MongoDB 7 (NoSQL document database), Redis 7 (caching)
- **DevOps**: Docker, Docker Compose
- **Development**: Windows PowerShell environment

## Quick Start

### Prerequisites
- Go 1.19+
- Node.js 18+
- MongoDB 7+
- Docker & Docker Compose
- Python 3.11+ (for AI services)

### Development Setup

**Using Docker Compose (Recommended):**
1. Clone the repository
2. Run `docker-compose up -d` to start all services
3. MongoDB will be automatically initialized with databases and collections
4. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - AI Service: http://localhost:5000
   - Blockchain Node: http://localhost:8545
   - MongoDB: mongodb://localhost:27017

**Manual Development Setup:**
1. Start MongoDB and Redis services locally
2. Navigate to `backend/` and run `go run cmd/server/main.go` (Port 8080)
3. Navigate to `frontend/` and run `npm start` (Port 3000)
4. Navigate to `ai-service/` and run `uvicorn main:app --reload --port 5000`
5. Deploy smart contracts: `cd blockchain && npx hardhat deploy`

## Microservices Architecture

**Core Services:**
- **Backend API** (Go:8080) - Main REST API and business logic
- **Frontend** (React:3000) - Web application interface
- **AI Service** (Python:5000) - ML models for credit scoring and fraud detection
- **Blockchain Node** (Hardhat:8545) - Local Ethereum development node

**Specialized Services:**
- **User Management** (Go:8081) - Authentication and user profiles
- **Financing Workflow** (Go:8082) - Invoice financing process management
- **Credit Scoring** (Go:8083) - Credit assessment algorithms
- **Blockchain Ledger** (Go:8084) - Epic 4 compliance, Hyperledger Fabric integration
- **Payment Service** (Python:8085) - Epic 5 compliance, payment processing
- **Notification Service** (Python:8086) - Email, SMS, and push notifications
- **Bank Integration** (Go:8087) - Banking API connections and Epic 4 compliance
- **External Data Integration** (Python:8088) - Third-party data sources
- **Document Management** (Go:8089) - Document storage and management
- **OCR Service** (Go:8090) - Invoice text extraction and processing

**Infrastructure:**
- **MongoDB** (Port 27017) - Primary NoSQL database with multiple databases
- **Redis** (Port 6379) - Caching and session management

## Project Structure

```
invoice_finance/
├── backend/                          # Main Go API server (Port 8080)
├── frontend/                         # React web application (Port 3000)
├── ai-service/                       # Python FastAPI ML service (Port 5000)
├── blockchain/                       # Hardhat & smart contracts (Port 8545)
├── user-management-service/          # Go user service (Port 8081)
├── financing-workflow-service/       # Go workflow service (Port 8082)
├── credit-scoring-service/          # Go credit service (Port 8083)
├── blockchain-ledger-service/       # Go blockchain ledger (Port 8084)
├── payment-service/                 # Python payment service (Port 8085)
├── notification-service/            # Python notification service (Port 8086)
├── bank-integration-service/        # Go bank integration (Port 8087)
├── integration-external-data-service/ # Python external data (Port 8088)
├── document-management-service/     # Go document service (Port 8089)
├── ocr-service/                     # Go OCR service (Port 8090)
├── chaincode/                       # Hyperledger Fabric chaincode
├── scripts/                         # Database and deployment scripts
└── docker-compose.yml               # Complete service orchestration
```

## License
MIT License - see LICENSE file for details
