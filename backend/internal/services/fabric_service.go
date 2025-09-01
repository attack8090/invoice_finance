package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"invoice-financing-platform/internal/config"
	"invoice-financing-platform/internal/models"

	"github.com/google/uuid"
)

// FabricService handles interactions with Hyperledger Fabric
type FabricService struct {
	ledgerServiceURL string
	channelName      string
	chaincodeName    string
	client           *http.Client
}

// NewFabricService creates a new FabricService instance
func NewFabricService(cfg *config.Config) *FabricService {
	return &FabricService{
		ledgerServiceURL: cfg.FabricLedgerServiceURL,
		channelName:      cfg.FabricChannelName,
		chaincodeName:    cfg.FabricChaincodeName,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FabricInvoice represents invoice data for Fabric chaincode
type FabricInvoice struct {
	InvoiceNumber string    `json:"invoice_number"`
	SMEAddress    string    `json:"sme_address"`
	InvoiceAmount float64   `json:"invoice_amount"`
	DueDate       time.Time `json:"due_date"`
	IssueDate     time.Time `json:"issue_date"`
	CustomerName  string    `json:"customer_name"`
	Description   string    `json:"description"`
	RiskLevel     string    `json:"risk_level"`
	DocumentHash  string    `json:"document_hash"`
}

// FabricFinancingRequest represents financing request data for Fabric
type FabricFinancingRequest struct {
	InvoiceID       string    `json:"invoice_id"`
	SMEAddress      string    `json:"sme_address"`
	RequestedAmount float64   `json:"requested_amount"`
	InterestRate    float64   `json:"interest_rate"`
	FinancingFee    float64   `json:"financing_fee"`
	NetAmount       float64   `json:"net_amount"`
	DueDate         time.Time `json:"due_date"`
	RiskLevel       string    `json:"risk_level"`
}

// FabricInvestment represents investment data for Fabric
type FabricInvestment struct {
	FinancingRequestID string  `json:"financing_request_id"`
	InvestorAddress    string  `json:"investor_address"`
	Amount             float64 `json:"amount"`
}

// TokenizeInvoice creates a tokenized invoice on Hyperledger Fabric
func (s *FabricService) TokenizeInvoice(invoice *models.Invoice, userWallet string) (string, error) {
	fabricInvoice := FabricInvoice{
		InvoiceNumber: invoice.InvoiceNumber,
		SMEAddress:    userWallet,
		InvoiceAmount: invoice.InvoiceAmount,
		DueDate:       invoice.DueDate,
		IssueDate:     invoice.IssueDate,
		CustomerName:  invoice.CustomerName,
		Description:   invoice.Description,
		RiskLevel:     string(invoice.VerificationStatus),
		DocumentHash:  fmt.Sprintf("hash_%s", invoice.InvoiceNumber), // In production, use actual document hash
	}

	requestBody := map[string]interface{}{
		"function": "TokenizeInvoice",
		"args":     []interface{}{fabricInvoice},
	}

	response, err := s.invokeChaincode(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to tokenize invoice: %v", err)
	}

	// Extract transaction ID from response
	if txID, ok := response["transaction_id"].(string); ok {
		return txID, nil
	}

	return "", fmt.Errorf("failed to get transaction ID from response")
}

// VerifyInvoice marks an invoice as verified on the blockchain
func (s *FabricService) VerifyInvoice(assetID string, verified bool) error {
	requestBody := map[string]interface{}{
		"function": "VerifyInvoice",
		"args":     []interface{}{assetID, verified},
	}

	_, err := s.invokeChaincode(requestBody)
	if err != nil {
		return fmt.Errorf("failed to verify invoice on blockchain: %v", err)
	}

	return nil
}

// CreateFinancingRequest creates a financing request on the blockchain
func (s *FabricService) CreateFinancingRequest(request *models.FinancingRequest, invoiceAssetID string, userWallet string) (string, error) {
	fabricRequest := FabricFinancingRequest{
		InvoiceID:       invoiceAssetID,
		SMEAddress:      userWallet,
		RequestedAmount: request.RequestedAmount,
		InterestRate:    request.InterestRate,
		FinancingFee:    request.FinancingFee,
		NetAmount:       request.NetAmount,
		RiskLevel:       string(request.RiskLevel),
	}

	requestBody := map[string]interface{}{
		"function": "CreateFinancingRequest",
		"args":     []interface{}{fabricRequest},
	}

	response, err := s.invokeChaincode(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create financing request: %v", err)
	}

	// Extract transaction ID from response
	if txID, ok := response["transaction_id"].(string); ok {
		return txID, nil
	}

	return "", fmt.Errorf("failed to get transaction ID from response")
}

// MakeInvestment records an investment on the blockchain
func (s *FabricService) MakeInvestment(investment *models.Investment, requestAssetID string, investorWallet string) (string, error) {
	fabricInvestment := FabricInvestment{
		FinancingRequestID: requestAssetID,
		InvestorAddress:    investorWallet,
		Amount:             investment.Amount,
	}

	requestBody := map[string]interface{}{
		"function": "MakeInvestment",
		"args":     []interface{}{fabricInvestment},
	}

	response, err := s.invokeChaincode(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to make investment: %v", err)
	}

	// Extract transaction ID from response
	if txID, ok := response["transaction_id"].(string); ok {
		return txID, nil
	}

	return "", fmt.Errorf("failed to get transaction ID from response")
}

// CompleteFinancing marks financing as complete on the blockchain
func (s *FabricService) CompleteFinancing(requestAssetID string) error {
	requestBody := map[string]interface{}{
		"function": "CompleteFinancing",
		"args":     []interface{}{requestAssetID},
	}

	_, err := s.invokeChaincode(requestBody)
	if err != nil {
		return fmt.Errorf("failed to complete financing: %v", err)
	}

	return nil
}

// ProcessRepayment records a repayment on the blockchain
func (s *FabricService) ProcessRepayment(requestAssetID string, repaymentAmount float64) error {
	requestBody := map[string]interface{}{
		"function": "ProcessRepayment",
		"args":     []interface{}{requestAssetID, repaymentAmount},
	}

	_, err := s.invokeChaincode(requestBody)
	if err != nil {
		return fmt.Errorf("failed to process repayment: %v", err)
	}

	return nil
}

// GetInvoiceFromBlockchain retrieves invoice data from the blockchain
func (s *FabricService) GetInvoiceFromBlockchain(assetID string) (map[string]interface{}, error) {
	requestBody := map[string]interface{}{
		"function": "GetInvoice",
		"args":     []interface{}{assetID},
	}

	response, err := s.queryChaincode(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice from blockchain: %v", err)
	}

	return response, nil
}

// GetTransactionHistory retrieves the transaction history for an asset
func (s *FabricService) GetTransactionHistory(assetID string) ([]map[string]interface{}, error) {
	requestBody := map[string]interface{}{
		"function": "GetInvoiceHistory",
		"args":     []interface{}{assetID},
	}

	response, err := s.queryChaincode(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %v", err)
	}

	// Convert response to history array
	if history, ok := response["history"].([]interface{}); ok {
		var result []map[string]interface{}
		for _, item := range history {
			if historyItem, ok := item.(map[string]interface{}); ok {
				result = append(result, historyItem)
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("invalid history response format")
}

// HealthCheck checks the health of the Fabric network
func (s *FabricService) HealthCheck() error {
	requestBody := map[string]interface{}{
		"function": "HealthCheck",
		"args":     []interface{}{},
	}

	_, err := s.queryChaincode(requestBody)
	return err
}

// invokeChaincode submits a transaction to the chaincode
func (s *FabricService) invokeChaincode(requestBody map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/blockchain/chaincode/invoke", s.ledgerServiceURL)
	
	// Add chaincode and channel info to request
	requestBody["chaincode_name"] = s.chaincodeName
	requestBody["channel_name"] = s.channelName

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chaincode invocation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return result, nil
}

// queryChaincode queries the chaincode without committing a transaction
func (s *FabricService) queryChaincode(requestBody map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/blockchain/chaincode/query", s.ledgerServiceURL)
	
	// Add chaincode and channel info to request
	requestBody["chaincode_name"] = s.chaincodeName
	requestBody["channel_name"] = s.channelName

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chaincode query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return result, nil
}
