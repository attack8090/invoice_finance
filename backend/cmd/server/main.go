package main

import (
	"log"
	"os"

	"invoice-financing-platform/internal/api"
	"invoice-financing-platform/internal/config"
	"invoice-financing-platform/internal/database"
	"invoice-financing-platform/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize services
	userService := services.NewUserService(db)
	invoiceService := services.NewInvoiceService(db)
	financingService := services.NewFinancingService(db)
	blockchainService := services.NewBlockchainService(cfg.EthereumRPC, cfg.ContractAddress)
	aiService := services.NewAIService(cfg.AIModelEndpoint)
	fileService := services.NewFileService()

	// Initialize API server
	server := api.NewServer(api.ServerConfig{
		UserService:       userService,
		InvoiceService:    invoiceService,
		FinancingService:  financingService,
		BlockchainService: blockchainService,
		AIService:         aiService,
		FileService:       fileService,
		JWTSecret:         cfg.JWTSecret,
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := server.Start(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
