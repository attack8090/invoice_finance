package api

import (
	"net/http"
	"strconv"
	"time"

	"invoice-financing-platform/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// User handlers
func (s *Server) getUserProfile(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	user, err := s.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Remove password hash
	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

func (s *Server) updateUserProfile(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var updateData struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		CompanyName string `json:"company_name"`
		TaxID       string `json:"tax_id"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.FirstName = updateData.FirstName
	user.LastName = updateData.LastName
	user.CompanyName = updateData.CompanyName
	user.TaxID = updateData.TaxID

	if err := s.userService.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

func (s *Server) verifyUser(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	if err := s.userService.Verify(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User verification submitted"})
}

func (s *Server) getUserStats(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	stats, err := s.userService.GetStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Invoice handlers
func (s *Server) getInvoices(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	invoices, err := s.invoiceService.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invoices"})
		return
	}

	c.JSON(http.StatusOK, invoices)
}

func (s *Server) createInvoice(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var invoice models.Invoice
	if err := c.ShouldBindJSON(&invoice); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice.UserID = userID
	if err := s.invoiceService.Create(&invoice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invoice"})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

func (s *Server) getInvoice(c *gin.Context) {
	idParam := c.Param("id")
	invoiceID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	invoice, err := s.invoiceService.GetByID(invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func (s *Server) updateInvoice(c *gin.Context) {
	idParam := c.Param("id")
	invoiceID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	// Check if invoice exists and belongs to user
	existingInvoice, err := s.invoiceService.GetByID(invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	if existingInvoice.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var updateData struct {
		InvoiceNumber string  `json:"invoice_number"`
		CustomerName  string  `json:"customer_name"`
		CustomerEmail string  `json:"customer_email"`
		InvoiceAmount float64 `json:"invoice_amount"`
		Description   string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update invoice fields
	if updateData.InvoiceNumber != "" {
		existingInvoice.InvoiceNumber = updateData.InvoiceNumber
	}
	if updateData.CustomerName != "" {
		existingInvoice.CustomerName = updateData.CustomerName
	}
	if updateData.CustomerEmail != "" {
		existingInvoice.CustomerEmail = updateData.CustomerEmail
	}
	if updateData.InvoiceAmount > 0 {
		existingInvoice.InvoiceAmount = updateData.InvoiceAmount
	}
	if updateData.Description != "" {
		existingInvoice.Description = updateData.Description
	}

	if err := s.invoiceService.Update(existingInvoice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice"})
		return
	}

	c.JSON(http.StatusOK, existingInvoice)
}

func (s *Server) deleteInvoice(c *gin.Context) {
	idParam := c.Param("id")
	invoiceID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	// Check if invoice exists and belongs to user
	existingInvoice, err := s.invoiceService.GetByID(invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	if existingInvoice.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check if invoice has active financing requests
	requests, err := s.financingService.GetRequestsByInvoiceID(invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check financing requests"})
		return
	}
	for _, req := range requests {
		if req.Status == models.FinancingStatusPending || req.Status == models.FinancingStatusApproved || req.Status == models.FinancingStatusFunded {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete invoice with active financing requests"})
			return
		}
	}

	if err := s.invoiceService.Delete(invoiceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice deleted successfully"})
}

func (s *Server) verifyInvoice(c *gin.Context) {
	// Implementation placeholder - would use AI service
	c.JSON(http.StatusOK, gin.H{"message": "Invoice verification endpoint"})
}

func (s *Server) uploadInvoiceDocument(c *gin.Context) {
	idParam := c.Param("id")
	invoiceID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	// Check if invoice exists and belongs to user
	existingInvoice, err := s.invoiceService.GetByID(invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	if existingInvoice.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded or invalid file"})
		return
	}
	defer file.Close()

	// Validate file type
	allowedTypes := map[string]bool{
		"application/pdf":  true,
		"image/jpeg":       true,
		"image/png":        true,
		"image/jpg":        true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only PDF, JPEG, and PNG files are allowed"})
		return
	}

	// Validate file size (10MB max)
	maxSize := int64(10 << 20) // 10MB
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size too large. Maximum size is 10MB"})
		return
	}

	// Save file using file service
	fileName, err := s.fileService.SaveInvoiceDocument(file, header, invoiceID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Update invoice with document URL
	existingInvoice.DocumentURL = fileName
	if err := s.invoiceService.Update(existingInvoice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "File uploaded successfully",
		"document_url": fileName,
		"invoice":      existingInvoice,
	})
}

// Financing handlers
func (s *Server) getFinancingRequests(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	requests, err := s.financingService.GetRequestsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get financing requests"})
		return
	}

	c.JSON(http.StatusOK, requests)
}

func (s *Server) createFinancingRequest(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var request models.FinancingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.UserID = userID
	if err := s.financingService.CreateRequest(&request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create financing request"})
		return
	}

	c.JSON(http.StatusCreated, request)
}

func (s *Server) getFinancingRequest(c *gin.Context) {
	idParam := c.Param("id")
	requestID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	request, err := s.financingService.GetRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Financing request not found"})
		return
	}

	c.JSON(http.StatusOK, request)
}

func (s *Server) updateFinancingRequest(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Update financing request endpoint"})
}

func (s *Server) approveFinancingRequest(c *gin.Context) {
	idParam := c.Param("id")
	requestID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	userRole, _ := c.Get("user_role")
	if userRole != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	request, err := s.financingService.GetRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Financing request not found"})
		return
	}

	if request.Status != models.FinancingStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending requests can be approved"})
		return
	}

	if err := s.financingService.UpdateRequestStatus(requestID, models.FinancingStatusApproved); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Financing request approved successfully"})
}

func (s *Server) rejectFinancingRequest(c *gin.Context) {
	idParam := c.Param("id")
	requestID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	userRole, _ := c.Get("user_role")
	if userRole != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var rejectionData struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&rejectionData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request, err := s.financingService.GetRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Financing request not found"})
		return
	}

	if request.Status != models.FinancingStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending requests can be rejected"})
		return
	}

	if err := s.financingService.UpdateRequestStatus(requestID, models.FinancingStatusRejected); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Financing request rejected", "reason": rejectionData.Reason})
}

func (s *Server) getInvestmentOpportunities(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	opportunities, err := s.financingService.GetInvestmentOpportunities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get investment opportunities"})
		return
	}

	c.JSON(http.StatusOK, opportunities)
}

func (s *Server) createInvestment(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var investmentReq struct {
		FinancingRequestID uuid.UUID `json:"financing_request_id" binding:"required"`
		Amount             float64   `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&investmentReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get financing request to validate
	financingRequest, err := s.financingService.GetRequestByID(investmentReq.FinancingRequestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Financing request not found"})
		return
	}

	if financingRequest.Status != models.FinancingStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Financing request not approved for investment"})
		return
	}

	// Create investment
	investment := &models.Investment{
		FinancingRequestID: investmentReq.FinancingRequestID,
		InvestorID:         userID,
		Amount:             investmentReq.Amount,
		ExpectedReturn:     investmentReq.Amount * (1 + financingRequest.InterestRate/100),
		Status:             models.InvestmentStatusActive,
		InvestmentDate:     time.Now(),
		MaturityDate:       time.Now().AddDate(0, 0, 30), // 30 days from now
	}

	if err := s.financingService.CreateInvestment(investment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create investment"})
		return
	}

	c.JSON(http.StatusCreated, investment)
}

func (s *Server) getUserInvestments(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	investments, err := s.financingService.GetInvestmentsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user investments"})
		return
	}

	c.JSON(http.StatusOK, investments)
}

// Blockchain handlers
func (s *Server) tokenizeInvoice(c *gin.Context) {
	var request struct {
		InvoiceID uuid.UUID `json:"invoice_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txHash, err := s.blockchainService.TokenizeInvoice(request.InvoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to tokenize invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_hash": txHash,
		"message":         "Invoice tokenized successfully",
	})
}

func (s *Server) getBlockchainTransaction(c *gin.Context) {
	txHash := c.Param("hash")
	
	verified, err := s.blockchainService.VerifyTransaction(txHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_hash": txHash,
		"verified":        verified,
	})
}

func (s *Server) verifyTransaction(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Verify transaction endpoint"})
}

// AI handlers
func (s *Server) calculateCreditScore(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score, err := s.aiService.CalculateCreditScore(userID, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate credit score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credit_score": score})
}

func (s *Server) assessRisk(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	riskLevel, score, err := s.aiService.AssessRisk(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assess risk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"risk_level": riskLevel,
		"risk_score": score,
	})
}

func (s *Server) detectFraud(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isFraud, confidence, err := s.aiService.DetectFraud(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to detect fraud"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_fraud":   isFraud,
		"confidence": confidence,
	})
}

func (s *Server) verifyDocument(c *gin.Context) {
	// Implementation placeholder - document verification using AI/OCR
	c.JSON(http.StatusOK, gin.H{"message": "Verify document endpoint"})
}

// Admin handlers
func (s *Server) getAdminDashboard(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Admin dashboard endpoint"})
}

func (s *Server) getAllUsers(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Get all users endpoint"})
}

func (s *Server) getAllTransactions(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Get all transactions endpoint"})
}

func (s *Server) adminVerifyUser(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Admin verify user endpoint"})
}

func (s *Server) suspendUser(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Suspend user endpoint"})
}

// Analytics handlers
func (s *Server) getDashboardAnalytics(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Dashboard analytics endpoint"})
}

func (s *Server) getPortfolioAnalytics(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Portfolio analytics endpoint"})
}

func (s *Server) getMarketTrends(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Market trends endpoint"})
}
