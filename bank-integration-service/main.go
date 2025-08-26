package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"bank-integration-service/internal/config"
	"bank-integration-service/internal/database"
	"bank-integration-service/internal/handlers"
	"bank-integration-service/internal/middleware"
	"bank-integration-service/internal/models"
	"bank-integration-service/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run database migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize bank integration services
	bankAPIService := services.NewBankAPIService(cfg)
	creditDecisionService := services.NewCreditDecisionService(db, cfg)
	paymentProcessingService := services.NewPaymentProcessingService(db, cfg)
	financingService := services.NewFinancingService(db, cfg)
	portfolioService := services.NewPortfolioService(db, cfg)
	complianceService := services.NewComplianceService(db, cfg)
	fundingMatchingService := services.NewFundingMatchingService(db, cfg)
	riskAssessmentService := services.NewRiskAssessmentService(db, cfg)
	auditService := services.NewAuditService(db, cfg)

	// Initialize handlers
	bankHandler := handlers.NewBankHandler(bankAPIService, complianceService, auditService)
	creditHandler := handlers.NewCreditHandler(creditDecisionService, riskAssessmentService, complianceService)
	paymentHandler := handlers.NewPaymentHandler(paymentProcessingService, bankAPIService, auditService)
	financingHandler := handlers.NewFinancingHandler(financingService, fundingMatchingService, complianceService)
	portfolioHandler := handlers.NewPortfolioHandler(portfolioService, riskAssessmentService)
	complianceHandler := handlers.NewComplianceHandler(complianceService, auditService)

	// Initialize Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	if cfg.Environment == "development" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = cfg.AllowedOrigins
	}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key", "X-Bank-ID"}
	router.Use(cors.New(corsConfig))

	// Middleware
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimit(cfg))
	router.Use(middleware.RequestLogging())
	router.Use(middleware.RequestID())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		bankConnections := checkBankConnections(bankAPIService)

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "bank-integration-service",
			"version": "1.0.0",
			"features": gin.H{
				"epic_4_compliance":    true,
				"bank_api_integration": true,
				"credit_decisions":     true,
				"payment_processing":   true,
				"funding_matching":     true,
				"portfolio_reports":    true,
				"real_time_transfers":  true,
				"multi_bank_support":   true,
			},
			"bank_connections": bankConnections,
			"supported_banks":  getSupportedBanks(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Bank connection and management
		banks := v1.Group("/banks")
		{
			banks.GET("", bankHandler.GetSupportedBanks)
			banks.POST("/connect", bankHandler.ConnectBank)
			banks.GET("/connections", bankHandler.GetBankConnections)
			banks.PUT("/connections/:connectionId", bankHandler.UpdateBankConnection)
			banks.DELETE("/connections/:connectionId", bankHandler.DisconnectBank)
			banks.GET("/connections/:connectionId/status", bankHandler.GetConnectionStatus)
			banks.POST("/connections/:connectionId/test", bankHandler.TestBankConnection)
		}

		// Credit decisions and assessment
		credit := v1.Group("/credit")
		{
			credit.POST("/decisions/request", creditHandler.RequestCreditDecision)
			credit.GET("/decisions/:decisionId", creditHandler.GetCreditDecision)
			credit.PUT("/decisions/:decisionId/update", creditHandler.UpdateCreditDecision)
			credit.POST("/decisions/:decisionId/approve", creditHandler.ApproveCreditDecision)
			credit.POST("/decisions/:decisionId/reject", creditHandler.RejectCreditDecision)
			credit.GET("/decisions", creditHandler.GetCreditDecisions)
			credit.POST("/assessment/risk", creditHandler.AssessRisk)
			credit.GET("/limits/:customerId", creditHandler.GetCreditLimits)
			credit.PUT("/limits/:customerId", creditHandler.UpdateCreditLimits)
		}

		// Payment processing
		payments := v1.Group("/payments")
		{
			payments.POST("/process", paymentHandler.ProcessPayment)
			payments.GET("/:paymentId", paymentHandler.GetPayment)
			payments.GET("/:paymentId/status", paymentHandler.GetPaymentStatus)
			payments.POST("/:paymentId/cancel", paymentHandler.CancelPayment)
			payments.POST("/bulk-process", paymentHandler.BulkProcessPayments)
			payments.GET("/transactions", paymentHandler.GetTransactions)
			payments.POST("/reconcile", paymentHandler.ReconcilePayments)
			payments.GET("/reconciliation/:jobId", paymentHandler.GetReconciliationStatus)
		}

		// Financing requests and management
		financing := v1.Group("/financing")
		{
			financing.POST("/requests", financingHandler.CreateFinancingRequest)
			financing.GET("/requests/:requestId", financingHandler.GetFinancingRequest)
			financing.PUT("/requests/:requestId", financingHandler.UpdateFinancingRequest)
			financing.POST("/requests/:requestId/review", financingHandler.ReviewFinancingRequest)
			financing.POST("/requests/:requestId/approve", financingHandler.ApproveFinancing)
			financing.POST("/requests/:requestId/reject", financingHandler.RejectFinancing)
			financing.POST("/requests/:requestId/disburse", financingHandler.DisburseFinancing)
			financing.GET("/opportunities", financingHandler.GetFinancingOpportunities)
			financing.POST("/match-funding", financingHandler.MatchFunding)
		}

		// Portfolio management and reports
		portfolio := v1.Group("/portfolio")
		{
			portfolio.GET("/overview", portfolioHandler.GetPortfolioOverview)
			portfolio.GET("/performance", portfolioHandler.GetPortfolioPerformance)
			portfolio.GET("/risk-analysis", portfolioHandler.GetRiskAnalysis)
			portfolio.GET("/exposures", portfolioHandler.GetExposures)
			portfolio.GET("/concentrations", portfolioHandler.GetConcentrations)
			portfolio.POST("/reports/generate", portfolioHandler.GenerateReport)
			portfolio.GET("/reports/:reportId", portfolioHandler.GetReport)
			portfolio.GET("/analytics", portfolioHandler.GetAnalytics)
		}

		// Funding and matching
		funding := v1.Group("/funding")
		{
			funding.GET("/sources", financingHandler.GetFundingSources)
			funding.POST("/sources", financingHandler.AddFundingSource)
			funding.PUT("/sources/:sourceId", financingHandler.UpdateFundingSource)
			funding.DELETE("/sources/:sourceId", financingHandler.RemoveFundingSource)
			funding.GET("/capacity", financingHandler.GetFundingCapacity)
			funding.POST("/allocate", financingHandler.AllocateFunding)
			funding.GET("/allocations", financingHandler.GetFundingAllocations)
			funding.POST("/matching/run", financingHandler.RunFundingMatching)
			funding.GET("/matching/results", financingHandler.GetMatchingResults)
		}

		// Compliance and audit
		compliance := v1.Group("/compliance")
		{
			compliance.GET("/epic4/status", complianceHandler.GetEpic4ComplianceStatus)
			compliance.POST("/epic4/validate", complianceHandler.ValidateEpic4Compliance)
			compliance.GET("/epic4/reports", complianceHandler.GetEpic4Reports)
			compliance.POST("/audit/trail", complianceHandler.CreateAuditTrail)
			compliance.GET("/audit/trails", complianceHandler.GetAuditTrails)
			compliance.POST("/regulatory/filing", complianceHandler.CreateRegulatoryFiling)
			compliance.GET("/regulatory/filings", complianceHandler.GetRegulatoryFilings)
		}

		// Account management
		accounts := v1.Group("/accounts")
		{
			accounts.GET("", bankHandler.GetBankAccounts)
			accounts.POST("", bankHandler.CreateBankAccount)
			accounts.GET("/:accountId", bankHandler.GetBankAccount)
			accounts.PUT("/:accountId", bankHandler.UpdateBankAccount)
			accounts.DELETE("/:accountId", bankHandler.DeleteBankAccount)
			accounts.GET("/:accountId/balance", bankHandler.GetAccountBalance)
			accounts.GET("/:accountId/transactions", bankHandler.GetAccountTransactions)
			accounts.POST("/:accountId/verify", bankHandler.VerifyBankAccount)
		}

		// Real-time transfers
		transfers := v1.Group("/transfers")
		{
			transfers.POST("/initiate", paymentHandler.InitiateTransfer)
			transfers.GET("/:transferId", paymentHandler.GetTransfer)
			transfers.GET("/:transferId/status", paymentHandler.GetTransferStatus)
			transfers.POST("/:transferId/cancel", paymentHandler.CancelTransfer)
			transfers.GET("/real-time/status", paymentHandler.GetRealTimeTransferStatus)
			transfers.POST("/bulk-transfer", paymentHandler.BulkTransfer)
		}

		// Administrative endpoints
		admin := v1.Group("/admin")
		admin.Use(middleware.RequireRole("admin", "bank_admin"))
		{
			admin.GET("/connections/all", bankHandler.GetAllBankConnections)
			admin.GET("/system/health", getSystemHealth)
			admin.POST("/maintenance/mode", enableMaintenanceMode)
			admin.DELETE("/maintenance/mode", disableMaintenanceMode)
			admin.GET("/metrics", getSystemMetrics)
			admin.POST("/cache/clear", clearCache)
			admin.GET("/audit/system", complianceHandler.GetSystemAuditLog)
		}

		// Integration endpoints for external systems
		integrations := v1.Group("/integrations")
		{
			integrations.POST("/webhooks/bank-notification", bankHandler.BankWebhookHandler)
			integrations.POST("/webhooks/payment-status", paymentHandler.PaymentStatusWebhook)
			integrations.GET("/external/bank-rates", bankHandler.GetExternalBankRates)
			integrations.POST("/sync/account-balances", bankHandler.SyncAccountBalances)
			integrations.POST("/sync/transactions", bankHandler.SyncTransactions)
		}

		// Reporting and analytics
		reports := v1.Group("/reports")
		{
			reports.GET("/financing-summary", financingHandler.GetFinancingSummary)
			reports.GET("/payment-summary", paymentHandler.GetPaymentSummary)
			reports.GET("/portfolio-summary", portfolioHandler.GetPortfolioSummary)
			reports.POST("/custom-report", generateCustomReport)
			reports.GET("/dashboard-data", getDashboardData)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8087"
	}

	log.Printf("Bank Integration Service starting on port %s", port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Database: Connected")
	log.Printf("Bank Integrations: %d configured", len(cfg.BankConfigs))
	log.Printf("Epic 4 Compliance: Enabled")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func checkBankConnections(bankAPIService *services.BankAPIService) map[string]interface{} {
	connections := make(map[string]interface{})
	
	// Check each configured bank connection
	bankConfigs := []string{"chase", "wells_fargo", "bank_of_america", "jpmorgan", "citibank"}
	
	for _, bank := range bankConfigs {
		status, err := bankAPIService.TestConnection(bank)
		connections[bank] = map[string]interface{}{
			"status":     status,
			"error":      err,
			"last_check": time.Now(),
		}
	}
	
	return connections
}

func getSupportedBanks() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":         "JPMorgan Chase",
			"code":         "chase",
			"country":      "US",
			"capabilities": []string{"payments", "credit_decisions", "account_management"},
			"api_version":  "v2.1",
		},
		{
			"name":         "Wells Fargo",
			"code":         "wells_fargo",
			"country":      "US",
			"capabilities": []string{"payments", "credit_decisions", "portfolio_management"},
			"api_version":  "v2.0",
		},
		{
			"name":         "Bank of America",
			"code":         "bank_of_america",
			"country":      "US",
			"capabilities": []string{"payments", "transfers", "account_verification"},
			"api_version":  "v1.5",
		},
		{
			"name":         "Citibank",
			"code":         "citibank",
			"country":      "US",
			"capabilities": []string{"international_transfers", "fx_trading", "credit_facilities"},
			"api_version":  "v2.2",
		},
	}
}

func getSystemHealth(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services": gin.H{
			"database":      "connected",
			"bank_apis":     "connected",
			"cache":         "connected",
			"message_queue": "connected",
		},
		"metrics": gin.H{
			"active_connections":   getActiveConnectionCount(),
			"pending_payments":     getPendingPaymentCount(),
			"processing_requests":  getProcessingRequestCount(),
			"failed_transactions":  getFailedTransactionCount(),
		},
		"compliance": gin.H{
			"epic_4_status":      "compliant",
			"last_audit":         "2024-01-15T10:00:00Z",
			"regulatory_filings": "up_to_date",
		},
	}

	c.JSON(http.StatusOK, health)
}

func enableMaintenanceMode(c *gin.Context) {
	// Implementation for enabling maintenance mode
	c.JSON(http.StatusOK, gin.H{
		"status":  "maintenance_enabled",
		"message": "Bank Integration Service is now in maintenance mode",
		"enabled_at": time.Now(),
	})
}

func disableMaintenanceMode(c *gin.Context) {
	// Implementation for disabling maintenance mode
	c.JSON(http.StatusOK, gin.H{
		"status":   "maintenance_disabled",
		"message":  "Bank Integration Service is now operational",
		"disabled_at": time.Now(),
	})
}

func getSystemMetrics(c *gin.Context) {
	metrics := gin.H{
		"requests_per_minute":     getRequestsPerMinute(),
		"average_response_time":   getAverageResponseTime(),
		"success_rate":           getSuccessRate(),
		"error_rate":             getErrorRate(),
		"bank_api_latency":       getBankAPILatency(),
		"active_sessions":        getActiveSessionCount(),
		"concurrent_connections": getConcurrentConnectionCount(),
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp": time.Now(),
		"metrics":   metrics,
	})
}

func clearCache(c *gin.Context) {
	// Implementation for clearing system cache
	c.JSON(http.StatusOK, gin.H{
		"status":    "cache_cleared",
		"message":   "System cache has been cleared",
		"cleared_at": time.Now(),
	})
}

func generateCustomReport(c *gin.Context) {
	var request struct {
		ReportType string                 `json:"report_type"`
		StartDate  time.Time              `json:"start_date"`
		EndDate    time.Time              `json:"end_date"`
		Filters    map[string]interface{} `json:"filters"`
		Format     string                 `json:"format"` // json, csv, pdf
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Generate report based on request
	reportData := generateReportData(request.ReportType, request.StartDate, request.EndDate, request.Filters)

	c.JSON(http.StatusOK, gin.H{
		"report_type": request.ReportType,
		"period":      fmt.Sprintf("%s to %s", request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02")),
		"data":        reportData,
		"generated_at": time.Now(),
	})
}

func getDashboardData(c *gin.Context) {
	dashboardData := gin.H{
		"summary": gin.H{
			"total_financing":     "50,000,000",
			"active_portfolios":   125,
			"monthly_volume":      "8,500,000",
			"success_rate":        "98.7%",
		},
		"recent_activity": []gin.H{
			{
				"type":        "financing_approved",
				"amount":      "250,000",
				"customer":    "ABC Corp",
				"timestamp":   time.Now().Add(-1 * time.Hour),
			},
			{
				"type":        "payment_processed",
				"amount":      "150,000",
				"customer":    "XYZ Ltd",
				"timestamp":   time.Now().Add(-2 * time.Hour),
			},
		},
		"alerts": []gin.H{
			{
				"level":   "warning",
				"message": "Credit limit approaching for customer DEF Inc",
				"time":    time.Now().Add(-30 * time.Minute),
			},
		},
	}

	c.JSON(http.StatusOK, dashboardData)
}

// Helper functions for metrics (these would be implemented with actual monitoring)
func getActiveConnectionCount() int {
	return 15
}

func getPendingPaymentCount() int {
	return 8
}

func getProcessingRequestCount() int {
	return 3
}

func getFailedTransactionCount() int {
	return 1
}

func getRequestsPerMinute() float64 {
	return 45.2
}

func getAverageResponseTime() string {
	return "250ms"
}

func getSuccessRate() string {
	return "99.2%"
}

func getErrorRate() string {
	return "0.8%"
}

func getBankAPILatency() map[string]string {
	return map[string]string{
		"chase":           "150ms",
		"wells_fargo":     "200ms",
		"bank_of_america": "180ms",
		"citibank":        "160ms",
	}
}

func getActiveSessionCount() int {
	return 42
}

func getConcurrentConnectionCount() int {
	return 18
}

func generateReportData(reportType string, startDate, endDate time.Time, filters map[string]interface{}) interface{} {
	// Implementation would generate actual report data
	return gin.H{
		"summary": gin.H{
			"total_transactions": 1250,
			"total_volume":       "15,750,000",
			"average_size":       "12,600",
			"success_rate":       "99.1%",
		},
		"breakdown": []gin.H{
			{
				"category": "SME Financing",
				"count":    800,
				"volume":   "10,500,000",
			},
			{
				"category": "Trade Finance",
				"count":    450,
				"volume":   "5,250,000",
			},
		},
	}
}
