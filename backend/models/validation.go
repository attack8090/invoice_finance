package models

import "time"

// User validation models
type UserRegistrationRequest struct {
	Email           string `json:"email" binding:"required,email,max=255"`
	Password        string `json:"password" binding:"required,password,min=8,max=128"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	CompanyName     string `json:"company_name" binding:"required,company_name,min=2,max=100"`
	Phone           string `json:"phone" binding:"omitempty,phone"`
	Role            string `json:"role" binding:"required,user_role"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=1"`
}

type UserUpdateRequest struct {
	CompanyName   string  `json:"company_name" binding:"omitempty,company_name,min=2,max=100"`
	Phone         string  `json:"phone" binding:"omitempty,phone"`
	Address       string  `json:"address" binding:"omitempty,max=500"`
	CreditScore   *int    `json:"credit_score" binding:"omitempty,min=300,max=850"`
	AnnualRevenue *float64 `json:"annual_revenue" binding:"omitempty,min=0,amount"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=1"`
	NewPassword     string `json:"new_password" binding:"required,password,min=8,max=128"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// Invoice validation models
type InvoiceCreateRequest struct {
	InvoiceNumber   string    `json:"invoice_number" binding:"required,invoice_number"`
	CustomerName    string    `json:"customer_name" binding:"required,min=2,max=100"`
	CustomerEmail   string    `json:"customer_email" binding:"omitempty,email"`
	InvoiceAmount   float64   `json:"invoice_amount" binding:"required,amount,gt=0"`
	DueDate         string    `json:"due_date" binding:"required,future_date"`
	Description     string    `json:"description" binding:"omitempty,max=1000"`
	Terms           string    `json:"terms" binding:"omitempty,max=2000"`
	TaxAmount       *float64  `json:"tax_amount" binding:"omitempty,min=0"`
	DiscountAmount  *float64  `json:"discount_amount" binding:"omitempty,min=0"`
}

type InvoiceUpdateRequest struct {
	CustomerName    string   `json:"customer_name" binding:"omitempty,min=2,max=100"`
	CustomerEmail   string   `json:"customer_email" binding:"omitempty,email"`
	InvoiceAmount   *float64 `json:"invoice_amount" binding:"omitempty,amount,gt=0"`
	DueDate         string   `json:"due_date" binding:"omitempty,future_date"`
	Description     string   `json:"description" binding:"omitempty,max=1000"`
	Terms           string   `json:"terms" binding:"omitempty,max=2000"`
	Status          string   `json:"status" binding:"omitempty,oneof=draft sent paid overdue cancelled"`
	TaxAmount       *float64 `json:"tax_amount" binding:"omitempty,min=0"`
	DiscountAmount  *float64 `json:"discount_amount" binding:"omitempty,min=0"`
}

// Financing validation models
type FinancingRequestCreate struct {
	InvoiceID       string  `json:"invoice_id" binding:"required,uuid"`
	RequestedAmount float64 `json:"requested_amount" binding:"required,amount,gt=0"`
	InterestRate    float64 `json:"interest_rate" binding:"required,min=0.01,max=50"`
	RepaymentTerm   int     `json:"repayment_term" binding:"required,min=1,max=365"`
	Purpose         string  `json:"purpose" binding:"omitempty,max=500"`
}

type FinancingRequestUpdate struct {
	Status          string   `json:"status" binding:"omitempty,oneof=pending approved rejected funded completed"`
	InterestRate    *float64 `json:"interest_rate" binding:"omitempty,min=0.01,max=50"`
	RepaymentTerm   *int     `json:"repayment_term" binding:"omitempty,min=1,max=365"`
	ApprovalNotes   string   `json:"approval_notes" binding:"omitempty,max=1000"`
	RejectionReason string   `json:"rejection_reason" binding:"omitempty,max=1000"`
}

// Investment validation models
type InvestmentCreateRequest struct {
	FinancingRequestID string  `json:"financing_request_id" binding:"required,uuid"`
	Amount             float64 `json:"amount" binding:"required,amount,gt=0"`
}

type InvestmentUpdateRequest struct {
	Status       string   `json:"status" binding:"omitempty,oneof=pending active completed defaulted cancelled"`
	ActualReturn *float64 `json:"actual_return" binding:"omitempty,min=0"`
	Notes        string   `json:"notes" binding:"omitempty,max=1000"`
}

// File upload validation
type FileUploadRequest struct {
	InvoiceID   string `form:"invoice_id" binding:"required,uuid"`
	FileType    string `form:"file_type" binding:"required,oneof=pdf png jpg jpeg"`
	Description string `form:"description" binding:"omitempty,max=255"`
}

// Query parameter validation
type PaginationQuery struct {
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" binding:"omitempty,min=0"`
	Sort   string `form:"sort" binding:"omitempty"`
	Order  string `form:"order" binding:"omitempty,oneof=asc desc"`
}

type InvoiceFilterQuery struct {
	PaginationQuery
	Status       string  `form:"status" binding:"omitempty,oneof=draft sent paid overdue cancelled"`
	MinAmount    float64 `form:"min_amount" binding:"omitempty,min=0"`
	MaxAmount    float64 `form:"max_amount" binding:"omitempty,min=0"`
	CustomerName string  `form:"customer_name" binding:"omitempty,max=100"`
	DueDateFrom  string  `form:"due_date_from" binding:"omitempty"`
	DueDateTo    string  `form:"due_date_to" binding:"omitempty"`
}

type FinancingFilterQuery struct {
	PaginationQuery
	Status    string  `form:"status" binding:"omitempty,oneof=pending approved rejected funded completed"`
	RiskLevel string  `form:"risk_level" binding:"omitempty,risk_level"`
	MinAmount float64 `form:"min_amount" binding:"omitempty,min=0"`
	MaxAmount float64 `form:"max_amount" binding:"omitempty,min=0"`
}

type InvestmentFilterQuery struct {
	PaginationQuery
	Status       string  `form:"status" binding:"omitempty,oneof=pending active completed defaulted cancelled"`
	MinAmount    float64 `form:"min_amount" binding:"omitempty,min=0"`
	MaxAmount    float64 `form:"max_amount" binding:"omitempty,min=0"`
	MinReturn    float64 `form:"min_return" binding:"omitempty,min=0"`
	MaxReturn    float64 `form:"max_return" binding:"omitempty,min=0"`
	RiskLevel    string  `form:"risk_level" binding:"omitempty,risk_level"`
}

// Blockchain validation models
type ContractDeployRequest struct {
	InvoiceID    string `json:"invoice_id" binding:"required,uuid"`
	Amount       float64 `json:"amount" binding:"required,amount,gt=0"`
	InterestRate float64 `json:"interest_rate" binding:"required,min=0.01,max=50"`
	Duration     int    `json:"duration" binding:"required,min=1,max=365"`
}

type ContractInteractionRequest struct {
	ContractAddress string  `json:"contract_address" binding:"required,eth_addr"`
	Action          string  `json:"action" binding:"required,oneof=invest withdraw claim"`
	Amount          *float64 `json:"amount" binding:"omitempty,amount,gt=0"`
}

// AI Service validation models
type CreditScoreRequest struct {
	UserID          string  `json:"user_id" binding:"required,uuid"`
	AnnualRevenue   float64 `json:"annual_revenue" binding:"required,amount,gt=0"`
	MonthsInBusiness int    `json:"months_in_business" binding:"required,min=1,max=1200"`
	Industry        string  `json:"industry" binding:"required,min=2,max=100"`
	CreditHistory   int     `json:"credit_history" binding:"omitempty,min=0,max=100"`
}

type RiskAssessmentRequest struct {
	InvoiceID       string  `json:"invoice_id" binding:"required,uuid"`
	CustomerName    string  `json:"customer_name" binding:"required,min=2,max=100"`
	InvoiceAmount   float64 `json:"invoice_amount" binding:"required,amount,gt=0"`
	PaymentHistory  int     `json:"payment_history" binding:"omitempty,min=0,max=100"`
	IndustryRisk    string  `json:"industry_risk" binding:"omitempty,oneof=low medium high"`
}

type FraudDetectionRequest struct {
	InvoiceID     string  `json:"invoice_id" binding:"required,uuid"`
	CustomerEmail string  `json:"customer_email" binding:"required,email"`
	Amount        float64 `json:"amount" binding:"required,amount,gt=0"`
	IPAddress     string  `json:"ip_address" binding:"omitempty,ip"`
	UserAgent     string  `json:"user_agent" binding:"omitempty,max=500"`
}

type DocumentVerificationRequest struct {
	DocumentType string `json:"document_type" binding:"required,oneof=invoice contract identity bank_statement"`
	DocumentURL  string `json:"document_url" binding:"required,url"`
	UserID       string `json:"user_id" binding:"required,uuid"`
}

// Response validation models
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	APIResponse
	Pagination PaginationMeta `json:"pagination"`
}

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
}

// Webhook validation models
type WebhookPayload struct {
	Event     string                 `json:"event" binding:"required"`
	Data      map[string]interface{} `json:"data" binding:"required"`
	Timestamp int64                  `json:"timestamp" binding:"required"`
	Signature string                 `json:"signature" binding:"required"`
}

// Notification validation models
type NotificationCreateRequest struct {
	UserID  string `json:"user_id" binding:"required,uuid"`
	Type    string `json:"type" binding:"required,oneof=info warning error success"`
	Title   string `json:"title" binding:"required,min=1,max=255"`
	Message string `json:"message" binding:"required,min=1,max=1000"`
	Link    string `json:"link" binding:"omitempty,url"`
}

// Admin validation models
type AdminUserUpdateRequest struct {
	Role        string   `json:"role" binding:"omitempty,user_role"`
	IsActive    *bool    `json:"is_active" binding:"omitempty"`
	CreditScore *int     `json:"credit_score" binding:"omitempty,min=300,max=850"`
	Notes       string   `json:"notes" binding:"omitempty,max=2000"`
}

type SystemConfigRequest struct {
	Key   string      `json:"key" binding:"required,min=1,max=100"`
	Value interface{} `json:"value" binding:"required"`
}

// Search validation models
type SearchRequest struct {
	Query    string   `json:"query" binding:"required,min=1,max=255"`
	Type     string   `json:"type" binding:"omitempty,oneof=invoices users financing_requests investments"`
	Filters  map[string]interface{} `json:"filters" binding:"omitempty"`
	SortBy   string   `json:"sort_by" binding:"omitempty"`
	SortOrder string  `json:"sort_order" binding:"omitempty,oneof=asc desc"`
	Limit    int      `json:"limit" binding:"omitempty,min=1,max=100"`
	Offset   int      `json:"offset" binding:"omitempty,min=0"`
}

// Helper types for validation
type TimeRange struct {
	Start time.Time `json:"start" binding:"required"`
	End   time.Time `json:"end" binding:"required,gtfield=Start"`
}

type AmountRange struct {
	Min float64 `json:"min" binding:"omitempty,min=0"`
	Max float64 `json:"max" binding:"omitempty,min=0,gtfield=Min"`
}

// Custom validation tags documentation:
// - password: min 8 chars, must contain uppercase, lowercase, digit, special char
// - phone: valid phone number format
// - company_name: 2-100 chars, letters, numbers, spaces, common punctuation
// - invoice_number: 3-50 chars, letters, numbers, hyphens, underscores
// - future_date: date must be in the future
// - amount: positive number, max 10,000,000
// - risk_level: must be 'low', 'medium', or 'high'
// - user_role: must be 'sme', 'investor', or 'admin'
// - uuid: valid UUID format
// - eth_addr: valid Ethereum address format
