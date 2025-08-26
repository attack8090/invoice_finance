package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"blockchain-ledger-service/internal/config"
	"blockchain-ledger-service/internal/database"
	"blockchain-ledger-service/internal/handlers"
	"blockchain-ledger-service/internal/middleware"
	"blockchain-ledger-service/internal/models"
	"blockchain-ledger-service/internal/services"
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

	// Initialize Hyperledger Fabric gateway
	fabricGateway, err := initializeFabricGateway(cfg)
	if err != nil {
		log.Fatal("Failed to initialize Hyperledger Fabric gateway:", err)
	}
	defer fabricGateway.Close()

	// Initialize services
	ledgerService := services.NewLedgerService(db, fabricGateway, cfg)
	tokenizationService := services.NewTokenizationService(db, fabricGateway, cfg)
	auditService := services.NewAuditService(db, fabricGateway, cfg)
	complianceService := services.NewComplianceService(db, cfg)
	duplicateCheckService := services.NewDuplicateCheckService(db, fabricGateway, cfg)

	// Initialize handlers
	ledgerHandler := handlers.NewLedgerHandler(ledgerService, auditService, complianceService)
	tokenHandler := handlers.NewTokenHandler(tokenizationService, duplicateCheckService, complianceService)
	auditHandler := handlers.NewAuditHandler(auditService, ledgerService)
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
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key", "X-Request-ID"}
	router.Use(cors.New(corsConfig))

	// Middleware
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RateLimit(cfg))
	router.Use(middleware.RequestLogging())
	router.Use(middleware.RequestID())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		fabricStatus := "connected"
		if err := checkFabricConnection(fabricGateway); err != nil {
			fabricStatus = "disconnected"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "blockchain-ledger-service",
			"version": "1.0.0",
			"features": gin.H{
				"epic_4_compliance":     true,
				"hyperledger_fabric":    true,
				"private_blockchain":    true,
				"asset_tokenization":    true,
				"duplicate_prevention":  true,
				"audit_trail":          true,
				"transparency":         true,
			},
			"fabric_status": fabricStatus,
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Ledger management
		ledger := v1.Group("/ledger")
		{
			ledger.POST("/record", ledgerHandler.RecordTransaction)
			ledger.GET("/transaction/:id", ledgerHandler.GetTransaction)
			ledger.GET("/transactions", ledgerHandler.GetTransactions)
			ledger.POST("/verify", ledgerHandler.VerifyTransaction)
			ledger.GET("/balance/:assetId", ledgerHandler.GetAssetBalance)
			ledger.GET("/history/:assetId", ledgerHandler.GetAssetHistory)
			ledger.POST("/transfer", ledgerHandler.TransferAsset)
		}

		// Asset tokenization
		tokens := v1.Group("/tokens")
		{
			tokens.POST("/create", tokenHandler.CreateToken)
			tokens.GET("/:tokenId", tokenHandler.GetToken)
			tokens.POST("/:tokenId/transfer", tokenHandler.TransferToken)
			tokens.POST("/:tokenId/split", tokenHandler.SplitToken)
			tokens.POST("/:tokenId/merge", tokenHandler.MergeTokens)
			tokens.GET("/:tokenId/ownership", tokenHandler.GetTokenOwnership)
			tokens.POST("/:tokenId/freeze", tokenHandler.FreezeToken)
			tokens.POST("/:tokenId/unfreeze", tokenHandler.UnfreezeToken)
		}

		// Invoice/PO tokenization
		invoices := v1.Group("/invoices")
		{
			invoices.POST("/tokenize", tokenHandler.TokenizeInvoice)
			invoices.GET("/tokenized/:invoiceId", tokenHandler.GetTokenizedInvoice)
			invoices.POST("/finance", tokenHandler.FinanceTokenizedInvoice)
			invoices.GET("/financed", tokenHandler.GetFinancedInvoices)
			invoices.POST("/settle", tokenHandler.SettleInvoice)
			invoices.GET("/settlement-history/:invoiceId", tokenHandler.GetSettlementHistory)
		}

		// Purchase Order tokenization
		pos := v1.Group("/purchase-orders")
		{
			pos.POST("/tokenize", tokenHandler.TokenizePO)
			pos.GET("/tokenized/:poId", tokenHandler.GetTokenizedPO)
			pos.POST("/finance", tokenHandler.FinancePO)
			pos.GET("/financed", tokenHandler.GetFinancedPOs)
			pos.POST("/fulfill", tokenHandler.FulfillPO)
		}

		// Audit and compliance
		audit := v1.Group("/audit")
		{
			audit.GET("/trail/:assetId", auditHandler.GetAuditTrail)
			audit.GET("/transactions", auditHandler.GetAllTransactions)
			audit.POST("/verify-integrity", auditHandler.VerifyLedgerIntegrity)
			audit.GET("/compliance-report", auditHandler.GenerateComplianceReport)
			audit.GET("/duplicate-check/:assetId", auditHandler.CheckDuplicates)
		}

		// Duplicate prevention
		duplicates := v1.Group("/duplicates")
		{
			duplicates.POST("/check", tokenHandler.CheckForDuplicates)
			duplicates.GET("/potential", tokenHandler.GetPotentialDuplicates)
			duplicates.POST("/resolve", tokenHandler.ResolveDuplicate)
			duplicates.GET("/history", tokenHandler.GetDuplicateHistory)
		}

		// Private blockchain management
		blockchain := v1.Group("/blockchain")
		{
			blockchain.GET("/network-status", ledgerHandler.GetNetworkStatus)
			blockchain.GET("/peers", ledgerHandler.GetPeers)
			blockchain.GET("/channels", ledgerHandler.GetChannels)
			blockchain.POST("/channel/join", ledgerHandler.JoinChannel)
			blockchain.GET("/chaincode/list", ledgerHandler.ListChaincodes)
			blockchain.POST("/chaincode/invoke", ledgerHandler.InvokeChaincode)
			blockchain.POST("/chaincode/query", ledgerHandler.QueryChaincode)
		}

		// Admin endpoints
		admin := v1.Group("/admin")
		admin.Use(middleware.RequireRole("admin", "blockchain_admin"))
		{
			admin.POST("/emergency-freeze", tokenHandler.EmergencyFreezeAsset)
			admin.POST("/emergency-unfreeze", tokenHandler.EmergencyUnfreezeAsset)
			admin.GET("/system-metrics", ledgerHandler.GetSystemMetrics)
			admin.POST("/backup-ledger", auditHandler.BackupLedger)
			admin.POST("/restore-ledger", auditHandler.RestoreLedger)
			admin.GET("/node-health", ledgerHandler.GetNodeHealth)
			admin.POST("/consensus-check", ledgerHandler.CheckConsensus)
		}

		// Epic 4 compliance endpoints
		compliance := v1.Group("/compliance")
		{
			compliance.GET("/epic4/status", complianceHandler.GetEpic4ComplianceStatus)
			compliance.POST("/epic4/validate", complianceHandler.ValidateEpic4Compliance)
			compliance.GET("/epic4/audit-report", complianceHandler.GenerateEpic4AuditReport)
			compliance.GET("/transparency/public", complianceHandler.GetPublicTransparencyData)
			compliance.GET("/transparency/asset/:assetId", complianceHandler.GetAssetTransparencyData)
		}

		// Analytics and reporting
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/transaction-volume", auditHandler.GetTransactionVolume)
			analytics.GET("/asset-utilization", auditHandler.GetAssetUtilization)
			analytics.GET("/financing-trends", auditHandler.GetFinancingTrends)
			analytics.GET("/network-performance", ledgerHandler.GetNetworkPerformance)
			analytics.GET("/compliance-metrics", complianceHandler.GetComplianceMetrics)
		}

		// Integration endpoints
		integrations := v1.Group("/integrations")
		{
			integrations.POST("/webhook/settlement", tokenHandler.WebhookSettlement)
			integrations.POST("/webhook/payment", tokenHandler.WebhookPayment)
			integrations.GET("/external-verify/:transactionId", auditHandler.ExternalVerification)
			integrations.POST("/sync/external-system", ledgerHandler.SyncExternalSystem)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("Blockchain Ledger Service starting on port %s", port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Database: Connected")
	log.Printf("Hyperledger Fabric: %s", cfg.FabricNetwork)
	log.Printf("Epic 4 Compliance: Enabled")
	log.Printf("Private Blockchain: Enabled")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initializeFabricGateway(cfg *config.Config) (*gateway.Gateway, error) {
	// Initialize the Hyperledger Fabric SDK
	ccpPath := cfg.FabricConnectionProfile

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(ccpPath)),
		gateway.WithIdentity(cfg.FabricWallet, cfg.FabricUser),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	return gw, nil
}

func checkFabricConnection(gw *gateway.Gateway) error {
	// Get the network
	network, err := gw.GetNetwork(os.Getenv("FABRIC_CHANNEL_NAME"))
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	// Get a contract
	contract := network.GetContract(os.Getenv("FABRIC_CHAINCODE_NAME"))

	// Test the connection with a simple query
	_, err = contract.EvaluateTransaction("HealthCheck")
	if err != nil {
		return fmt.Errorf("failed to evaluate transaction: %w", err)
	}

	return nil
}

// Transaction represents a blockchain transaction
type Transaction struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	AssetID   string                 `json:"asset_id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Amount    float64                `json:"amount"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	Hash      string                 `json:"hash"`
	Status    string                 `json:"status"`
}

// TokenizedAsset represents a tokenized invoice or PO
type TokenizedAsset struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"` // invoice, purchase_order
	OriginalAssetID  string                 `json:"original_asset_id"`
	TokenID          string                 `json:"token_id"`
	Owner            string                 `json:"owner"`
	Value            float64                `json:"value"`
	Status           string                 `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	FinancingDetails map[string]interface{} `json:"financing_details"`
	ComplianceFlags  []string               `json:"compliance_flags"`
}

// AuditTrail represents an audit trail entry
type AuditTrail struct {
	ID          string                 `json:"id"`
	AssetID     string                 `json:"asset_id"`
	Action      string                 `json:"action"`
	Actor       string                 `json:"actor"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details"`
	Hash        string                 `json:"hash"`
	PrevHash    string                 `json:"prev_hash"`
	BlockNumber uint64                 `json:"block_number"`
}

// Helper functions
func generateTransactionHash(tx Transaction) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%f:%d",
		tx.ID, tx.Type, tx.From, tx.To, tx.Amount, tx.Timestamp.Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func generateAssetID(assetType, originalID string) string {
	data := fmt.Sprintf("%s:%s:%d", assetType, originalID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter ID
}
