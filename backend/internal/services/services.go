package services

import (
	"context"
	"fmt"
	"time"

	"invoice-financing-platform/internal/database"
	"invoice-financing-platform/internal/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// InvoiceService handles invoice-related operations
type InvoiceService struct {
	db *database.MongoDB
}

func NewInvoiceService(db *database.MongoDB) *InvoiceService {
	return &InvoiceService{db: db}
}

func (s *InvoiceService) Create(invoice *models.Invoice) error {
	invoice.UUID = uuid.New()
	invoice.CreatedAt = time.Now()
	invoice.UpdatedAt = time.Now()
	
	collection := s.db.Database.Collection("invoices")
	_, err := collection.InsertOne(context.Background(), invoice)
	return err
}

func (s *InvoiceService) GetByUserID(userID uuid.UUID) ([]models.Invoice, error) {
	var invoices []models.Invoice
	collection := s.db.Database.Collection("invoices")
	
	filter := bson.M{"user_id": userID}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	err = cursor.All(context.Background(), &invoices)
	return invoices, err
}

func (s *InvoiceService) GetByID(id uuid.UUID) (*models.Invoice, error) {
	var invoice models.Invoice
	collection := s.db.Database.Collection("invoices")
	
	filter := bson.M{"uuid": id}
	err := collection.FindOne(context.Background(), filter).Decode(&invoice)
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (s *InvoiceService) Update(invoice *models.Invoice) error {
	invoice.UpdatedAt = time.Now()
	collection := s.db.Database.Collection("invoices")
	
	filter := bson.M{"uuid": invoice.UUID}
	update := bson.M{"$set": invoice}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (s *InvoiceService) Delete(id uuid.UUID) error {
	collection := s.db.Database.Collection("invoices")
	
	// Soft delete by setting deleted_at
	filter := bson.M{"uuid": id}
	now := time.Now()
	update := bson.M{"$set": bson.M{"deleted_at": now}}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

// FinancingService handles financing request operations
type FinancingService struct {
	db *database.MongoDB
}

func NewFinancingService(db *database.MongoDB) *FinancingService {
	return &FinancingService{db: db}
}

func (s *FinancingService) CreateRequest(request *models.FinancingRequest) error {
	request.UUID = uuid.New()
	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()
	
	collection := s.db.Database.Collection("financing_requests")
	_, err := collection.InsertOne(context.Background(), request)
	return err
}

func (s *FinancingService) GetRequestsByUserID(userID uuid.UUID) ([]models.FinancingRequest, error) {
	var requests []models.FinancingRequest
	collection := s.db.Database.Collection("financing_requests")
	
	filter := bson.M{"user_id": userID}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	err = cursor.All(context.Background(), &requests)
	return requests, err
}

func (s *FinancingService) GetInvestmentOpportunities(limit int) ([]models.FinancingRequest, error) {
	var requests []models.FinancingRequest
	collection := s.db.Database.Collection("financing_requests")
	
	filter := bson.M{"status": models.FinancingStatusApproved}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	err = cursor.All(context.Background(), &requests)
	return requests, err
}

func (s *FinancingService) GetRequestByID(id uuid.UUID) (*models.FinancingRequest, error) {
	var request models.FinancingRequest
	collection := s.db.Database.Collection("financing_requests")
	
	filter := bson.M{"uuid": id}
	err := collection.FindOne(context.Background(), filter).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (s *FinancingService) CreateInvestment(investment *models.Investment) error {
	investment.UUID = uuid.New()
	investment.CreatedAt = time.Now()
	investment.UpdatedAt = time.Now()
	
	collection := s.db.Database.Collection("investments")
	_, err := collection.InsertOne(context.Background(), investment)
	return err
}

func (s *FinancingService) GetInvestmentsByUserID(userID uuid.UUID) ([]models.Investment, error) {
	var investments []models.Investment
	collection := s.db.Database.Collection("investments")
	
	filter := bson.M{"investor_id": userID}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	err = cursor.All(context.Background(), &investments)
	return investments, err
}

func (s *FinancingService) UpdateRequestStatus(requestID uuid.UUID, status models.FinancingStatus) error {
	collection := s.db.Database.Collection("financing_requests")
	
	filter := bson.M{"uuid": requestID}
	update := bson.M{"$set": bson.M{
		"status": status,
		"updated_at": time.Now(),
	}}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
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
