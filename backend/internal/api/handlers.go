package api

import (
	"net/http"
	"strconv"

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
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Update invoice endpoint"})
}

func (s *Server) deleteInvoice(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Delete invoice endpoint"})
}

func (s *Server) verifyInvoice(c *gin.Context) {
	// Implementation placeholder - would use AI service
	c.JSON(http.StatusOK, gin.H{"message": "Invoice verification endpoint"})
}

func (s *Server) uploadInvoiceDocument(c *gin.Context) {
	// Implementation placeholder - file upload
	c.JSON(http.StatusOK, gin.H{"message": "Upload invoice document endpoint"})
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
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Get financing request endpoint"})
}

func (s *Server) updateFinancingRequest(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Update financing request endpoint"})
}

func (s *Server) approveFinancingRequest(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Approve financing request endpoint"})
}

func (s *Server) rejectFinancingRequest(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Reject financing request endpoint"})
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
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Create investment endpoint"})
}

func (s *Server) getUserInvestments(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Get user investments endpoint"})
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
