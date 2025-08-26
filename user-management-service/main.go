package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"user-management-service/internal/config"
	"user-management-service/internal/database"
	"user-management-service/internal/handlers"
	"user-management-service/internal/middleware"
	"user-management-service/internal/services"
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
	userService := services.NewUserService(db, cfg)
	kycService := services.NewKYCService(db, cfg)
	authService := services.NewAuthService(db, cfg)
	mfaService := services.NewMFAService(db, cfg)
	complianceService := services.NewComplianceService(db, cfg)
	notificationService := services.NewNotificationService(cfg)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, kycService, authService, mfaService)
	kycHandler := handlers.NewKYCHandler(kycService, complianceService)
	authHandler := handlers.NewAuthHandler(authService, mfaService)
	adminHandler := handlers.NewAdminHandler(userService, kycService, complianceService)

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
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key"}
	router.Use(cors.New(corsConfig))

	// Middleware
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimit(cfg))
	router.Use(middleware.RequestLogging())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "user-management-service",
			"version": "1.0.0",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/resend-verification", authHandler.ResendVerification)
		}

		// MFA routes
		mfa := v1.Group("/mfa")
		mfa.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			mfa.POST("/setup", authHandler.SetupMFA)
			mfa.POST("/verify", authHandler.VerifyMFA)
			mfa.POST("/disable", authHandler.DisableMFA)
			mfa.GET("/backup-codes", authHandler.GenerateBackupCodes)
		}

		// User management routes (authenticated)
		users := v1.Group("/users")
		users.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.POST("/change-password", userHandler.ChangePassword)
			users.DELETE("/account", userHandler.DeleteAccount)
			users.GET("/sessions", userHandler.GetActiveSessions)
			users.DELETE("/sessions/:sessionId", userHandler.RevokeSession)
		}

		// KYC routes (authenticated)
		kyc := v1.Group("/kyc")
		kyc.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			kyc.GET("/status", kycHandler.GetKYCStatus)
			kyc.POST("/submit", kycHandler.SubmitKYCDocuments)
			kyc.GET("/requirements", kycHandler.GetKYCRequirements)
			kyc.PUT("/update", kycHandler.UpdateKYCInfo)
			kyc.POST("/resubmit", kycHandler.ResubmitKYC)
		}

		// Company management routes (authenticated SME/Buyer users)
		companies := v1.Group("/companies")
		companies.Use(middleware.JWTAuth(cfg.JWTSecret))
		companies.Use(middleware.RequireRole("sme", "buyer", "admin"))
		{
			companies.GET("/profile", userHandler.GetCompanyProfile)
			companies.PUT("/profile", userHandler.UpdateCompanyProfile)
			companies.POST("/documents", userHandler.UploadCompanyDocument)
			companies.GET("/documents", userHandler.GetCompanyDocuments)
			companies.DELETE("/documents/:documentId", userHandler.DeleteCompanyDocument)
		}

		// Admin routes (admin only)
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(cfg.JWTSecret))
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", adminHandler.GetUsers)
			admin.GET("/users/:userId", adminHandler.GetUser)
			admin.PUT("/users/:userId/status", adminHandler.UpdateUserStatus)
			admin.PUT("/users/:userId/role", adminHandler.UpdateUserRole)
			admin.GET("/kyc/pending", adminHandler.GetPendingKYC)
			admin.PUT("/kyc/:kycId/review", adminHandler.ReviewKYC)
			admin.GET("/compliance/reports", adminHandler.GetComplianceReports)
			admin.POST("/compliance/audit", adminHandler.TriggerComplianceAudit)
			admin.GET("/analytics/users", adminHandler.GetUserAnalytics)
			admin.GET("/analytics/kyc", adminHandler.GetKYCAnalytics)
		}

		// Bank routes (bank users only)
		bank := v1.Group("/bank")
		bank.Use(middleware.JWTAuth(cfg.JWTSecret))
		bank.Use(middleware.RequireRole("bank", "admin"))
		{
			bank.GET("/customers", userHandler.GetBankCustomers)
			bank.GET("/customers/:customerId", userHandler.GetBankCustomer)
			bank.PUT("/customers/:customerId/credit-limit", userHandler.UpdateCustomerCreditLimit)
			bank.GET("/risk-assessment/:customerId", userHandler.GetCustomerRiskAssessment)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("User Management Service starting on port %s", port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Database: Connected")
	log.Printf("KYC/AML: Enabled")
	log.Printf("MFA: Enabled")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
