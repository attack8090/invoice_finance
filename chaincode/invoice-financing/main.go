package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// InvoiceFinancingContract provides functions for managing invoice financing
type InvoiceFinancingContract struct {
	contractapi.Contract
}

// Invoice represents a tokenized invoice
type Invoice struct {
	ID               string    `json:"id"`
	InvoiceNumber    string    `json:"invoice_number"`
	SMEAddress       string    `json:"sme_address"`
	InvoiceAmount    float64   `json:"invoice_amount"`
	DueDate          time.Time `json:"due_date"`
	IssueDate        time.Time `json:"issue_date"`
	CustomerName     string    `json:"customer_name"`
	Description      string    `json:"description"`
	IsVerified       bool      `json:"is_verified"`
	IsFinanced       bool      `json:"is_financed"`
	RiskLevel        string    `json:"risk_level"` // low, medium, high
	DocumentHash     string    `json:"document_hash"`
	TokenizedAt      time.Time `json:"tokenized_at"`
	Status           string    `json:"status"` // pending, verified, financed, paid, overdue
}

// FinancingRequest represents a financing request for an invoice
type FinancingRequest struct {
	ID              string    `json:"id"`
	InvoiceID       string    `json:"invoice_id"`
	SMEAddress      string    `json:"sme_address"`
	RequestedAmount float64   `json:"requested_amount"`
	InterestRate    float64   `json:"interest_rate"`
	FinancingFee    float64   `json:"financing_fee"`
	NetAmount       float64   `json:"net_amount"`
	RepaymentAmount float64   `json:"repayment_amount"`
	Status          string    `json:"status"` // pending, approved, funded, completed, rejected
	CreatedAt       time.Time `json:"created_at"`
	DueDate         time.Time `json:"due_date"`
	RiskLevel       string    `json:"risk_level"`
}

// Investment represents an investment in a financing request
type Investment struct {
	ID                 string    `json:"id"`
	FinancingRequestID string    `json:"financing_request_id"`
	InvestorAddress    string    `json:"investor_address"`
	Amount             float64   `json:"amount"`
	ExpectedReturn     float64   `json:"expected_return"`
	ActualReturn       float64   `json:"actual_return"`
	Status             string    `json:"status"` // pending, active, completed, defaulted
	InvestmentDate     time.Time `json:"investment_date"`
	MaturityDate       time.Time `json:"maturity_date"`
	ReturnDate         time.Time `json:"return_date"`
}

// TokenizeInvoice creates a new tokenized invoice on the ledger
func (c *InvoiceFinancingContract) TokenizeInvoice(ctx contractapi.TransactionContextInterface, invoiceData string) (*Invoice, error) {
	var invoice Invoice
	err := json.Unmarshal([]byte(invoiceData), &invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice data: %v", err)
	}

	// Validate invoice data
	if invoice.InvoiceNumber == "" {
		return nil, fmt.Errorf("invoice number is required")
	}
	if invoice.InvoiceAmount <= 0 {
		return nil, fmt.Errorf("invoice amount must be greater than 0")
	}
	if invoice.SMEAddress == "" {
		return nil, fmt.Errorf("SME address is required")
	}

	// Check if invoice already exists
	existingInvoice, err := c.GetInvoiceByNumber(ctx, invoice.InvoiceNumber)
	if err == nil && existingInvoice != nil {
		return nil, fmt.Errorf("invoice with number %s already exists", invoice.InvoiceNumber)
	}

	// Generate unique ID and set timestamps
	invoice.ID = ctx.GetStub().GetTxID()
	invoice.TokenizedAt = time.Now()
	invoice.Status = "pending"
	invoice.IsVerified = false
	invoice.IsFinanced = false

	// Store invoice on ledger
	invoiceJSON, err := json.Marshal(invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice: %v", err)
	}

	err = ctx.GetStub().PutState(invoice.ID, invoiceJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to put invoice to world state: %v", err)
	}

	// Create index for invoice number lookup
	err = ctx.GetStub().PutState("invoice_number_"+invoice.InvoiceNumber, []byte(invoice.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice number index: %v", err)
	}

	// Emit event
	eventPayload := map[string]interface{}{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"sme_address":    invoice.SMEAddress,
		"amount":         invoice.InvoiceAmount,
		"tokenized_at":   invoice.TokenizedAt,
	}
	eventData, _ := json.Marshal(eventPayload)
	ctx.GetStub().SetEvent("InvoiceTokenized", eventData)

	return &invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (c *InvoiceFinancingContract) GetInvoice(ctx contractapi.TransactionContextInterface, invoiceID string) (*Invoice, error) {
	invoiceJSON, err := ctx.GetStub().GetState(invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if invoiceJSON == nil {
		return nil, fmt.Errorf("invoice %s does not exist", invoiceID)
	}

	var invoice Invoice
	err = json.Unmarshal(invoiceJSON, &invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice: %v", err)
	}

	return &invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number
func (c *InvoiceFinancingContract) GetInvoiceByNumber(ctx contractapi.TransactionContextInterface, invoiceNumber string) (*Invoice, error) {
	// Get invoice ID from index
	invoiceIDBytes, err := ctx.GetStub().GetState("invoice_number_" + invoiceNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to read invoice number index: %v", err)
	}
	if invoiceIDBytes == nil {
		return nil, fmt.Errorf("invoice with number %s does not exist", invoiceNumber)
	}

	invoiceID := string(invoiceIDBytes)
	return c.GetInvoice(ctx, invoiceID)
}

// VerifyInvoice marks an invoice as verified
func (c *InvoiceFinancingContract) VerifyInvoice(ctx contractapi.TransactionContextInterface, invoiceID string, verified bool) error {
	invoice, err := c.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	invoice.IsVerified = verified
	if verified {
		invoice.Status = "verified"
	}

	invoiceJSON, err := json.Marshal(invoice)
	if err != nil {
		return fmt.Errorf("failed to marshal invoice: %v", err)
	}

	err = ctx.GetStub().PutState(invoiceID, invoiceJSON)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %v", err)
	}

	// Emit event
	eventPayload := map[string]interface{}{
		"invoice_id": invoiceID,
		"verified":   verified,
	}
	eventData, _ := json.Marshal(eventPayload)
	ctx.GetStub().SetEvent("InvoiceVerified", eventData)

	return nil
}

// CreateFinancingRequest creates a new financing request
func (c *InvoiceFinancingContract) CreateFinancingRequest(ctx contractapi.TransactionContextInterface, requestData string) (*FinancingRequest, error) {
	var request FinancingRequest
	err := json.Unmarshal([]byte(requestData), &request)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal financing request data: %v", err)
	}

	// Validate request
	invoice, err := c.GetInvoice(ctx, request.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %v", err)
	}

	if !invoice.IsVerified {
		return nil, fmt.Errorf("invoice must be verified before financing")
	}

	if invoice.IsFinanced {
		return nil, fmt.Errorf("invoice is already financed")
	}

	if request.RequestedAmount > invoice.InvoiceAmount {
		return nil, fmt.Errorf("requested amount exceeds invoice amount")
	}

	// Generate unique ID and set fields
	request.ID = ctx.GetStub().GetTxID()
	request.CreatedAt = time.Now()
	request.Status = "pending"
	request.RepaymentAmount = request.RequestedAmount + (request.RequestedAmount * request.InterestRate / 100)

	// Store financing request
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal financing request: %v", err)
	}

	err = ctx.GetStub().PutState("financing_request_"+request.ID, requestJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to store financing request: %v", err)
	}

	// Emit event
	eventPayload := map[string]interface{}{
		"request_id":       request.ID,
		"invoice_id":       request.InvoiceID,
		"sme_address":      request.SMEAddress,
		"requested_amount": request.RequestedAmount,
		"interest_rate":    request.InterestRate,
	}
	eventData, _ := json.Marshal(eventPayload)
	ctx.GetStub().SetEvent("FinancingRequestCreated", eventData)

	return &request, nil
}

// MakeInvestment creates an investment in a financing request
func (c *InvoiceFinancingContract) MakeInvestment(ctx contractapi.TransactionContextInterface, investmentData string) (*Investment, error) {
	var investment Investment
	err := json.Unmarshal([]byte(investmentData), &investment)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal investment data: %v", err)
	}

	// Get financing request
	requestJSON, err := ctx.GetStub().GetState("financing_request_" + investment.FinancingRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to read financing request: %v", err)
	}
	if requestJSON == nil {
		return nil, fmt.Errorf("financing request does not exist")
	}

	var request FinancingRequest
	err = json.Unmarshal(requestJSON, &request)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal financing request: %v", err)
	}

	if request.Status != "approved" {
		return nil, fmt.Errorf("financing request is not approved")
	}

	// Generate investment ID and set fields
	investment.ID = ctx.GetStub().GetTxID()
	investment.InvestmentDate = time.Now()
	investment.Status = "active"
	investment.MaturityDate = request.DueDate

	// Calculate expected return
	proportion := investment.Amount / request.RequestedAmount
	investment.ExpectedReturn = request.RepaymentAmount * proportion

	// Store investment
	investmentJSON, err := json.Marshal(investment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal investment: %v", err)
	}

	err = ctx.GetStub().PutState("investment_"+investment.ID, investmentJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to store investment: %v", err)
	}

	// Emit event
	eventPayload := map[string]interface{}{
		"investment_id":         investment.ID,
		"financing_request_id":  investment.FinancingRequestID,
		"investor_address":      investment.InvestorAddress,
		"amount":                investment.Amount,
		"expected_return":       investment.ExpectedReturn,
	}
	eventData, _ := json.Marshal(eventPayload)
	ctx.GetStub().SetEvent("InvestmentMade", eventData)

	return &investment, nil
}

// CompleteFinancing marks an invoice as financed when fully funded
func (c *InvoiceFinancingContract) CompleteFinancing(ctx contractapi.TransactionContextInterface, requestID string) error {
	// Get financing request
	requestJSON, err := ctx.GetStub().GetState("financing_request_" + requestID)
	if err != nil {
		return fmt.Errorf("failed to read financing request: %v", err)
	}
	if requestJSON == nil {
		return fmt.Errorf("financing request does not exist")
	}

	var request FinancingRequest
	err = json.Unmarshal(requestJSON, &request)
	if err != nil {
		return fmt.Errorf("failed to unmarshal financing request: %v", err)
	}

	// Update financing request status
	request.Status = "funded"

	requestJSON, err = json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal financing request: %v", err)
	}

	err = ctx.GetStub().PutState("financing_request_"+requestID, requestJSON)
	if err != nil {
		return fmt.Errorf("failed to update financing request: %v", err)
	}

	// Update invoice status
	invoice, err := c.GetInvoice(ctx, request.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %v", err)
	}

	invoice.IsFinanced = true
	invoice.Status = "financed"

	invoiceJSON, err := json.Marshal(invoice)
	if err != nil {
		return fmt.Errorf("failed to marshal invoice: %v", err)
	}

	err = ctx.GetStub().PutState(request.InvoiceID, invoiceJSON)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %v", err)
	}

	// Emit event
	eventPayload := map[string]interface{}{
		"request_id": requestID,
		"invoice_id": request.InvoiceID,
		"status":     "funded",
	}
	eventData, _ := json.Marshal(eventPayload)
	ctx.GetStub().SetEvent("FinancingCompleted", eventData)

	return nil
}

// ProcessRepayment handles loan repayment and investor returns
func (c *InvoiceFinancingContract) ProcessRepayment(ctx contractapi.TransactionContextInterface, requestID string, repaymentAmount float64) error {
	// Get financing request
	requestJSON, err := ctx.GetStub().GetState("financing_request_" + requestID)
	if err != nil {
		return fmt.Errorf("failed to read financing request: %v", err)
	}
	if requestJSON == nil {
		return fmt.Errorf("financing request does not exist")
	}

	var request FinancingRequest
	err = json.Unmarshal(requestJSON, &request)
	if err != nil {
		return fmt.Errorf("failed to unmarshal financing request: %v", err)
	}

	if request.Status != "funded" {
		return fmt.Errorf("financing request is not in funded status")
	}

	if repaymentAmount < request.RepaymentAmount {
		return fmt.Errorf("insufficient repayment amount")
	}

	// Update request status
	request.Status = "completed"

	requestJSON, err = json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal financing request: %v", err)
	}

	err = ctx.GetStub().PutState("financing_request_"+requestID, requestJSON)
	if err != nil {
		return fmt.Errorf("failed to update financing request: %v", err)
	}

	// Update invoice status
	invoice, err := c.GetInvoice(ctx, request.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %v", err)
	}

	invoice.Status = "paid"

	invoiceJSON, err := json.Marshal(invoice)
	if err != nil {
		return fmt.Errorf("failed to marshal invoice: %v", err)
	}

	err = ctx.GetStub().PutState(request.InvoiceID, invoiceJSON)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %v", err)
	}

	// Emit event
	eventPayload := map[string]interface{}{
		"request_id":       requestID,
		"repayment_amount": repaymentAmount,
		"completed_at":     time.Now(),
	}
	eventData, _ := json.Marshal(eventPayload)
	ctx.GetStub().SetEvent("RepaymentProcessed", eventData)

	return nil
}

// GetFinancingRequest retrieves a financing request by ID
func (c *InvoiceFinancingContract) GetFinancingRequest(ctx contractapi.TransactionContextInterface, requestID string) (*FinancingRequest, error) {
	requestJSON, err := ctx.GetStub().GetState("financing_request_" + requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to read financing request: %v", err)
	}
	if requestJSON == nil {
		return nil, fmt.Errorf("financing request %s does not exist", requestID)
	}

	var request FinancingRequest
	err = json.Unmarshal(requestJSON, &request)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal financing request: %v", err)
	}

	return &request, nil
}

// GetInvestment retrieves an investment by ID
func (c *InvoiceFinancingContract) GetInvestment(ctx contractapi.TransactionContextInterface, investmentID string) (*Investment, error) {
	investmentJSON, err := ctx.GetStub().GetState("investment_" + investmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to read investment: %v", err)
	}
	if investmentJSON == nil {
		return nil, fmt.Errorf("investment %s does not exist", investmentID)
	}

	var investment Investment
	err = json.Unmarshal(investmentJSON, &investment)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal investment: %v", err)
	}

	return &investment, nil
}

// GetInvoiceHistory returns the transaction history of an invoice
func (c *InvoiceFinancingContract) GetInvoiceHistory(ctx contractapi.TransactionContextInterface, invoiceID string) ([]map[string]interface{}, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history for invoice %s: %v", invoiceID, err)
	}
	defer resultsIterator.Close()

	var history []map[string]interface{}
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next history item: %v", err)
		}

		var invoice Invoice
		err = json.Unmarshal(response.Value, &invoice)
		if err != nil {
			continue // Skip invalid entries
		}

		historyItem := map[string]interface{}{
			"tx_id":     response.TxId,
			"timestamp": response.Timestamp,
			"is_delete": response.IsDelete,
			"invoice":   invoice,
		}
		history = append(history, historyItem)
	}

	return history, nil
}

// QueryInvoicesByRange performs a range query to find invoices
func (c *InvoiceFinancingContract) QueryInvoicesByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*Invoice, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var invoices []*Invoice
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next query result: %v", err)
		}

		var invoice Invoice
		err = json.Unmarshal(queryResponse.Value, &invoice)
		if err != nil {
			continue // Skip invalid entries
		}

		invoices = append(invoices, &invoice)
	}

	return invoices, nil
}

// QueryInvoicesByStatus finds invoices with a specific status
func (c *InvoiceFinancingContract) QueryInvoicesByStatus(ctx contractapi.TransactionContextInterface, status string) ([]*Invoice, error) {
	queryString := fmt.Sprintf(`{"selector":{"status":"%s"}}`, status)
	
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer resultsIterator.Close()

	var invoices []*Invoice
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next query result: %v", err)
		}

		var invoice Invoice
		err = json.Unmarshal(queryResponse.Value, &invoice)
		if err != nil {
			continue // Skip invalid entries
		}

		invoices = append(invoices, &invoice)
	}

	return invoices, nil
}

// HealthCheck returns the health status of the chaincode
func (c *InvoiceFinancingContract) HealthCheck(ctx contractapi.TransactionContextInterface) string {
	return "OK"
}

// InitLedger initializes the ledger with sample data (for testing)
func (c *InvoiceFinancingContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// This function can be used to initialize the ledger with sample data
	// In production, this might be empty or contain initial configuration
	log.Println("Invoice Financing Chaincode initialized successfully")
	return nil
}

func main() {
	invoiceContract := new(InvoiceFinancingContract)

	cc, err := contractapi.NewChaincode(invoiceContract)
	if err != nil {
		log.Fatalf("Error creating invoice financing chaincode: %v", err)
	}

	if err := cc.Start(); err != nil {
		log.Fatalf("Error starting invoice financing chaincode: %v", err)
	}
}
