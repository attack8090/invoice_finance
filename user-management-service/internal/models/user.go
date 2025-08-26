package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents different user types in the system
type UserRole string

const (
	RoleSME   UserRole = "sme"   // Small/Medium Enterprise
	RoleBuyer UserRole = "buyer" // Invoice buyer/customer
	RoleBank  UserRole = "bank"  // Financial institution
	RoleAdmin UserRole = "admin" // Platform administrator
)

// UserStatus represents the current status of a user account
type UserStatus string

const (
	StatusPending    UserStatus = "pending"     // Account created, email verification pending
	StatusActive     UserStatus = "active"      // Account active and verified
	StatusSuspended  UserStatus = "suspended"   // Account temporarily suspended
	StatusDeactivated UserStatus = "deactivated" // Account permanently deactivated
	StatusKYCPending UserStatus = "kyc_pending" // KYC verification pending
)

// KYCStatus represents the KYC verification status
type KYCStatus string

const (
	KYCNotStarted KYCStatus = "not_started" // KYC process not initiated
	KYCPending    KYCStatus = "pending"     // KYC documents submitted, under review
	KYCApproved   KYCStatus = "approved"    // KYC approved
	KYCRejected   KYCStatus = "rejected"    // KYC rejected, requires resubmission
	KYCExpired    KYCStatus = "expired"     // KYC approval expired, requires renewal
)

// User represents the main user entity
type User struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email     string     `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Phone     string     `json:"phone" gorm:"uniqueIndex" validate:"required,min=10,max=15"`
	Password  string     `json:"-" gorm:"not null" validate:"required,min=8"`
	Role      UserRole   `json:"role" gorm:"not null" validate:"required,oneof=sme buyer bank admin"`
	Status    UserStatus `json:"status" gorm:"default:pending"`
	
	// Profile information
	FirstName    string `json:"first_name" validate:"required,min=2,max=50"`
	LastName     string `json:"last_name" validate:"required,min=2,max=50"`
	ProfileImage string `json:"profile_image,omitempty"`
	
	// Account verification
	EmailVerified     bool       `json:"email_verified" gorm:"default:false"`
	EmailVerifiedAt   *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerified     bool       `json:"phone_verified" gorm:"default:false"`
	PhoneVerifiedAt   *time.Time `json:"phone_verified_at,omitempty"`
	
	// Security
	MFAEnabled        bool       `json:"mfa_enabled" gorm:"default:false"`
	MFASecret         string     `json:"-"` // TOTP secret
	BackupCodes       []string   `json:"-" gorm:"type:text[]"` // Recovery codes
	PasswordChangedAt *time.Time `json:"password_changed_at,omitempty"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP       string     `json:"last_login_ip,omitempty"`
	
	// Compliance and KYC
	KYCStatus           KYCStatus  `json:"kyc_status" gorm:"default:not_started"`
	KYCCompletedAt      *time.Time `json:"kyc_completed_at,omitempty"`
	KYCExpiresAt        *time.Time `json:"kyc_expires_at,omitempty"`
	ComplianceScore     int        `json:"compliance_score" gorm:"default:0"` // 0-100
	RiskRating          string     `json:"risk_rating,omitempty"`             // low, medium, high
	
	// Relationships
	Company       *Company       `json:"company,omitempty" gorm:"foreignKey:UserID"`
	KYCData       *KYCData       `json:"kyc_data,omitempty" gorm:"foreignKey:UserID"`
	Sessions      []UserSession  `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	LoginHistory  []LoginHistory `json:"login_history,omitempty" gorm:"foreignKey:UserID"`
	
	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	CreatedBy *uuid.UUID     `json:"created_by,omitempty"` // Admin who created the account
	UpdatedBy *uuid.UUID     `json:"updated_by,omitempty"` // Last admin who updated the account
}

// Company represents business entity information
type Company struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID `json:"user_id" gorm:"not null"`
	
	// Basic company information
	Name            string `json:"name" validate:"required,min=2,max=200"`
	LegalName       string `json:"legal_name" validate:"required,min=2,max=200"`
	RegistrationNum string `json:"registration_number" gorm:"uniqueIndex" validate:"required"`
	TaxID           string `json:"tax_id" gorm:"uniqueIndex" validate:"required"`
	
	// Company details
	Industry        string    `json:"industry" validate:"required"`
	FoundedYear     int       `json:"founded_year" validate:"min=1800,max=2024"`
	EmployeeCount   int       `json:"employee_count"`
	AnnualRevenue   float64   `json:"annual_revenue"`
	Website         string    `json:"website,omitempty" validate:"omitempty,url"`
	Description     string    `json:"description,omitempty" validate:"max=1000"`
	
	// Address information
	AddressLine1 string `json:"address_line_1" validate:"required"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	City         string `json:"city" validate:"required"`
	State        string `json:"state" validate:"required"`
	PostalCode   string `json:"postal_code" validate:"required"`
	Country      string `json:"country" validate:"required,len=2"` // ISO country code
	
	// Financial information
	BankName       string `json:"bank_name,omitempty"`
	AccountNumber  string `json:"account_number,omitempty"`
	RoutingNumber  string `json:"routing_number,omitempty"`
	CreditLimit    float64 `json:"credit_limit" gorm:"default:0"`
	CreditUsed     float64 `json:"credit_used" gorm:"default:0"`
	
	// Verification status
	Verified         bool       `json:"verified" gorm:"default:false"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	
	// Relationships
	User      User                `json:"user" gorm:"foreignKey:UserID"`
	Documents []CompanyDocument   `json:"documents,omitempty" gorm:"foreignKey:CompanyID"`
	
	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// KYCData represents Know Your Customer verification data
type KYCData struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID uuid.UUID `json:"user_id" gorm:"not null;uniqueIndex"`
	
	// Personal information
	DateOfBirth  time.Time `json:"date_of_birth" validate:"required"`
	Nationality  string    `json:"nationality" validate:"required,len=2"` // ISO country code
	IDType       string    `json:"id_type" validate:"required,oneof=passport driving_license national_id"`
	IDNumber     string    `json:"id_number" validate:"required"`
	IDExpiryDate time.Time `json:"id_expiry_date" validate:"required"`
	
	// Address verification
	AddressLine1     string `json:"address_line_1" validate:"required"`
	AddressLine2     string `json:"address_line_2,omitempty"`
	City             string `json:"city" validate:"required"`
	State            string `json:"state" validate:"required"`
	PostalCode       string `json:"postal_code" validate:"required"`
	Country          string `json:"country" validate:"required,len=2"`
	AddressProofType string `json:"address_proof_type" validate:"required,oneof=utility_bill bank_statement lease_agreement"`
	
	// Enhanced due diligence for high-risk users
	PoliticallyExposed   bool   `json:"politically_exposed" gorm:"default:false"`
	SourceOfFunds        string `json:"source_of_funds,omitempty"`
	PurposeOfAccount     string `json:"purpose_of_account,omitempty"`
	ExpectedTransVolume  string `json:"expected_transaction_volume,omitempty"`
	
	// AML checks
	AMLCheckStatus     string     `json:"aml_check_status" gorm:"default:pending"` // pending, passed, failed
	AMLCheckDate       *time.Time `json:"aml_check_date,omitempty"`
	SanctionsListCheck bool       `json:"sanctions_list_check" gorm:"default:false"`
	WatchlistCheck     bool       `json:"watchlist_check" gorm:"default:false"`
	
	// Document references
	IDDocumentURL       string `json:"id_document_url,omitempty"`
	AddressProofURL     string `json:"address_proof_url,omitempty"`
	SelfieURL           string `json:"selfie_url,omitempty"`
	AdditionalDocsURL   string `json:"additional_docs_url,omitempty"`
	
	// Review information
	ReviewStatus   KYCStatus  `json:"review_status" gorm:"default:pending"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy     *uuid.UUID `json:"reviewed_by,omitempty"` // Admin who reviewed
	ReviewComments string     `json:"review_comments,omitempty"`
	RejectionReason string    `json:"rejection_reason,omitempty"`
	
	// Compliance scores
	RiskScore        int    `json:"risk_score" gorm:"default:0"`        // 0-100
	ComplianceLevel  string `json:"compliance_level,omitempty"`         // basic, enhanced, premium
	NextReviewDate   time.Time `json:"next_review_date,omitempty"`
	
	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
	
	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// CompanyDocument represents uploaded company documents
type CompanyDocument struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID uuid.UUID `json:"company_id" gorm:"not null"`
	
	// Document information
	Name         string `json:"name" validate:"required"`
	Type         string `json:"type" validate:"required,oneof=certificate_of_incorporation tax_certificate bank_statement financial_report other"`
	Description  string `json:"description,omitempty"`
	FileURL      string `json:"file_url" validate:"required"`
	FileName     string `json:"file_name" validate:"required"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	
	// Verification
	Verified     bool       `json:"verified" gorm:"default:false"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	VerifiedBy   *uuid.UUID `json:"verified_by,omitempty"`
	
	// Relationships
	Company Company `json:"company" gorm:"foreignKey:CompanyID"`
	
	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// UserSession represents active user sessions
type UserSession struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"not null"`
	
	// Session details
	Token         string    `json:"-" gorm:"uniqueIndex;not null"`
	RefreshToken  string    `json:"-" gorm:"uniqueIndex;not null"`
	DeviceInfo    string    `json:"device_info,omitempty"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent,omitempty"`
	Location      string    `json:"location,omitempty"`
	
	// Session timing
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	
	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// LoginHistory tracks user login attempts
type LoginHistory struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    *uuid.UUID `json:"user_id,omitempty"` // Can be null for failed attempts
	
	// Attempt details
	Email       string    `json:"email"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Success     bool      `json:"success"`
	FailReason  string    `json:"fail_reason,omitempty"`
	AttemptedAt time.Time `json:"attempted_at"`
	Location    string    `json:"location,omitempty"`
	
	// Security flags
	SuspiciousActivity bool   `json:"suspicious_activity" gorm:"default:false"`
	BlockedReason      string `json:"blocked_reason,omitempty"`
	
	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	
	// Audit fields
	CreatedAt time.Time `json:"created_at"`
}

// ComplianceReport represents compliance audit reports
type ComplianceReport struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	
	// Report details
	Type        string    `json:"type" validate:"required,oneof=kyc_audit user_activity transaction_monitoring risk_assessment"`
	Status      string    `json:"status" gorm:"default:pending"` // pending, completed, failed
	Period      string    `json:"period"`  // reporting period
	GeneratedBy uuid.UUID `json:"generated_by"`
	
	// Report data
	Summary     string                 `json:"summary,omitempty"`
	Data        map[string]interface{} `json:"data" gorm:"type:jsonb"`
	Findings    []string               `json:"findings" gorm:"type:text[]"`
	Actions     []string               `json:"actions" gorm:"type:text[]"`
	
	// File reference
	ReportURL string `json:"report_url,omitempty"`
	
	// Audit fields
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// BeforeCreate hooks for setting up default values
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (k *KYCData) BeforeCreate(tx *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return nil
}

// Helper methods
func (u *User) IsKYCCompliant() bool {
	return u.KYCStatus == KYCApproved && 
		   (u.KYCExpiresAt == nil || u.KYCExpiresAt.After(time.Now()))
}

func (u *User) RequiresKYCRenewal() bool {
	return u.KYCExpiresAt != nil && u.KYCExpiresAt.Before(time.Now().AddDate(0, 0, 30))
}

func (c *Company) GetCreditAvailable() float64 {
	return c.CreditLimit - c.CreditUsed
}

func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}
