package api

import (
	"invoice-financing-platform/internal/services"
	"invoice-financing-platform/middleware"
	"invoice-financing-platform/models"
	"os"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router            *gin.Engine
	userService       *services.UserService
	invoiceService    *services.InvoiceService
	financingService  *services.FinancingService
	blockchainService *services.BlockchainService
	aiService         *services.AIService
	fileService       *services.FileService
	jwtSecret         string
}

type ServerConfig struct {
	UserService       *services.UserService
	InvoiceService    *services.InvoiceService
	FinancingService  *services.FinancingService
	BlockchainService *services.BlockchainService
	AIService         *services.AIService
	FileService       *services.FileService
	JWTSecret         string
}

func NewServer(config ServerConfig) *Server {
	server := &Server{
		router:            gin.Default(),
		userService:       config.UserService,
		invoiceService:    config.InvoiceService,
		financingService:  config.FinancingService,
		blockchainService: config.BlockchainService,
		aiService:         config.AIService,
		fileService:       config.FileService,
		jwtSecret:         config.JWTSecret,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

func (s *Server) setupMiddleware() {
	// Setup comprehensive security middleware
	var securityConfig middleware.SecurityConfig
	if gin.Mode() == gin.ReleaseMode {
		securityConfig = middleware.ProductionSecurityConfig()
	} else {
		securityConfig = middleware.DefaultSecurityConfig()
	}
	middleware.SetupSecurity(s.router, securityConfig)

	// Add request ID and logging
	s.router.Use(middleware.RequestID())
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())

	// Add global rate limiting
	s.router.Use(middleware.RateLimit())

	// Add input sanitization
	s.router.Use(middleware.SanitizeInput())

	// Add query parameter validation to all routes
	s.router.Use(middleware.ValidateQueryParams)
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")
	
	// Health check
	api.GET("/health", s.healthCheck)
	
	// Authentication routes with strict rate limiting
	auth := api.Group("/auth")
	auth.Use(middleware.AuthRateLimit()) // Strict rate limiting for auth
	{
		auth.POST("/register", middleware.ValidateJSON(&models.UserRegistrationRequest{}), s.register)
		auth.POST("/login", middleware.ValidateJSON(&models.UserLoginRequest{}), s.login)
		auth.POST("/refresh", s.refreshToken)
		auth.POST("/logout", s.AuthMiddleware(), s.logout)
	}

	// User routes
	users := api.Group("/users")
	users.Use(s.AuthMiddleware())
	{
		users.GET("/profile", s.getUserProfile)
		users.PUT("/profile", s.updateUserProfile)
		users.POST("/verify", s.verifyUser)
		users.GET("/stats", s.getUserStats)
	}

	// Invoice routes
	invoices := api.Group("/invoices")
	invoices.Use(s.AuthMiddleware())
	{
		invoices.GET("", s.getInvoices)
		invoices.POST("", s.createInvoice)
		invoices.GET("/:id", s.getInvoice)
		invoices.PUT("/:id", s.updateInvoice)
		invoices.DELETE("/:id", s.deleteInvoice)
		invoices.POST("/:id/verify", s.verifyInvoice)
		invoices.POST("/:id/upload", s.uploadInvoiceDocument)
	}

	// Financing routes
	financing := api.Group("/financing")
	financing.Use(s.AuthMiddleware())
	{
		financing.GET("/requests", s.getFinancingRequests)
		financing.POST("/requests", s.createFinancingRequest)
		financing.GET("/requests/:id", s.getFinancingRequest)
		financing.PUT("/requests/:id", s.updateFinancingRequest)
		financing.POST("/requests/:id/approve", s.approveFinancingRequest)
		financing.POST("/requests/:id/reject", s.rejectFinancingRequest)
		
		// Investment endpoints
		financing.GET("/opportunities", s.getInvestmentOpportunities)
		financing.POST("/invest", s.createInvestment)
		financing.GET("/investments", s.getUserInvestments)
	}

	// Blockchain routes
	blockchain := api.Group("/blockchain")
	blockchain.Use(s.AuthMiddleware())
	{
		blockchain.POST("/tokenize-invoice", s.tokenizeInvoice)
		blockchain.GET("/transactions/:hash", s.getBlockchainTransaction)
		blockchain.POST("/verify-transaction", s.verifyTransaction)
	}

	// AI/ML routes
	ai := api.Group("/ai")
	ai.Use(s.AuthMiddleware())
	{
		ai.POST("/credit-score", s.calculateCreditScore)
		ai.POST("/risk-assessment", s.assessRisk)
		ai.POST("/fraud-detection", s.detectFraud)
		ai.POST("/verify-document", s.verifyDocument)
	}

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(s.AuthMiddleware(), s.AdminMiddleware())
	{
		admin.GET("/dashboard", s.getAdminDashboard)
		admin.GET("/users", s.getAllUsers)
		admin.GET("/transactions", s.getAllTransactions)
		admin.POST("/users/:id/verify", s.adminVerifyUser)
		admin.POST("/users/:id/suspend", s.suspendUser)
	}

	// Analytics routes
	analytics := api.Group("/analytics")
	analytics.Use(s.AuthMiddleware())
	{
		analytics.GET("/dashboard", s.getDashboardAnalytics)
		analytics.GET("/portfolio", s.getPortfolioAnalytics)
		analytics.GET("/market-trends", s.getMarketTrends)
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Invoice Financing Platform API is running",
		"version": "1.0.0",
	})
}
