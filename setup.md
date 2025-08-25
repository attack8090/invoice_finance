# Invoice Financing Platform Setup Guide

This guide will help you set up and run the AI-Enabled Invoice Financing Platform locally.

## Prerequisites

- Docker and Docker Compose
- Node.js 18+ (for development)
- Go 1.21+ (for development)
- Python 3.11+ (for AI service development)

## Quick Start (Docker)

1. Clone or navigate to the project directory:
```bash
cd invoice-financing-platform
```

2. Start all services using Docker Compose:
```bash
docker-compose up -d
```

3. Wait for all services to be ready (this may take a few minutes on first run)

4. Access the services:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - AI Service: http://localhost:5000
   - Blockchain (Hardhat): http://localhost:8545

## Development Setup

### Backend (Go)

1. Navigate to the backend directory:
```bash
cd backend
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export DATABASE_URL="postgres://admin:password123@localhost:5432/invoice_financing?sslmode=disable"
export JWT_SECRET="your-secret-key"
export ETHEREUM_RPC="http://localhost:8545"
export AI_MODEL_ENDPOINT="http://localhost:5000/api/ml"
```

4. Run the server:
```bash
go run cmd/server/main.go
```

### Frontend (React)

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Set up environment variables in `.env`:
```
REACT_APP_API_URL=http://localhost:8080/api/v1
REACT_APP_WEB3_NETWORK=http://localhost:8545
```

4. Start the development server:
```bash
npm start
```

### Blockchain (Hardhat)

1. Navigate to the blockchain directory:
```bash
cd blockchain
```

2. Install dependencies:
```bash
npm install
```

3. Start local Ethereum node:
```bash
npx hardhat node
```

4. Deploy contracts (in a new terminal):
```bash
npx hardhat run scripts/deploy.js --network localhost
```

### AI Service (Python)

1. Navigate to the ai-service directory:
```bash
cd ai-service
```

2. Create virtual environment:
```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

3. Install dependencies:
```bash
pip install -r requirements.txt
```

4. Run the service:
```bash
python main.py
```

## Architecture Overview

### Components

1. **Backend (Go)**: REST API server handling authentication, invoice management, financing operations
2. **Frontend (React)**: User interface for SMEs, investors, and administrators
3. **Blockchain (Solidity)**: Smart contracts for invoice tokenization and escrow
4. **AI Service (Python)**: Machine learning services for credit scoring, risk assessment, fraud detection
5. **Database (PostgreSQL)**: Primary data storage
6. **Cache (Redis)**: Session management and caching

### Key Features

- **For SMEs**: Invoice upload, financing requests, automated verification
- **For Investors**: Investment opportunities, portfolio management, returns tracking
- **AI-Powered**: Credit scoring, fraud detection, risk assessment
- **Blockchain**: Invoice tokenization, transparent transactions, automated escrow
- **Security**: JWT authentication, role-based access control

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Token refresh

### User Management
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `GET /api/v1/users/stats` - Get user statistics

### Invoices
- `GET /api/v1/invoices` - List user invoices
- `POST /api/v1/invoices` - Create new invoice
- `GET /api/v1/invoices/:id` - Get invoice details
- `POST /api/v1/invoices/:id/verify` - Verify invoice

### Financing
- `GET /api/v1/financing/requests` - List financing requests
- `POST /api/v1/financing/requests` - Create financing request
- `GET /api/v1/financing/opportunities` - Investment opportunities
- `POST /api/v1/financing/invest` - Make investment

### Blockchain
- `POST /api/v1/blockchain/tokenize-invoice` - Tokenize invoice as NFT
- `GET /api/v1/blockchain/transactions/:hash` - Get transaction details

### AI Services
- `POST /api/v1/ai/credit-score` - Calculate credit score
- `POST /api/v1/ai/risk-assessment` - Assess investment risk
- `POST /api/v1/ai/fraud-detection` - Detect fraud

## Testing

### Backend Tests
```bash
cd backend
go test ./...
```

### Frontend Tests
```bash
cd frontend
npm test
```

### Smart Contract Tests
```bash
cd blockchain
npx hardhat test
```

## Deployment

### Production Environment Variables

Create a `.env` file with:
```
DATABASE_URL=postgresql://user:password@host:5432/dbname
JWT_SECRET=your-production-jwt-secret
ETHEREUM_RPC=https://mainnet.infura.io/v3/your-project-id
CONTRACT_ADDRESS=deployed-contract-address
AI_MODEL_ENDPOINT=https://your-ai-service.com/api/ml
```

### Docker Production Deployment
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Ensure PostgreSQL is running
   - Check connection string and credentials

2. **Smart Contract Deployment Failed**
   - Ensure Hardhat node is running
   - Check account balance for gas fees

3. **Frontend Not Loading**
   - Check if backend API is accessible
   - Verify CORS configuration

4. **AI Service Errors**
   - Check Python dependencies
   - Ensure FastAPI service is running

### Logs
- Backend logs: `docker-compose logs backend`
- Frontend logs: `docker-compose logs frontend`
- Database logs: `docker-compose logs postgres`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details
