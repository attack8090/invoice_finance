package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email             string    `json:"email" gorm:"unique;not null"`
	PasswordHash      string    `json:"-" gorm:"not null"`
	FirstName         string    `json:"first_name" gorm:"not null"`
	LastName          string    `json:"last_name" gorm:"not null"`
	Role              UserRole  `json:"role" gorm:"not null"`
	CompanyName       string    `json:"company_name"`
	TaxID             string    `json:"tax_id"`
	WalletAddress     string    `json:"wallet_address"`
	IsVerified        bool      `json:"is_verified" gorm:"default:false"`
	CreditScore       int       `json:"credit_score" gorm:"default:0"`
	TotalInvestment   float64   `json:"total_investment" gorm:"default:0"`
	TotalFinanced     float64   `json:"total_financed" gorm:"default:0"`
	ProfileCompleted  bool      `json:"profile_completed" gorm:"default:false"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

type UserRole string

const (
	RoleSME      UserRole = "sme"
	RoleInvestor UserRole = "investor"
	RoleAdmin    UserRole = "admin"
)

type Invoice struct {
	ID                  uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID              uuid.UUID       `json:"user_id" gorm:"type:uuid;not null"`
	InvoiceNumber       string          `json:"invoice_number" gorm:"not null"`
	CustomerName        string          `json:"customer_name" gorm:"not null"`
	CustomerEmail       string          `json:"customer_email"`
	InvoiceAmount       float64         `json:"invoice_amount" gorm:"not null"`
	DueDate             time.Time       `json:"due_date" gorm:"not null"`
	IssueDate           time.Time       `json:"issue_date" gorm:"not null"`
	Description         string          `json:"description"`
	Status              InvoiceStatus   `json:"status" gorm:"not null;default:'pending'"`
	DocumentURL         string          `json:"document_url"`
	VerificationStatus  VerificationStatus `json:"verification_status" gorm:"default:'pending'"`
	AIRiskScore         float64         `json:"ai_risk_score" gorm:"default:0"`
	BlockchainTxHash    string          `json:"blockchain_tx_hash"`
	TokenID             string          `json:"token_id"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	DeletedAt           gorm.DeletedAt  `json:"-" gorm:"index"`
	
	// Relationships
	User                User            `json:"user" gorm:"foreignKey:UserID"`
	FinancingRequests   []FinancingRequest `json:"financing_requests"`
}

type InvoiceStatus string

const (
	InvoiceStatusPending   InvoiceStatus = "pending"
	InvoiceStatusVerified  InvoiceStatus = "verified"
	InvoiceStatusFinanced  InvoiceStatus = "financed"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusRejected  InvoiceStatus = "rejected"
)

type VerificationStatus string

const (
	VerificationPending   VerificationStatus = "pending"
	VerificationApproved  VerificationStatus = "approved"
	VerificationRejected  VerificationStatus = "rejected"
	VerificationRequiresReview VerificationStatus = "requires_review"
)

type FinancingRequest struct {
	ID                uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	InvoiceID         uuid.UUID         `json:"invoice_id" gorm:"type:uuid;not null"`
	UserID            uuid.UUID         `json:"user_id" gorm:"type:uuid;not null"`
	RequestedAmount   float64           `json:"requested_amount" gorm:"not null"`
	InterestRate      float64           `json:"interest_rate" gorm:"not null"`
	FinancingFee      float64           `json:"financing_fee" gorm:"not null"`
	NetAmount         float64           `json:"net_amount" gorm:"not null"`
	Status            FinancingStatus   `json:"status" gorm:"not null;default:'pending'"`
	Description       string            `json:"description"`
	ExpectedReturn    float64           `json:"expected_return"`
	RiskLevel         RiskLevel         `json:"risk_level" gorm:"not null"`
	ApprovedAt        *time.Time        `json:"approved_at"`
	FundedAt          *time.Time        `json:"funded_at"`
	CompletedAt       *time.Time        `json:"completed_at"`
	BlockchainTxHash  string            `json:"blockchain_tx_hash"`
	SmartContractAddr string            `json:"smart_contract_addr"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt    `json:"-" gorm:"index"`
	
	// Relationships
	Invoice           Invoice           `json:"invoice" gorm:"foreignKey:InvoiceID"`
	User              User              `json:"user" gorm:"foreignKey:UserID"`
	Investments       []Investment      `json:"investments"`
}

type FinancingStatus string

const (
	FinancingStatusPending   FinancingStatus = "pending"
	FinancingStatusApproved  FinancingStatus = "approved"
	FinancingStatusFunded    FinancingStatus = "funded"
	FinancingStatusCompleted FinancingStatus = "completed"
	FinancingStatusRejected  FinancingStatus = "rejected"
	FinancingStatusExpired   FinancingStatus = "expired"
)

type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

type Investment struct {
	ID                 uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FinancingRequestID uuid.UUID         `json:"financing_request_id" gorm:"type:uuid;not null"`
	InvestorID         uuid.UUID         `json:"investor_id" gorm:"type:uuid;not null"`
	Amount             float64           `json:"amount" gorm:"not null"`
	ExpectedReturn     float64           `json:"expected_return" gorm:"not null"`
	ActualReturn       float64           `json:"actual_return" gorm:"default:0"`
	Status             InvestmentStatus  `json:"status" gorm:"not null;default:'pending'"`
	InvestmentDate     time.Time         `json:"investment_date"`
	MaturityDate       time.Time         `json:"maturity_date"`
	ReturnDate         *time.Time        `json:"return_date"`
	BlockchainTxHash   string            `json:"blockchain_tx_hash"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	DeletedAt          gorm.DeletedAt    `json:"-" gorm:"index"`
	
	// Relationships
	FinancingRequest   FinancingRequest  `json:"financing_request" gorm:"foreignKey:FinancingRequestID"`
	Investor           User              `json:"investor" gorm:"foreignKey:InvestorID"`
}

type InvestmentStatus string

const (
	InvestmentStatusPending   InvestmentStatus = "pending"
	InvestmentStatusActive    InvestmentStatus = "active"
	InvestmentStatusCompleted InvestmentStatus = "completed"
	InvestmentStatusDefaulted InvestmentStatus = "defaulted"
)

type Transaction struct {
	ID                uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID            uuid.UUID        `json:"user_id" gorm:"type:uuid;not null"`
	Type              TransactionType  `json:"type" gorm:"not null"`
	Amount            float64          `json:"amount" gorm:"not null"`
	Status            TransactionStatus `json:"status" gorm:"not null;default:'pending'"`
	Description       string           `json:"description"`
	Reference         string           `json:"reference"`
	BlockchainTxHash  string           `json:"blockchain_tx_hash"`
	GasUsed           int64            `json:"gas_used"`
	GasPrice          int64            `json:"gas_price"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	
	// Relationships
	User              User             `json:"user" gorm:"foreignKey:UserID"`
}

type TransactionType string

const (
	TransactionTypeInvestment TransactionType = "investment"
	TransactionTypeFinancing  TransactionType = "financing"
	TransactionTypeRepayment  TransactionType = "repayment"
	TransactionTypeFee        TransactionType = "fee"
	TransactionTypeWithdraw   TransactionType = "withdraw"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusConfirmed TransactionStatus = "confirmed"
	TransactionStatusFailed    TransactionStatus = "failed"
)
