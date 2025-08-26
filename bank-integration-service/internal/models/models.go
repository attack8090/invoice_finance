package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BankConnection represents a connection to a bank's API
type BankConnection struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	BankCode       string    `gorm:"type:varchar(50);not null" json:"bank_code"`
	BankName       string    `gorm:"type:varchar(255);not null" json:"bank_name"`
	ConnectionType string    `gorm:"type:varchar(50);default:'api'" json:"connection_type"` // api, webhook, file
	Status         string    `gorm:"type:varchar(50);default:'pending'" json:"status"`       // pending, active, inactive, error
	APICredentials string    `gorm:"type:text;encrypted" json:"-"`                          // Encrypted credentials
	TestMode       bool      `gorm:"default:true" json:"test_mode"`
	LastSyncAt     *time.Time `json:"last_sync_at"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreditDecision represents a credit decision from a bank
type CreditDecision struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CustomerID       uuid.UUID `gorm:"type:uuid;not null" json:"customer_id"`
	BankConnectionID uuid.UUID `gorm:"type:uuid;not null" json:"bank_connection_id"`
	RequestType      string    `gorm:"type:varchar(50);not null" json:"request_type"` // credit_line, loan, trade_finance
	RequestedAmount  float64   `gorm:"type:decimal(15,2)" json:"requested_amount"`
	ApprovedAmount   float64   `gorm:"type:decimal(15,2)" json:"approved_amount"`
	InterestRate     float64   `gorm:"type:decimal(5,2)" json:"interest_rate"`
	Term             int       `json:"term"` // Term in days
	Status           string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, approved, rejected, expired
	Decision         string    `gorm:"type:text" json:"decision"`
	Conditions       string    `gorm:"type:text" json:"conditions"`
	RiskScore        float64   `gorm:"type:decimal(5,2)" json:"risk_score"`
	RiskFactors      string    `gorm:"type:json" json:"risk_factors"`
	DecisionDate     *time.Time `json:"decision_date"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	BankConnection BankConnection `gorm:"foreignKey:BankConnectionID" json:"bank_connection,omitempty"`
}

// PaymentTransaction represents a payment processed through bank APIs
type PaymentTransaction struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BankConnectionID   uuid.UUID `gorm:"type:uuid;not null" json:"bank_connection_id"`
	PaymentID          string    `gorm:"type:varchar(255);unique" json:"payment_id"`
	ExternalPaymentID  string    `gorm:"type:varchar(255)" json:"external_payment_id"`
	Type               string    `gorm:"type:varchar(50);not null" json:"type"` // transfer, payment, settlement
	FromAccountID      string    `gorm:"type:varchar(255)" json:"from_account_id"`
	ToAccountID        string    `gorm:"type:varchar(255)" json:"to_account_id"`
	Amount             float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Currency           string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Status             string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, processing, completed, failed, cancelled
	Description        string    `gorm:"type:text" json:"description"`
	Reference          string    `gorm:"type:varchar(255)" json:"reference"`
	Fees               float64   `gorm:"type:decimal(15,2);default:0" json:"fees"`
	ProcessedAt        *time.Time `json:"processed_at"`
	FailureReason      string     `gorm:"type:text" json:"failure_reason,omitempty"`
	RetryCount         int        `gorm:"default:0" json:"retry_count"`
	Epic4ComplianceData string    `gorm:"type:json" json:"epic4_compliance_data"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	BankConnection BankConnection `gorm:"foreignKey:BankConnectionID" json:"bank_connection,omitempty"`
}

// FinancingRequest represents a financing request to banks
type FinancingRequest struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CustomerID           uuid.UUID `gorm:"type:uuid;not null" json:"customer_id"`
	RequestType          string    `gorm:"type:varchar(50);not null" json:"request_type"` // invoice_finance, purchase_order_finance, working_capital
	RequestedAmount      float64   `gorm:"type:decimal(15,2);not null" json:"requested_amount"`
	Currency             string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Term                 int       `json:"term"` // Term in days
	Purpose              string    `gorm:"type:text" json:"purpose"`
	Status               string    `gorm:"type:varchar(50);default:'draft'" json:"status"` // draft, submitted, under_review, approved, rejected, disbursed
	SubmittedToBanks     []string  `gorm:"type:json" json:"submitted_to_banks"`
	ApprovedByBanks      []string  `gorm:"type:json" json:"approved_by_banks"`
	BestOffer            string    `gorm:"type:json" json:"best_offer"`
	Documents            []string  `gorm:"type:json" json:"documents"`
	CollateralDetails    string    `gorm:"type:json" json:"collateral_details"`
	BusinessDetails      string    `gorm:"type:json" json:"business_details"`
	FinancialInformation string    `gorm:"type:json" json:"financial_information"`
	RequestedAt          time.Time `json:"requested_at"`
	ReviewedAt           *time.Time `json:"reviewed_at"`
	ApprovedAt           *time.Time `json:"approved_at"`
	DisbursedAt          *time.Time `json:"disbursed_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// PortfolioItem represents an item in a bank's financing portfolio
type PortfolioItem struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BankConnectionID    uuid.UUID `gorm:"type:uuid;not null" json:"bank_connection_id"`
	CustomerID          uuid.UUID `gorm:"type:uuid;not null" json:"customer_id"`
	FinancingRequestID  uuid.UUID `gorm:"type:uuid" json:"financing_request_id"`
	Type                string    `gorm:"type:varchar(50);not null" json:"type"` // invoice, purchase_order, credit_line
	Principal           float64   `gorm:"type:decimal(15,2);not null" json:"principal"`
	Outstanding         float64   `gorm:"type:decimal(15,2);not null" json:"outstanding"`
	InterestRate        float64   `gorm:"type:decimal(5,2)" json:"interest_rate"`
	MaturityDate        time.Time `json:"maturity_date"`
	Status              string    `gorm:"type:varchar(50);default:'active'" json:"status"` // active, matured, defaulted, written_off
	RiskRating          string    `gorm:"type:varchar(10)" json:"risk_rating"`
	LastPaymentDate     *time.Time `json:"last_payment_date"`
	NextPaymentDue      *time.Time `json:"next_payment_due"`
	PaymentHistory      string     `gorm:"type:json" json:"payment_history"`
	PerformanceMetrics  string     `gorm:"type:json" json:"performance_metrics"`
	ComplianceStatus    string     `gorm:"type:varchar(50);default:'compliant'" json:"compliance_status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	BankConnection     BankConnection     `gorm:"foreignKey:BankConnectionID" json:"bank_connection,omitempty"`
	FinancingRequest   FinancingRequest   `gorm:"foreignKey:FinancingRequestID" json:"financing_request,omitempty"`
}

// FundingSource represents available funding sources from banks
type FundingSource struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BankConnectionID uuid.UUID `gorm:"type:uuid;not null" json:"bank_connection_id"`
	Name             string    `gorm:"type:varchar(255);not null" json:"name"`
	Type             string    `gorm:"type:varchar(50);not null" json:"type"` // credit_line, term_loan, revolving_facility
	TotalCapacity    float64   `gorm:"type:decimal(15,2);not null" json:"total_capacity"`
	AvailableCapacity float64  `gorm:"type:decimal(15,2);not null" json:"available_capacity"`
	UtilizedCapacity float64   `gorm:"type:decimal(15,2);default:0" json:"utilized_capacity"`
	InterestRate     float64   `gorm:"type:decimal(5,2)" json:"interest_rate"`
	Currency         string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Status           string    `gorm:"type:varchar(50);default:'active'" json:"status"` // active, suspended, expired
	Terms            string    `gorm:"type:json" json:"terms"`
	Restrictions     string    `gorm:"type:json" json:"restrictions"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	LastRefreshed    time.Time  `json:"last_refreshed"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	BankConnection BankConnection `gorm:"foreignKey:BankConnectionID" json:"bank_connection,omitempty"`
}

// ComplianceRecord represents Epic 4 compliance records
type ComplianceRecord struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EntityType      string    `gorm:"type:varchar(50);not null" json:"entity_type"` // payment, transaction, financing
	EntityID        uuid.UUID `gorm:"type:uuid;not null" json:"entity_id"`
	ComplianceType  string    `gorm:"type:varchar(50);not null" json:"compliance_type"` // epic4, aml, kyc, sanctions
	Status          string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, compliant, non_compliant, under_review
	CheckedAt       time.Time `json:"checked_at"`
	ComplianceData  string    `gorm:"type:json" json:"compliance_data"`
	Issues          string    `gorm:"type:json" json:"issues"`
	Remediation     string    `gorm:"type:text" json:"remediation"`
	ExpiryDate      *time.Time `json:"expiry_date"`
	ReportedToAuthority bool   `gorm:"default:false" json:"reported_to_authority"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// AuditTrail represents audit trail entries for Epic 4 compliance
type AuditTrail struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EntityType  string    `gorm:"type:varchar(50);not null" json:"entity_type"`
	EntityID    uuid.UUID `gorm:"type:uuid;not null" json:"entity_id"`
	Action      string    `gorm:"type:varchar(100);not null" json:"action"`
	UserID      uuid.UUID `gorm:"type:uuid" json:"user_id"`
	IPAddress   string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent   string    `gorm:"type:text" json:"user_agent"`
	Details     string    `gorm:"type:json" json:"details"`
	BeforeState string    `gorm:"type:json" json:"before_state"`
	AfterState  string    `gorm:"type:json" json:"after_state"`
	Timestamp   time.Time `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
}

// BankAccount represents bank account information
type BankAccount struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BankConnectionID uuid.UUID `gorm:"type:uuid;not null" json:"bank_connection_id"`
	CustomerID       uuid.UUID `gorm:"type:uuid;not null" json:"customer_id"`
	AccountNumber    string    `gorm:"type:varchar(50);encrypted" json:"-"` // Encrypted
	RoutingNumber    string    `gorm:"type:varchar(20)" json:"routing_number"`
	AccountType      string    `gorm:"type:varchar(50)" json:"account_type"` // checking, savings, business
	AccountName      string    `gorm:"type:varchar(255)" json:"account_name"`
	Currency         string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Status           string    `gorm:"type:varchar(50);default:'active'" json:"status"` // active, inactive, closed, verification_pending
	Balance          float64   `gorm:"type:decimal(15,2)" json:"balance,omitempty"`
	AvailableBalance float64   `gorm:"type:decimal(15,2)" json:"available_balance,omitempty"`
	LastBalanceCheck *time.Time `json:"last_balance_check"`
	IsVerified       bool       `gorm:"default:false" json:"is_verified"`
	VerificationData string     `gorm:"type:json" json:"verification_data"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	BankConnection BankConnection `gorm:"foreignKey:BankConnectionID" json:"bank_connection,omitempty"`
}

// FundingMatching represents funding matching records
type FundingMatching struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FinancingRequestID uuid.UUID `gorm:"type:uuid;not null" json:"financing_request_id"`
	FundingSourceID    uuid.UUID `gorm:"type:uuid;not null" json:"funding_source_id"`
	MatchedAmount      float64   `gorm:"type:decimal(15,2);not null" json:"matched_amount"`
	MatchScore         float64   `gorm:"type:decimal(5,2)" json:"match_score"`
	Status             string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, confirmed, rejected, expired
	MatchingReason     string    `gorm:"type:text" json:"matching_reason"`
	MatchingCriteria   string    `gorm:"type:json" json:"matching_criteria"`
	MatchedAt          time.Time `json:"matched_at"`
	ConfirmedAt        *time.Time `json:"confirmed_at"`
	ExpiresAt          time.Time  `json:"expires_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	FinancingRequest FinancingRequest `gorm:"foreignKey:FinancingRequestID" json:"financing_request,omitempty"`
	FundingSource    FundingSource    `gorm:"foreignKey:FundingSourceID" json:"funding_source,omitempty"`
}

// RiskAssessment represents risk assessment data
type RiskAssessment struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EntityType       string    `gorm:"type:varchar(50);not null" json:"entity_type"` // customer, transaction, portfolio_item
	EntityID         uuid.UUID `gorm:"type:uuid;not null" json:"entity_id"`
	AssessmentType   string    `gorm:"type:varchar(50);not null" json:"assessment_type"` // credit, operational, market, liquidity
	RiskScore        float64   `gorm:"type:decimal(5,2)" json:"risk_score"`
	RiskRating       string    `gorm:"type:varchar(10)" json:"risk_rating"` // low, medium, high, critical
	RiskFactors      string    `gorm:"type:json" json:"risk_factors"`
	Recommendations  string    `gorm:"type:json" json:"recommendations"`
	AssessedBy       string    `gorm:"type:varchar(100)" json:"assessed_by"` // system, human, hybrid
	AssessmentModel  string    `gorm:"type:varchar(100)" json:"assessment_model"`
	Confidence       float64   `gorm:"type:decimal(5,2)" json:"confidence"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	AssessedAt       time.Time  `json:"assessed_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ReconciliationJob represents payment reconciliation jobs
type ReconciliationJob struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	JobType       string    `gorm:"type:varchar(50);not null" json:"job_type"` // daily, weekly, monthly, on_demand
	Status        string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, running, completed, failed
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	TotalRecords  int       `json:"total_records"`
	ProcessedRecords int    `json:"processed_records"`
	MatchedRecords   int    `json:"matched_records"`
	UnmatchedRecords int    `json:"unmatched_records"`
	ErrorRecords     int    `json:"error_records"`
	Summary          string `gorm:"type:json" json:"summary"`
	ErrorDetails     string `gorm:"type:json" json:"error_details"`
	StartedAt        *time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// BeforeCreate hooks for UUID generation
func (bc *BankConnection) BeforeCreate(tx *gorm.DB) error {
	if bc.ID == uuid.Nil {
		bc.ID = uuid.New()
	}
	return nil
}

func (cd *CreditDecision) BeforeCreate(tx *gorm.DB) error {
	if cd.ID == uuid.Nil {
		cd.ID = uuid.New()
	}
	return nil
}

func (pt *PaymentTransaction) BeforeCreate(tx *gorm.DB) error {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	return nil
}

func (fr *FinancingRequest) BeforeCreate(tx *gorm.DB) error {
	if fr.ID == uuid.Nil {
		fr.ID = uuid.New()
	}
	return nil
}

func (pi *PortfolioItem) BeforeCreate(tx *gorm.DB) error {
	if pi.ID == uuid.Nil {
		pi.ID = uuid.New()
	}
	return nil
}

func (fs *FundingSource) BeforeCreate(tx *gorm.DB) error {
	if fs.ID == uuid.Nil {
		fs.ID = uuid.New()
	}
	return nil
}

func (cr *ComplianceRecord) BeforeCreate(tx *gorm.DB) error {
	if cr.ID == uuid.Nil {
		cr.ID = uuid.New()
	}
	return nil
}

func (at *AuditTrail) BeforeCreate(tx *gorm.DB) error {
	if at.ID == uuid.Nil {
		at.ID = uuid.New()
	}
	return nil
}

func (ba *BankAccount) BeforeCreate(tx *gorm.DB) error {
	if ba.ID == uuid.Nil {
		ba.ID = uuid.New()
	}
	return nil
}

func (fm *FundingMatching) BeforeCreate(tx *gorm.DB) error {
	if fm.ID == uuid.Nil {
		fm.ID = uuid.New()
	}
	return nil
}

func (ra *RiskAssessment) BeforeCreate(tx *gorm.DB) error {
	if ra.ID == uuid.Nil {
		ra.ID = uuid.New()
	}
	return nil
}

func (rj *ReconciliationJob) BeforeCreate(tx *gorm.DB) error {
	if rj.ID == uuid.Nil {
		rj.ID = uuid.New()
	}
	return nil
}
