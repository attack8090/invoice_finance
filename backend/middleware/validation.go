package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("company_name", validateCompanyName)
	validate.RegisterValidation("invoice_number", validateInvoiceNumber)
	validate.RegisterValidation("future_date", validateFutureDate)
	validate.RegisterValidation("amount", validateAmount)
	validate.RegisterValidation("risk_level", validateRiskLevel)
	validate.RegisterValidation("user_role", validateUserRole)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details []ValidationError `json:"details"`
}

// ValidateJSON middleware for validating JSON input
func ValidateJSON(obj interface{}) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check Content-Type for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}

		// Bind and validate JSON
		if err := c.ShouldBindJSON(obj); err != nil {
			var validationErrors []ValidationError
			
			// Handle JSON syntax errors
			if jsonErr, ok := err.(*json.SyntaxError); ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Invalid JSON syntax at position %d", jsonErr.Offset),
				})
				c.Abort()
				return
			}

			// Handle validation errors
			if validationErr, ok := err.(validator.ValidationErrors); ok {
				for _, fieldErr := range validationErr {
					validationErrors = append(validationErrors, ValidationError{
						Field:   strings.ToLower(fieldErr.Field()),
						Message: getValidationMessage(fieldErr),
					})
				}
			} else {
				// Handle other binding errors
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid request format",
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusBadRequest, ValidationErrorResponse{
				Error:   "Validation failed",
				Details: validationErrors,
			})
			c.Abort()
			return
		}

		// Store validated object in context
		c.Set("validated_data", obj)
		c.Next()
	})
}

// ValidateQueryParams validates query parameters
func ValidateQueryParams(c *gin.Context) {
	// Sanitize and validate common query parameters
	if limit := c.Query("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err != nil || limitInt < 1 || limitInt > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid limit parameter (must be between 1 and 100)",
			})
			c.Abort()
			return
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if offsetInt, err := strconv.Atoi(offset); err != nil || offsetInt < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid offset parameter (must be non-negative)",
			})
			c.Abort()
			return
		}
	}

	// Validate sort parameter
	if sort := c.Query("sort"); sort != "" {
		allowedSortFields := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"amount":     true,
			"due_date":   true,
			"status":     true,
		}
		
		sortField := strings.TrimPrefix(sort, "-") // Remove DESC prefix
		if !allowedSortFields[sortField] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid sort field",
			})
			c.Abort()
			return
		}
	}

	c.Next()
}

// SanitizeInput middleware to sanitize string inputs
func SanitizeInput() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get the request body for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err == nil {
				sanitizedBody := sanitizeMap(body)
				
				// Convert back to JSON and replace request body
				jsonData, _ := json.Marshal(sanitizedBody)
				c.Request.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))
				c.Request.Body = &readCloser{strings.NewReader(string(jsonData))}
			}
		}

		c.Next()
	})
}

// Custom validator functions
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	
	// Check for at least one uppercase, lowercase, digit, and special character
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	phoneRegex := regexp.MustCompile(`^(\+\d{1,3}[- ]?)?\d{10}$`)
	return phoneRegex.MatchString(phone)
}

func validateCompanyName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	// Check for minimum length and no special characters except allowed ones
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-&.,()]+$`)
	return len(name) >= 2 && len(name) <= 100 && nameRegex.MatchString(name)
}

func validateInvoiceNumber(fl validator.FieldLevel) bool {
	invoiceNum := fl.Field().String()
	// Allow alphanumeric with hyphens and underscores
	invoiceRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	return len(invoiceNum) >= 3 && len(invoiceNum) <= 50 && invoiceRegex.MatchString(invoiceNum)
}

func validateFutureDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	if dateStr == "" {
		return false
	}
	
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	
	return date.After(time.Now())
}

func validateAmount(fl validator.FieldLevel) bool {
	amount := fl.Field().Float()
	return amount > 0 && amount <= 10000000 // Max 10 million
}

func validateRiskLevel(fl validator.FieldLevel) bool {
	riskLevel := fl.Field().String()
	allowedLevels := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}
	return allowedLevels[riskLevel]
}

func validateUserRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	allowedRoles := map[string]bool{
		"sme":      true,
		"investor": true,
		"admin":    true,
	}
	return allowedRoles[role]
}

// Helper functions
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("Must be no more than %s characters", fe.Param())
	case "password":
		return "Password must be at least 8 characters with uppercase, lowercase, digit and special character"
	case "phone":
		return "Invalid phone number format"
	case "company_name":
		return "Company name must be 2-100 characters with only letters, numbers, spaces and common punctuation"
	case "invoice_number":
		return "Invoice number must be 3-50 characters with only letters, numbers, hyphens and underscores"
	case "future_date":
		return "Date must be in the future"
	case "amount":
		return "Amount must be positive and not exceed 10,000,000"
	case "risk_level":
		return "Risk level must be 'low', 'medium', or 'high'"
	case "user_role":
		return "User role must be 'sme', 'investor', or 'admin'"
	default:
		return "Invalid value"
	}
}

func sanitizeMap(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	
	for key, value := range data {
		switch v := value.(type) {
		case string:
			sanitized[key] = sanitizeString(v)
		case map[string]interface{}:
			sanitized[key] = sanitizeMap(v)
		case []interface{}:
			sanitized[key] = sanitizeSlice(v)
		default:
			sanitized[key] = value
		}
	}
	
	return sanitized
}

func sanitizeSlice(data []interface{}) []interface{} {
	sanitized := make([]interface{}, len(data))
	
	for i, value := range data {
		switch v := value.(type) {
		case string:
			sanitized[i] = sanitizeString(v)
		case map[string]interface{}:
			sanitized[i] = sanitizeMap(v)
		case []interface{}:
			sanitized[i] = sanitizeSlice(v)
		default:
			sanitized[i] = value
		}
	}
	
	return sanitized
}

func sanitizeString(input string) string {
	// Remove potential XSS characters
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")
	input = strings.ReplaceAll(input, "&", "&amp;")
	
	// Remove potential SQL injection patterns (basic)
	sqlPatterns := []string{"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_", "DROP", "SELECT", "INSERT", "UPDATE", "DELETE"}
	for _, pattern := range sqlPatterns {
		input = strings.ReplaceAll(input, pattern, "")
	}
	
	// Trim whitespace
	return strings.TrimSpace(input)
}

// Helper types
type readCloser struct {
	*strings.Reader
}

func (r *readCloser) Close() error {
	return nil
}
