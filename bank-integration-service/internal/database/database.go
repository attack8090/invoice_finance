package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"bank-integration-service/internal/models"
)

// Initialize creates a new database connection
func Initialize(databaseURL string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Determine database driver based on URL
	if databaseURL == "" || databaseURL == ":memory:" || databaseURL == "test.db" {
		// SQLite for development/testing
		db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		// PostgreSQL for production
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(300) // 5 minutes

	log.Println("Database connection established successfully")
	return db, nil
}

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.BankConnection{},
		&models.CreditDecision{},
		&models.PaymentTransaction{},
		&models.FinancingRequest{},
		&models.PortfolioItem{},
		&models.FundingSource{},
		&models.ComplianceRecord{},
		&models.AuditTrail{},
		&models.BankAccount{},
		&models.FundingMatching{},
		&models.RiskAssessment{},
		&models.ReconciliationJob{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create indexes for better performance
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// createIndexes creates database indexes for better query performance
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		// BankConnection indexes
		"CREATE INDEX IF NOT EXISTS idx_bank_connections_user_id ON bank_connections(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_bank_connections_bank_code ON bank_connections(bank_code)",
		"CREATE INDEX IF NOT EXISTS idx_bank_connections_status ON bank_connections(status)",

		// CreditDecision indexes
		"CREATE INDEX IF NOT EXISTS idx_credit_decisions_customer_id ON credit_decisions(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_credit_decisions_bank_connection_id ON credit_decisions(bank_connection_id)",
		"CREATE INDEX IF NOT EXISTS idx_credit_decisions_status ON credit_decisions(status)",
		"CREATE INDEX IF NOT EXISTS idx_credit_decisions_created_at ON credit_decisions(created_at)",

		// PaymentTransaction indexes
		"CREATE INDEX IF NOT EXISTS idx_payment_transactions_bank_connection_id ON payment_transactions(bank_connection_id)",
		"CREATE INDEX IF NOT EXISTS idx_payment_transactions_payment_id ON payment_transactions(payment_id)",
		"CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions(status)",
		"CREATE INDEX IF NOT EXISTS idx_payment_transactions_created_at ON payment_transactions(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_payment_transactions_processed_at ON payment_transactions(processed_at)",

		// FinancingRequest indexes
		"CREATE INDEX IF NOT EXISTS idx_financing_requests_customer_id ON financing_requests(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_financing_requests_status ON financing_requests(status)",
		"CREATE INDEX IF NOT EXISTS idx_financing_requests_request_type ON financing_requests(request_type)",
		"CREATE INDEX IF NOT EXISTS idx_financing_requests_created_at ON financing_requests(created_at)",

		// PortfolioItem indexes
		"CREATE INDEX IF NOT EXISTS idx_portfolio_items_bank_connection_id ON portfolio_items(bank_connection_id)",
		"CREATE INDEX IF NOT EXISTS idx_portfolio_items_customer_id ON portfolio_items(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_portfolio_items_status ON portfolio_items(status)",
		"CREATE INDEX IF NOT EXISTS idx_portfolio_items_maturity_date ON portfolio_items(maturity_date)",
		"CREATE INDEX IF NOT EXISTS idx_portfolio_items_risk_rating ON portfolio_items(risk_rating)",

		// FundingSource indexes
		"CREATE INDEX IF NOT EXISTS idx_funding_sources_bank_connection_id ON funding_sources(bank_connection_id)",
		"CREATE INDEX IF NOT EXISTS idx_funding_sources_type ON funding_sources(type)",
		"CREATE INDEX IF NOT EXISTS idx_funding_sources_status ON funding_sources(status)",

		// ComplianceRecord indexes
		"CREATE INDEX IF NOT EXISTS idx_compliance_records_entity_type ON compliance_records(entity_type)",
		"CREATE INDEX IF NOT EXISTS idx_compliance_records_entity_id ON compliance_records(entity_id)",
		"CREATE INDEX IF NOT EXISTS idx_compliance_records_compliance_type ON compliance_records(compliance_type)",
		"CREATE INDEX IF NOT EXISTS idx_compliance_records_status ON compliance_records(status)",
		"CREATE INDEX IF NOT EXISTS idx_compliance_records_checked_at ON compliance_records(checked_at)",

		// AuditTrail indexes
		"CREATE INDEX IF NOT EXISTS idx_audit_trails_entity_type ON audit_trails(entity_type)",
		"CREATE INDEX IF NOT EXISTS idx_audit_trails_entity_id ON audit_trails(entity_id)",
		"CREATE INDEX IF NOT EXISTS idx_audit_trails_user_id ON audit_trails(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_audit_trails_timestamp ON audit_trails(timestamp)",

		// BankAccount indexes
		"CREATE INDEX IF NOT EXISTS idx_bank_accounts_bank_connection_id ON bank_accounts(bank_connection_id)",
		"CREATE INDEX IF NOT EXISTS idx_bank_accounts_customer_id ON bank_accounts(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_bank_accounts_status ON bank_accounts(status)",

		// FundingMatching indexes
		"CREATE INDEX IF NOT EXISTS idx_funding_matching_financing_request_id ON funding_matchings(financing_request_id)",
		"CREATE INDEX IF NOT EXISTS idx_funding_matching_funding_source_id ON funding_matchings(funding_source_id)",
		"CREATE INDEX IF NOT EXISTS idx_funding_matching_status ON funding_matchings(status)",
		"CREATE INDEX IF NOT EXISTS idx_funding_matching_matched_at ON funding_matchings(matched_at)",

		// RiskAssessment indexes
		"CREATE INDEX IF NOT EXISTS idx_risk_assessments_entity_type ON risk_assessments(entity_type)",
		"CREATE INDEX IF NOT EXISTS idx_risk_assessments_entity_id ON risk_assessments(entity_id)",
		"CREATE INDEX IF NOT EXISTS idx_risk_assessments_assessment_type ON risk_assessments(assessment_type)",
		"CREATE INDEX IF NOT EXISTS idx_risk_assessments_risk_rating ON risk_assessments(risk_rating)",
		"CREATE INDEX IF NOT EXISTS idx_risk_assessments_assessed_at ON risk_assessments(assessed_at)",

		// ReconciliationJob indexes
		"CREATE INDEX IF NOT EXISTS idx_reconciliation_jobs_job_type ON reconciliation_jobs(job_type)",
		"CREATE INDEX IF NOT EXISTS idx_reconciliation_jobs_status ON reconciliation_jobs(status)",
		"CREATE INDEX IF NOT EXISTS idx_reconciliation_jobs_created_at ON reconciliation_jobs(created_at)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			// Log the error but don't fail the migration if index creation fails
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	return nil
}

// SeedData inserts initial data for development/testing
func SeedData(db *gorm.DB) error {
	log.Println("Seeding database with initial data...")

	// Check if data already exists
	var count int64
	db.Model(&models.BankConnection{}).Count(&count)
	if count > 0 {
		log.Println("Database already contains data, skipping seed")
		return nil
	}

	// Add seed data here if needed for development
	log.Println("Database seeded successfully")
	return nil
}

// HealthCheck verifies database connectivity
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
