package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UUID              uuid.UUID          `json:"uuid" bson:"uuid"`
	Email             string             `json:"email" bson:"email"`
	PasswordHash      string             `json:"-" bson:"password_hash"`
	FirstName         string             `json:"first_name" bson:"first_name"`
	LastName          string             `json:"last_name" bson:"last_name"`
	Role              UserRole           `json:"role" bson:"role"`
	CompanyName       string             `json:"company_name" bson:"company_name"`
	TaxID             string             `json:"tax_id" bson:"tax_id"`
	WalletAddress     string             `json:"wallet_address" bson:"wallet_address"`
	IsVerified        bool               `json:"is_verified" bson:"is_verified"`
	CreditScore       int                `json:"credit_score" bson:"credit_score"`
	TotalInvestment   float64            `json:"total_investment" bson:"total_investment"`
	TotalFinanced     float64            `json:"total_financed" bson:"total_financed"`
	ProfileCompleted  bool               `json:"profile_completed" bson:"profile_completed"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt         *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

type UserRole string

const (
	RoleSME      UserRole = "sme"
	RoleInvestor UserRole = "investor"
	RoleAdmin    UserRole = "admin"
)

type Invoice struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UUID                uuid.UUID          `json:"uuid" bson:"uuid"`
	UserID              uuid.UUID          `json:"user_id" bson:"user_id"`
	InvoiceNumber       string             `json:"invoice_number" bson:"invoice_number"`
	CustomerName        string             `json:"customer_name" bson:"customer_name"`
	CustomerEmail       string             `json:"customer_email" bson:"customer_email"`
	InvoiceAmount       float64            `json:"invoice_amount" bson:"invoice_amount"`
	DueDate             time.Time          `json:"due_date" bson:"due_date"`
	IssueDate           time.Time          `json:"issue_date" bson:"issue_date"`
	Description         string             `json:"description" bson:"description"`
	Status              InvoiceStatus      `json:"status" bson:"status"`
	DocumentURL         string             `json:"document_url" bson:"document_url"`
	VerificationStatus  VerificationStatus `json:"verification_status" bson:"verification_status"`
	AIRiskScore         float64            `json:"ai_risk_score" bson:"ai_risk_score"`
	FabricTxID          string             `json:"fabric_tx_id" bson:"fabric_tx_id"`
	AssetID             string             `json:"asset_id" bson:"asset_id"`
	CreatedAt           time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt           *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
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
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UUID              uuid.UUID          `json:"uuid" bson:"uuid"`
	InvoiceID         uuid.UUID          `json:"invoice_id" bson:"invoice_id"`
	UserID            uuid.UUID          `json:"user_id" bson:"user_id"`
	RequestedAmount   float64            `json:"requested_amount" bson:"requested_amount"`
	InterestRate      float64            `json:"interest_rate" bson:"interest_rate"`
	FinancingFee      float64            `json:"financing_fee" bson:"financing_fee"`
	NetAmount         float64            `json:"net_amount" bson:"net_amount"`
	Status            FinancingStatus    `json:"status" bson:"status"`
	Description       string             `json:"description" bson:"description"`
	ExpectedReturn    float64            `json:"expected_return" bson:"expected_return"`
	RiskLevel         RiskLevel          `json:"risk_level" bson:"risk_level"`
	ApprovedAt        *time.Time         `json:"approved_at" bson:"approved_at,omitempty"`
	FundedAt          *time.Time         `json:"funded_at" bson:"funded_at,omitempty"`
	CompletedAt       *time.Time         `json:"completed_at" bson:"completed_at,omitempty"`
	FabricTxID        string             `json:"fabric_tx_id" bson:"fabric_tx_id"`
	FabricAssetID     string             `json:"fabric_asset_id" bson:"fabric_asset_id"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt         *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
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
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UUID               uuid.UUID          `json:"uuid" bson:"uuid"`
	FinancingRequestID uuid.UUID          `json:"financing_request_id" bson:"financing_request_id"`
	InvestorID         uuid.UUID          `json:"investor_id" bson:"investor_id"`
	Amount             float64            `json:"amount" bson:"amount"`
	ExpectedReturn     float64            `json:"expected_return" bson:"expected_return"`
	ActualReturn       float64            `json:"actual_return" bson:"actual_return"`
	Status             InvestmentStatus   `json:"status" bson:"status"`
	InvestmentDate     time.Time          `json:"investment_date" bson:"investment_date"`
	MaturityDate       time.Time          `json:"maturity_date" bson:"maturity_date"`
	ReturnDate         *time.Time         `json:"return_date" bson:"return_date,omitempty"`
	FabricTxID         string             `json:"fabric_tx_id" bson:"fabric_tx_id"`
	CreatedAt          time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt          *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

type InvestmentStatus string

const (
	InvestmentStatusPending   InvestmentStatus = "pending"
	InvestmentStatusActive    InvestmentStatus = "active"
	InvestmentStatusCompleted InvestmentStatus = "completed"
	InvestmentStatusDefaulted InvestmentStatus = "defaulted"
)

type Transaction struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UUID              uuid.UUID          `json:"uuid" bson:"uuid"`
	UserID            uuid.UUID          `json:"user_id" bson:"user_id"`
	Type              TransactionType    `json:"type" bson:"type"`
	Amount            float64            `json:"amount" bson:"amount"`
	Status            TransactionStatus  `json:"status" bson:"status"`
	Description       string             `json:"description" bson:"description"`
	Reference         string             `json:"reference" bson:"reference"`
	FabricTxID        string             `json:"fabric_tx_id" bson:"fabric_tx_id"`
	FabricBlockNum    int64              `json:"fabric_block_num" bson:"fabric_block_num"`
	FabricChannelName string             `json:"fabric_channel_name" bson:"fabric_channel_name"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt         *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
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
