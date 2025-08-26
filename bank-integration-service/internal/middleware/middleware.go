package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"bank-integration-service/internal/config"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})
}

// RequestLogging logs request details
func RequestLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RateLimit implements rate limiting
func RateLimit(cfg *config.Config) gin.HandlerFunc {
	// Simple in-memory rate limiter (in production, use Redis)
	clients := make(map[string][]time.Time)

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Clean old entries
		if requests, exists := clients[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < cfg.RateLimitWindow {
					validRequests = append(validRequests, reqTime)
				}
			}
			clients[clientIP] = validRequests
		}

		// Check rate limit
		if len(clients[clientIP]) >= cfg.RateLimitRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Maximum %d requests per %v allowed", cfg.RateLimitRequests, cfg.RateLimitWindow),
			})
			c.Abort()
			return
		}

		// Add current request
		clients[clientIP] = append(clients[clientIP], now)
		c.Next()
	})
}

// JWTAuth validates JWT tokens
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token required",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("userID", claims["user_id"])
			c.Set("userRole", claims["role"])
			c.Set("bankID", claims["bank_id"])
		}

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Role information not found",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid role format",
			})
			c.Abort()
			return
		}

		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient privileges",
			"required_roles": requiredRoles,
			"user_role": role,
		})
		c.Abort()
	}
}

// BankAccess ensures user can only access their bank's data
func BankAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userBankID, exists := c.Get("bankID")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Bank access information not found",
			})
			c.Abort()
			return
		}

		// Extract bank ID from URL parameters if present
		bankIDParam := c.Param("bankId")
		connectionIDParam := c.Param("connectionId")

		// If bank ID is specified in URL, ensure it matches user's bank
		if bankIDParam != "" && bankIDParam != userBankID {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied to bank data",
			})
			c.Abort()
			return
		}

		// Store bank access info for handlers
		c.Set("allowedBankID", userBankID)
		c.Set("connectionFilter", connectionIDParam)
		c.Next()
	}
}

// APIKeyAuth validates API keys for external integrations
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
			})
			c.Abort()
			return
		}

		// In production, validate against database
		// For now, simple validation
		if !isValidAPIKey(apiKey) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Set("apiKeyAuth", true)
		c.Next()
	}
}

// Epic4Compliance adds Epic 4 compliance headers and validation
func Epic4Compliance() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add Epic 4 compliance headers
		c.Header("X-Epic4-Compliant", "true")
		c.Header("X-Audit-Trail", "enabled")
		c.Header("X-Data-Retention", "7-years")

		// Set compliance context
		c.Set("epic4Compliant", true)
		c.Set("auditRequired", true)
		c.Set("encryptionRequired", true)

		c.Next()
	}
}

// TransactionLimits enforces transaction limits for compliance
func TransactionLimits(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this is a transaction endpoint
		if !isTransactionEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract amount from request if present
		var requestData struct {
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		}

		if err := c.ShouldBindJSON(&requestData); err == nil {
			// Validate transaction limits
			if requestData.Amount > cfg.MaxDailyTransactionAmount {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Transaction amount exceeds daily limit",
					"limit": cfg.MaxDailyTransactionAmount,
					"amount": requestData.Amount,
				})
				c.Abort()
				return
			}

			// Check for suspicious activity threshold
			if requestData.Amount > cfg.SuspiciousActivityThreshold {
				c.Set("suspiciousActivity", true)
				c.Set("requiresManualReview", true)
			}
		}

		c.Next()
	}
}

// ErrorHandler handles panics and errors gracefully
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"message": "An unexpected error occurred",
				"request_id": c.GetString("RequestID"),
			})
			
			// Log error for debugging (in production, use proper logging)
			fmt.Printf("Panic recovered: %s\n", err)
		}
		c.Abort()
	})
}

// Helper functions

func isValidAPIKey(apiKey string) bool {
	// In production, validate against database or external service
	validKeys := []string{
		"test-api-key-1",
		"test-api-key-2",
		"bank-integration-key",
	}

	for _, validKey := range validKeys {
		if apiKey == validKey {
			return true
		}
	}
	return false
}

func isTransactionEndpoint(path string) bool {
	transactionPaths := []string{
		"/payments/process",
		"/payments/bulk-process",
		"/transfers/initiate",
		"/transfers/bulk-transfer",
		"/financing/disburse",
	}

	for _, transactionPath := range transactionPaths {
		if strings.Contains(path, transactionPath) {
			return true
		}
	}
	return false
}

// Audit middleware for Epic 4 compliance
func AuditLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request details
		auditData := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"timestamp":  start,
		}

		if userID, exists := c.Get("userID"); exists {
			auditData["user_id"] = userID
		}

		if requestID, exists := c.Get("RequestID"); exists {
			auditData["request_id"] = requestID
		}

		c.Next()

		// Log after request completion
		auditData["status_code"] = c.Writer.Status()
		auditData["response_time"] = time.Since(start).Milliseconds()
		auditData["response_size"] = c.Writer.Size()

		// In production, send to audit service or log to database
		fmt.Printf("Audit Log: %+v\n", auditData)
	}
}

// Maintenance mode middleware
func MaintenanceMode(enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enabled && c.Request.URL.Path != "/health" {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service temporarily unavailable",
				"message": "Bank Integration Service is currently under maintenance",
				"status":  "maintenance_mode",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IP Whitelist middleware for sensitive operations
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		for _, allowedIP := range allowedIPs {
			if clientIP == allowedIP {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "IP address not allowed",
			"message": "Access denied from this IP address",
		})
		c.Abort()
	}
}

// Request size limiter
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request payload too large",
				"max_size": maxSize,
				"received": c.Request.ContentLength,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
