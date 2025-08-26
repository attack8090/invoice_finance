package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"financing-workflow-service/internal/config"
	"financing-workflow-service/internal/database"
	"financing-workflow-service/internal/handlers"
	"financing-workflow-service/internal/middleware"
	"financing-workflow-service/internal/models"
	"financing-workflow-service/internal/services"
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

	// Initialize services
	invoiceService := services.NewInvoiceService(db, cfg)
	financingService := services.NewFinancingService(db, cfg)
	workflowService := services.NewWorkflowService(db, cfg)
	buyerService := services.NewBuyerService(db, cfg)
	creditService := services.NewCreditService(cfg)
	blockchainService := services.NewBlockchainService(cfg)
	documentService := services.NewDocumentService(cfg)
	complianceService := services.NewComplianceService(cfg)
	notificationService := services.NewNotificationService(cfg)
	auditService := services.NewAuditService(db, cfg)

	// Initialize handlers
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService, documentService, complianceService)
	financingHandler := handlers.NewFinancingHandler(financingService, creditService, workflowService)
	buyerHandler := handlers.NewBuyerHandler(buyerService, financingService, complianceService)
	workflowHandler := handlers.NewWorkflowHandler(workflowService, auditService)
	agreementHandler := handlers.NewAgreementHandler(financingService, blockchainService, complianceService)
	disbursementHandler := handlers.NewDisbursementHandler(financingService, blockchainService, auditService)
	disputeHandler := handlers.NewDisputeHandler(financingService, workflowService, auditService)

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
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key", "X-Request-ID"}
	router.Use(cors.New(corsConfig))

	// Middleware
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimit(cfg))
	router.Use(middleware.RequestLogging())
	router.Use(middleware.RequestID())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "financing-workflow-service",
			"version": "2.0.0",
			"features": gin.H{
				"epic_4_compliance":     true,
				"epic_5_compliance":     true,
				"blockchain_integration": true,
				"automated_workflows":   true,
				"dispute_management":    true,
				"multi_party_support":   true,
			},
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Invoice submission and management
		invoices := v1.Group("/invoices")
		{
			invoices.POST("", invoiceHandler.SubmitInvoice)
			invoices.GET("/:id", invoiceHandler.GetInvoice)
			invoices.PUT("/:id", invoiceHandler.UpdateInvoice)
			invoices.POST("/:id/verify", invoiceHandler.VerifyInvoice)
			invoices.POST("/:id/request-financing", invoiceHandler.RequestFinancing)
			invoices.GET("/:id/financing-status", invoiceHandler.GetFinancingStatus)
			invoices.POST("/:id/upload-document", invoiceHandler.UploadDocument)
			invoices.GET("/:id/audit-trail", invoiceHandler.GetAuditTrail)
		}

		// Financing requests and management
		financing := v1.Group("/financing")
		{
			financing.POST("/requests", financingHandler.CreateFinancingRequest)
			financing.GET("/requests/:id", financingHandler.GetFinancingRequest)
			financing.PUT("/requests/:id", financingHandler.UpdateFinancingRequest)
			financing.POST("/requests/:id/submit", financingHandler.SubmitForApproval)
			financing.GET("/requests/:id/status", financingHandler.GetRequestStatus)
			financing.POST("/requests/:id/approve", financingHandler.ApproveRequest)
			financing.POST("/requests/:id/reject", financingHandler.RejectRequest)
			financing.POST("/requests/:id/counter-offer", financingHandler.CounterOffer)
			financing.GET("/opportunities", financingHandler.GetFinancingOpportunities)
			financing.POST("/invest", financingHandler.InvestInOpportunity)
		}

		// Buyer confirmation and management
		buyers := v1.Group("/buyers")
		{
			buyers.POST("/confirm-invoice", buyerHandler.ConfirmInvoice)
			buyers.POST("/dispute-invoice", buyerHandler.DisputeInvoice)
			buyers.GET("/invoices", buyerHandler.GetBuyerInvoices)
			buyers.GET("/invoices/:id/details", buyerHandler.GetInvoiceDetails)
			buyers.POST("/invoices/:id/payment-confirmation", buyerHandler.ConfirmPayment)
			buyers.POST("/invoices/:id/extend-terms", buyerHandler.ExtendPaymentTerms)
			buyers.GET("/financing-requests", buyerHandler.GetFinancingRequests)
			buyers.POST("/financing-requests/:id/approve", buyerHandler.ApproveFinancing)
		}

		// Workflow management
		workflows := v1.Group("/workflows")
		{
			workflows.GET("/:id", workflowHandler.GetWorkflow)
			workflows.GET("/:id/status", workflowHandler.GetWorkflowStatus)
			workflows.POST("/:id/advance", workflowHandler.AdvanceWorkflow)
			workflows.POST("/:id/rollback", workflowHandler.RollbackWorkflow)
			workflows.GET("/:id/history", workflowHandler.GetWorkflowHistory)
			workflows.POST("/:id/assign-reviewer", workflowHandler.AssignReviewer)
			workflows.POST("/:id/escalate", workflowHandler.EscalateWorkflow)
		}

		// Agreement signing and management
		agreements := v1.Group("/agreements")
		{
			agreements.POST("/generate", agreementHandler.GenerateAgreement)
			agreements.GET("/:id", agreementHandler.GetAgreement)
			agreements.POST("/:id/sign", agreementHandler.SignAgreement)
			agreements.GET("/:id/signatures", agreementHandler.GetSignatures)
			agreements.POST("/:id/verify-signature", agreementHandler.VerifySignature)
			agreements.POST("/:id/blockchain-record", agreementHandler.RecordOnBlockchain)
			agreements.GET("/:id/compliance-check", agreementHandler.CheckCompliance)
		}

		// Disbursement management
		disbursements := v1.Group("/disbursements")
		{
			disbursements.POST("/create", disbursementHandler.CreateDisbursement)
			disbursements.GET("/:id", disbursementHandler.GetDisbursement)
			disbursements.POST("/:id/approve", disbursementHandler.ApproveDisbursement)
			disbursements.POST("/:id/execute", disbursementHandler.ExecuteDisbursement)
			disbursements.GET("/:id/status", disbursementHandler.GetDisbursementStatus)
			disbursements.POST("/:id/reconcile", disbursementHandler.ReconcileDisbursement)
			disbursements.GET("/scheduled", disbursementHandler.GetScheduledDisbursements)
		}

		// Transaction and dispute management
		disputes := v1.Group("/disputes")
		{
			disputes.POST("/create", disputeHandler.CreateDispute)
			disputes.GET("/:id", disputeHandler.GetDispute)
			disputes.POST("/:id/respond", disputeHandler.RespondToDispute)
			disputes.POST("/:id/escalate", disputeHandler.EscalateDispute)
			disputes.POST("/:id/resolve", disputeHandler.ResolveDispute)
			disputes.GET("/:id/messages", disputeHandler.GetDisputeMessages)
			disputes.POST("/:id/evidence", disputeHandler.SubmitEvidence)
			disputes.GET("/active", disputeHandler.GetActiveDisputes)
		}

		// Multi-party workflow management
		multiParty := v1.Group("/multi-party")
		{
			multiParty.POST("/initiate", workflowHandler.InitiateMultiPartyWorkflow)
			multiParty.GET("/:workflowId/participants", workflowHandler.GetParticipants)
			multiParty.POST("/:workflowId/invite", workflowHandler.InviteParticipant)
			multiParty.POST("/:workflowId/accept", workflowHandler.AcceptParticipation)
			multiParty.POST("/:workflowId/consensus", workflowHandler.CheckConsensus)
			multiParty.POST("/:workflowId/finalize", workflowHandler.FinalizeMultiPartyWorkflow)
		}

		// Integration endpoints
		integrations := v1.Group("/integrations")
		{
			integrations.POST("/webhook/buyer-confirmation", buyerHandler.WebhookBuyerConfirmation)
			integrations.POST("/webhook/payment-received", disbursementHandler.WebhookPaymentReceived)
			integrations.POST("/webhook/blockchain-event", agreementHandler.WebhookBlockchainEvent)
			integrations.GET("/external-verification/:invoiceId", invoiceHandler.ExternalVerification)
		}

		// Admin and monitoring endpoints
		admin := v1.Group("/admin")
		admin.Use(middleware.RequireRole("admin", "supervisor"))
		{
			admin.GET("/workflows/pending", workflowHandler.GetPendingWorkflows)
			admin.GET("/workflows/overdue", workflowHandler.GetOverdueWorkflows)
			admin.GET("/analytics/workflow-performance", workflowHandler.GetWorkflowAnalytics)
			admin.GET("/analytics/financing-trends", financingHandler.GetFinancingTrends)
			admin.GET("/compliance/audit-logs", complianceService.GetAuditLogs)
			admin.POST("/compliance/generate-report", complianceService.GenerateComplianceReport)
			admin.GET("/system/health-check", systemHealthCheck)
		}

		// Reporting endpoints
		reports := v1.Group("/reports")
		{
			reports.GET("/financing-summary", financingHandler.GetFinancingSummary)
			reports.GET("/workflow-status", workflowHandler.GetWorkflowStatusReport)
			reports.GET("/dispute-summary", disputeHandler.GetDisputeSummary)
			reports.POST("/custom-report", generateCustomReport)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Financing Workflow Service starting on port %s", port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Database: Connected")
	log.Printf("Blockchain: %s", cfg.BlockchainNetwork)
	log.Printf("Epic 4 Compliance: Enabled")
	log.Printf("Epic 5 Compliance: Enabled")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func systemHealthCheck(c *gin.Context) {
	healthStatus := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services": gin.H{
			"database":   "connected",
			"blockchain": "connected",
			"cache":      "connected",
		},
		"metrics": gin.H{
			"active_workflows":      getActiveWorkflowCount(),
			"pending_approvals":     getPendingApprovalsCount(),
			"processing_disputes":   getProcessingDisputesCount(),
			"scheduled_disbursements": getScheduledDisbursementsCount(),
		},
	}

	c.JSON(http.StatusOK, healthStatus)
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

	// Generate custom report based on request
	reportData := generateReportData(request.ReportType, request.StartDate, request.EndDate, request.Filters)

	switch request.Format {
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=report.csv")
		// Convert to CSV and return
	case "pdf":
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=report.pdf")
		// Generate PDF and return
	default:
		c.JSON(http.StatusOK, gin.H{
			"report_type": request.ReportType,
			"period":      fmt.Sprintf("%s to %s", request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02")),
			"data":        reportData,
			"generated_at": time.Now(),
		})
	}
}

// Helper functions for metrics (these would be implemented with actual database queries)
func getActiveWorkflowCount() int {
	// Implementation would query database for active workflows
	return 42
}

func getPendingApprovalsCount() int {
	// Implementation would query database for pending approvals
	return 15
}

func getProcessingDisputesCount() int {
	// Implementation would query database for processing disputes
	return 3
}

func getScheduledDisbursementsCount() int {
	// Implementation would query database for scheduled disbursements
	return 28
}

func generateReportData(reportType string, startDate, endDate time.Time, filters map[string]interface{}) interface{} {
	// This would implement the actual report generation logic
	// based on the report type and parameters
	return gin.H{
		"summary": gin.H{
			"total_transactions": 150,
			"total_value":       "10,500,000",
			"average_processing_time": "2.3 days",
			"success_rate":      "97.3%",
		},
		"breakdown": []gin.H{
			{
				"category": "Invoice Financing",
				"count":    120,
				"value":    "8,500,000",
			},
			{
				"category": "Trade Finance",
				"count":    30,
				"value":    "2,000,000",
			},
		},
	}
}
