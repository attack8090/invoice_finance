package services

import (
	"fmt"

	"invoice-financing-platform/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InvoiceService handles invoice-related operations
type InvoiceService struct {
	db *gorm.DB
}

func NewInvoiceService(db *gorm.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

func (s *InvoiceService) Create(invoice *models.Invoice) error {
	return s.db.Create(invoice).Error
}

func (s *InvoiceService) GetByUserID(userID uuid.UUID) ([]models.Invoice, error) {
	var invoices []models.Invoice
	err := s.db.Preload("FinancingRequests").Where("user_id = ?", userID).Find(&invoices).Error
	return invoices, err
}

func (s *InvoiceService) GetByID(id uuid.UUID) (*models.Invoice, error) {
	var invoice models.Invoice
	err := s.db.Preload("User").Preload("FinancingRequests").First(&invoice, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (s *InvoiceService) Update(invoice *models.Invoice) error {
	return s.db.Save(invoice).Error
}

func (s *InvoiceService) Delete(id uuid.UUID) error {
	return s.db.Delete(&models.Invoice{}, "id = ?", id).Error
}

// FinancingService handles financing request operations
type FinancingService struct {
	db *gorm.DB
}

func NewFinancingService(db *gorm.DB) *FinancingService {
	return &FinancingService{db: db}
}

func (s *FinancingService) CreateRequest(request *models.FinancingRequest) error {
	return s.db.Create(request).Error
}

func (s *FinancingService) GetRequestsByUserID(userID uuid.UUID) ([]models.FinancingRequest, error) {
	var requests []models.FinancingRequest
	err := s.db.Preload("Invoice").Preload("Investments").
		Where("user_id = ?", userID).Find(&requests).Error
	return requests, err
}

func (s *FinancingService) GetInvestmentOpportunities(limit int) ([]models.FinancingRequest, error) {
	var requests []models.FinancingRequest
	err := s.db.Preload("Invoice").Preload("User").
		Where("status = ?", models.FinancingStatusApproved).
		Limit(limit).Find(&requests).Error
	return requests, err
}

func (s *FinancingService) GetRequestByID(id uuid.UUID) (*models.FinancingRequest, error) {
	var request models.FinancingRequest
	err := s.db.Preload("Invoice").Preload("User").Preload("Investments").
		First(&request, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (s *FinancingService) CreateInvestment(investment *models.Investment) error {
	return s.db.Create(investment).Error
}

func (s *FinancingService) GetInvestmentsByUserID(userID uuid.UUID) ([]models.Investment, error) {
	var investments []models.Investment
	err := s.db.Preload("FinancingRequest").Preload("FinancingRequest.Invoice").
		Where("investor_id = ?", userID).Find(&investments).Error
	return investments, err
}

func (s *FinancingService) UpdateRequestStatus(requestID uuid.UUID, status models.FinancingStatus) error {
	return s.db.Model(&models.FinancingRequest{}).
		Where("id = ?", requestID).
		Update("status", status).Error
}

// BlockchainService handles blockchain operations
type BlockchainService struct {
	rpcURL          string
	contractAddress string
}

func NewBlockchainService(rpcURL, contractAddress string) *BlockchainService {
	return &BlockchainService{
		rpcURL:          rpcURL,
		contractAddress: contractAddress,
	}
}

func (s *BlockchainService) TokenizeInvoice(invoiceID uuid.UUID) (string, error) {
	// Placeholder implementation
	// In production, this would interact with Ethereum smart contract
	txHash := fmt.Sprintf("0x%x", invoiceID)
	return txHash, nil
}

func (s *BlockchainService) VerifyTransaction(txHash string) (bool, error) {
	// Placeholder implementation
	return true, nil
}

// AIService handles AI/ML operations
type AIService struct {
	endpoint string
}

func NewAIService(endpoint string) *AIService {
	return &AIService{endpoint: endpoint}
}

func (s *AIService) CalculateCreditScore(userID uuid.UUID, data map[string]interface{}) (int, error) {
	// Placeholder implementation
	// In production, this would call ML service
	return 750, nil
}

func (s *AIService) AssessRisk(invoiceData map[string]interface{}) (models.RiskLevel, float64, error) {
	// Placeholder implementation
	return models.RiskLevelMedium, 0.35, nil
}

func (s *AIService) DetectFraud(invoiceData map[string]interface{}) (bool, float64, error) {
	// Placeholder implementation
	return false, 0.1, nil
}
