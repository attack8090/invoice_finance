package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Server configuration
	Environment    string
	Port           string
	Host           string
	TLSEnabled     bool
	CertFile       string
	KeyFile        string
	
	// Database
	DatabaseURL    string
	DatabaseDriver string
	MaxConnections int
	
	// JWT and Security
	JWTSecret         string
	JWTExpiration     time.Duration
	APIKeyRequired    bool
	RateLimitRequests int
	RateLimitWindow   time.Duration
	
	// CORS
	AllowedOrigins []string
	AllowedHeaders []string
	
	// Bank API configurations
	BankConfigs map[string]BankConfig
	
	// Epic 4 Compliance
	Epic4Config Epic4Config
	
	// External services
	RedisURL         string
	MessageQueueURL  string
	AuditServiceURL  string
	
	// Monitoring and logging
	LogLevel         string
	MetricsEnabled   bool
	TracingEnabled   bool
	
	// Business configuration
	MaxCreditAmount    float64
	MinCreditAmount    float64
	DefaultCreditTerms int
	FundingMatchingEnabled bool
	RealTimeProcessing     bool
	
	// Compliance thresholds
	MaxDailyTransactionAmount  float64
	MaxMonthlyTransactionAmount float64
	SuspiciousActivityThreshold float64
	RequiredKYCDocuments       []string
	AMLCheckRequired           bool
	
	// Feature flags
	EnableBulkProcessing      bool
	EnableRealTimeTransfers   bool
	EnableCreditScoring       bool
	EnableFraudDetection      bool
	EnableRegulatoryReporting bool
}

type BankConfig struct {
	Name            string
	Code            string
	APIBaseURL      string
	APIKey          string
	APISecret       string
	ClientID        string
	ClientSecret    string
	CertificatePath string
	PrivateKeyPath  string
	APIVersion      string
	Timeout         time.Duration
	RetryAttempts   int
	TestMode        bool
	Capabilities    []string
	RateLimit       int
	MaxAmount       float64
	SupportedCurrencies []string
}

type Epic4Config struct {
	Enabled                 bool
	ReportingEndpoint      string
	APIKey                 string
	CertificatePath        string
	AuditTrailRequired     bool
	TransactionReporting   bool
	RegulatoryFilingAuto   bool
	ComplianceChecks       []string
	ReportingFrequency     string
	DataRetentionPeriod    time.Duration
	EncryptionRequired     bool
}

func Load() *Config {
	return &Config{
		// Server configuration
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8087"),
		Host:        getEnv("HOST", "0.0.0.0"),
		TLSEnabled:  getEnvBool("TLS_ENABLED", false),
		CertFile:    getEnv("TLS_CERT_FILE", ""),
		KeyFile:     getEnv("TLS_KEY_FILE", ""),
		
		// Database
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/bank_integration"),
		DatabaseDriver: getEnv("DATABASE_DRIVER", "postgres"),
		MaxConnections: getEnvInt("DATABASE_MAX_CONNECTIONS", 25),
		
		// JWT and Security
		JWTSecret:         getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpiration:     getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
		APIKeyRequired:    getEnvBool("API_KEY_REQUIRED", true),
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvDuration("RATE_LIMIT_WINDOW", time.Minute),
		
		// CORS
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080"), ","),
		AllowedHeaders: strings.Split(getEnv("ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization,X-API-Key"), ","),
		
		// Bank configurations
		BankConfigs: loadBankConfigs(),
		
		// Epic 4 Compliance
		Epic4Config: loadEpic4Config(),
		
		// External services
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
		MessageQueueURL: getEnv("MESSAGE_QUEUE_URL", "amqp://guest:guest@localhost:5672/"),
		AuditServiceURL: getEnv("AUDIT_SERVICE_URL", "http://localhost:8089"),
		
		// Monitoring and logging
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MetricsEnabled: getEnvBool("METRICS_ENABLED", true),
		TracingEnabled: getEnvBool("TRACING_ENABLED", false),
		
		// Business configuration
		MaxCreditAmount:        getEnvFloat("MAX_CREDIT_AMOUNT", 10000000.0),
		MinCreditAmount:        getEnvFloat("MIN_CREDIT_AMOUNT", 1000.0),
		DefaultCreditTerms:     getEnvInt("DEFAULT_CREDIT_TERMS", 30),
		FundingMatchingEnabled: getEnvBool("FUNDING_MATCHING_ENABLED", true),
		RealTimeProcessing:     getEnvBool("REAL_TIME_PROCESSING", true),
		
		// Compliance thresholds
		MaxDailyTransactionAmount:   getEnvFloat("MAX_DAILY_TRANSACTION_AMOUNT", 1000000.0),
		MaxMonthlyTransactionAmount: getEnvFloat("MAX_MONTHLY_TRANSACTION_AMOUNT", 10000000.0),
		SuspiciousActivityThreshold: getEnvFloat("SUSPICIOUS_ACTIVITY_THRESHOLD", 50000.0),
		RequiredKYCDocuments:        strings.Split(getEnv("REQUIRED_KYC_DOCUMENTS", "id_document,proof_of_address,financial_statements"), ","),
		AMLCheckRequired:            getEnvBool("AML_CHECK_REQUIRED", true),
		
		// Feature flags
		EnableBulkProcessing:      getEnvBool("ENABLE_BULK_PROCESSING", true),
		EnableRealTimeTransfers:   getEnvBool("ENABLE_REAL_TIME_TRANSFERS", true),
		EnableCreditScoring:       getEnvBool("ENABLE_CREDIT_SCORING", true),
		EnableFraudDetection:      getEnvBool("ENABLE_FRAUD_DETECTION", true),
		EnableRegulatoryReporting: getEnvBool("ENABLE_REGULATORY_REPORTING", true),
	}
}

func loadBankConfigs() map[string]BankConfig {
	banks := make(map[string]BankConfig)
	
	// JPMorgan Chase
	banks["chase"] = BankConfig{
		Name:              "JPMorgan Chase",
		Code:              "chase",
		APIBaseURL:        getEnv("CHASE_API_BASE_URL", "https://api.chase.com"),
		APIKey:            getEnv("CHASE_API_KEY", ""),
		APISecret:         getEnv("CHASE_API_SECRET", ""),
		ClientID:          getEnv("CHASE_CLIENT_ID", ""),
		ClientSecret:      getEnv("CHASE_CLIENT_SECRET", ""),
		CertificatePath:   getEnv("CHASE_CERT_PATH", ""),
		PrivateKeyPath:    getEnv("CHASE_KEY_PATH", ""),
		APIVersion:        getEnv("CHASE_API_VERSION", "v2.1"),
		Timeout:           getEnvDuration("CHASE_TIMEOUT", 30*time.Second),
		RetryAttempts:     getEnvInt("CHASE_RETRY_ATTEMPTS", 3),
		TestMode:          getEnvBool("CHASE_TEST_MODE", true),
		Capabilities:      strings.Split(getEnv("CHASE_CAPABILITIES", "payments,credit_decisions,account_management"), ","),
		RateLimit:         getEnvInt("CHASE_RATE_LIMIT", 1000),
		MaxAmount:         getEnvFloat("CHASE_MAX_AMOUNT", 5000000.0),
		SupportedCurrencies: strings.Split(getEnv("CHASE_CURRENCIES", "USD,EUR,GBP"), ","),
	}
	
	// Wells Fargo
	banks["wells_fargo"] = BankConfig{
		Name:              "Wells Fargo",
		Code:              "wells_fargo",
		APIBaseURL:        getEnv("WELLS_FARGO_API_BASE_URL", "https://api.wellsfargo.com"),
		APIKey:            getEnv("WELLS_FARGO_API_KEY", ""),
		APISecret:         getEnv("WELLS_FARGO_API_SECRET", ""),
		ClientID:          getEnv("WELLS_FARGO_CLIENT_ID", ""),
		ClientSecret:      getEnv("WELLS_FARGO_CLIENT_SECRET", ""),
		CertificatePath:   getEnv("WELLS_FARGO_CERT_PATH", ""),
		PrivateKeyPath:    getEnv("WELLS_FARGO_KEY_PATH", ""),
		APIVersion:        getEnv("WELLS_FARGO_API_VERSION", "v2.0"),
		Timeout:           getEnvDuration("WELLS_FARGO_TIMEOUT", 30*time.Second),
		RetryAttempts:     getEnvInt("WELLS_FARGO_RETRY_ATTEMPTS", 3),
		TestMode:          getEnvBool("WELLS_FARGO_TEST_MODE", true),
		Capabilities:      strings.Split(getEnv("WELLS_FARGO_CAPABILITIES", "payments,credit_decisions,portfolio_management"), ","),
		RateLimit:         getEnvInt("WELLS_FARGO_RATE_LIMIT", 800),
		MaxAmount:         getEnvFloat("WELLS_FARGO_MAX_AMOUNT", 3000000.0),
		SupportedCurrencies: strings.Split(getEnv("WELLS_FARGO_CURRENCIES", "USD,CAD"), ","),
	}
	
	// Bank of America
	banks["bank_of_america"] = BankConfig{
		Name:              "Bank of America",
		Code:              "bank_of_america",
		APIBaseURL:        getEnv("BOA_API_BASE_URL", "https://api.bankofamerica.com"),
		APIKey:            getEnv("BOA_API_KEY", ""),
		APISecret:         getEnv("BOA_API_SECRET", ""),
		ClientID:          getEnv("BOA_CLIENT_ID", ""),
		ClientSecret:      getEnv("BOA_CLIENT_SECRET", ""),
		CertificatePath:   getEnv("BOA_CERT_PATH", ""),
		PrivateKeyPath:    getEnv("BOA_KEY_PATH", ""),
		APIVersion:        getEnv("BOA_API_VERSION", "v1.5"),
		Timeout:           getEnvDuration("BOA_TIMEOUT", 30*time.Second),
		RetryAttempts:     getEnvInt("BOA_RETRY_ATTEMPTS", 3),
		TestMode:          getEnvBool("BOA_TEST_MODE", true),
		Capabilities:      strings.Split(getEnv("BOA_CAPABILITIES", "payments,transfers,account_verification"), ","),
		RateLimit:         getEnvInt("BOA_RATE_LIMIT", 600),
		MaxAmount:         getEnvFloat("BOA_MAX_AMOUNT", 2000000.0),
		SupportedCurrencies: strings.Split(getEnv("BOA_CURRENCIES", "USD"), ","),
	}
	
	// Citibank
	banks["citibank"] = BankConfig{
		Name:              "Citibank",
		Code:              "citibank",
		APIBaseURL:        getEnv("CITI_API_BASE_URL", "https://api.citibank.com"),
		APIKey:            getEnv("CITI_API_KEY", ""),
		APISecret:         getEnv("CITI_API_SECRET", ""),
		ClientID:          getEnv("CITI_CLIENT_ID", ""),
		ClientSecret:      getEnv("CITI_CLIENT_SECRET", ""),
		CertificatePath:   getEnv("CITI_CERT_PATH", ""),
		PrivateKeyPath:    getEnv("CITI_KEY_PATH", ""),
		APIVersion:        getEnv("CITI_API_VERSION", "v2.2"),
		Timeout:           getEnvDuration("CITI_TIMEOUT", 30*time.Second),
		RetryAttempts:     getEnvInt("CITI_RETRY_ATTEMPTS", 3),
		TestMode:          getEnvBool("CITI_TEST_MODE", true),
		Capabilities:      strings.Split(getEnv("CITI_CAPABILITIES", "international_transfers,fx_trading,credit_facilities"), ","),
		RateLimit:         getEnvInt("CITI_RATE_LIMIT", 1200),
		MaxAmount:         getEnvFloat("CITI_MAX_AMOUNT", 10000000.0),
		SupportedCurrencies: strings.Split(getEnv("CITI_CURRENCIES", "USD,EUR,GBP,JPY,AUD,CAD"), ","),
	}
	
	return banks
}

func loadEpic4Config() Epic4Config {
	return Epic4Config{
		Enabled:                getEnvBool("EPIC4_ENABLED", true),
		ReportingEndpoint:      getEnv("EPIC4_REPORTING_ENDPOINT", "https://epic4.reporting.gov"),
		APIKey:                getEnv("EPIC4_API_KEY", ""),
		CertificatePath:       getEnv("EPIC4_CERT_PATH", ""),
		AuditTrailRequired:    getEnvBool("EPIC4_AUDIT_TRAIL_REQUIRED", true),
		TransactionReporting:  getEnvBool("EPIC4_TRANSACTION_REPORTING", true),
		RegulatoryFilingAuto:  getEnvBool("EPIC4_AUTO_FILING", true),
		ComplianceChecks:      strings.Split(getEnv("EPIC4_COMPLIANCE_CHECKS", "kyc,aml,sanctions,pep"), ","),
		ReportingFrequency:    getEnv("EPIC4_REPORTING_FREQUENCY", "daily"),
		DataRetentionPeriod:   getEnvDuration("EPIC4_DATA_RETENTION", 7*365*24*time.Hour), // 7 years
		EncryptionRequired:    getEnvBool("EPIC4_ENCRYPTION_REQUIRED", true),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
