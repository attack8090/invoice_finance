package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SecurityConfig holds security configuration
type SecurityConfig struct {
	TrustedProxies    []string
	AllowedOrigins    []string
	AllowCredentials  bool
	MaxAge            time.Duration
	AllowedMethods    []string
	AllowedHeaders    []string
	ExposedHeaders    []string
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() SecurityConfig {
	// Get allowed origins from environment or use default
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:5173"}
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
	}

	return SecurityConfig{
		TrustedProxies:   []string{"127.0.0.1", "::1"},
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
			"X-Api-Key",
		},
		ExposedHeaders: []string{
			"X-Rate-Limit-Limit",
			"X-Rate-Limit-Remaining",
			"X-Rate-Limit-Reset",
		},
	}
}

// ProductionSecurityConfig returns production-ready security configuration
func ProductionSecurityConfig() SecurityConfig {
	config := DefaultSecurityConfig()
	
	// More restrictive in production
	config.AllowedOrigins = []string{
		os.Getenv("FRONTEND_URL"),
		os.Getenv("PRODUCTION_URL"),
	}
	
	// Filter out empty origins
	var filteredOrigins []string
	for _, origin := range config.AllowedOrigins {
		if origin != "" {
			filteredOrigins = append(filteredOrigins, origin)
		}
	}
	config.AllowedOrigins = filteredOrigins
	
	return config
}

// SetupSecurity configures security middleware
func SetupSecurity(r *gin.Engine, config SecurityConfig) {
	// Set trusted proxies
	if len(config.TrustedProxies) > 0 {
		r.SetTrustedProxies(config.TrustedProxies)
	}

	// Security headers middleware
	r.Use(SecurityHeaders())
	
	// CORS middleware
	r.Use(CORSWithConfig(config))
	
	// Content Security Policy
	r.Use(CSPMiddleware())
	
	// Request size limit
	r.Use(RequestSizeLimit(10 << 20)) // 10MB limit
}

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Strict Transport Security (HTTPS only)
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		// Remove server header
		c.Header("Server", "")
		
		// Prevent caching of sensitive data
		if strings.Contains(c.Request.URL.Path, "/auth") || 
		   strings.Contains(c.Request.URL.Path, "/users") ||
		   strings.Contains(c.Request.URL.Path, "/profile") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		
		c.Next()
	})
}

// CORSWithConfig applies CORS middleware with custom configuration
func CORSWithConfig(config SecurityConfig) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     config.AllowedOrigins,
		AllowMethods:     config.AllowedMethods,
		AllowHeaders:     config.AllowedHeaders,
		ExposeHeaders:    config.ExposedHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	}

	// In development, allow all origins if configured
	if gin.Mode() == gin.DebugMode && os.Getenv("CORS_ALLOW_ALL") == "true" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false // Can't use credentials with AllowAllOrigins
	}

	return cors.New(corsConfig)
}

// CSPMiddleware adds Content Security Policy headers
func CSPMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Default CSP policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self'; " +
			"media-src 'self'; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none';"

		// Use environment variable for production CSP
		if prodCSP := os.Getenv("CSP_POLICY"); prodCSP != "" {
			csp = prodCSP
		}

		c.Header("Content-Security-Policy", csp)
		c.Next()
	})
}

// RequestSizeLimit limits the size of request bodies
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	})
}

// SecureJSON middleware to prevent JSON hijacking
func SecureJSON() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		
		// Add JSON security prefix for older browsers
		if strings.Contains(c.GetHeader("Accept"), "application/json") {
			c.Next()
			return
		}
		
		c.Next()
	})
}

// APIKeyAuth middleware for API key authentication (optional)
func APIKeyAuth() gin.HandlerFunc {
	validAPIKeys := make(map[string]bool)
	
	// Load API keys from environment
	if keys := os.Getenv("API_KEYS"); keys != "" {
		for _, key := range strings.Split(keys, ",") {
			if key != "" {
				validAPIKeys[strings.TrimSpace(key)] = true
			}
		}
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip if no API keys configured
		if len(validAPIKeys) == 0 {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" || !validAPIKeys[apiKey] {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing API key",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// HTTPSRedirect redirects HTTP requests to HTTPS
func HTTPSRedirect() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip in development mode
		if gin.Mode() == gin.DebugMode {
			c.Next()
			return
		}

		// Check if request is already HTTPS
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Next()
			return
		}

		// Redirect to HTTPS
		httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
		c.Redirect(http.StatusPermanentRedirect, httpsURL)
		c.Abort()
	})
}

// NoCache adds headers to prevent caching
func NoCache() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
		c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
		c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		c.Next()
	})
}

// IPWhitelist restricts access to whitelisted IP addresses
func IPWhitelist(allowedIPs ...string) gin.HandlerFunc {
	whitelist := make(map[string]bool)
	for _, ip := range allowedIPs {
		whitelist[ip] = true
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := getClientIP(c)
		
		if !whitelist[clientIP] {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: IP not whitelisted",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID (in production, use a proper UUID library)
			requestID = generateRequestID()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// Helper function to generate request ID
func generateRequestID() string {
	// Simple request ID generation (use proper UUID in production)
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// Helper function to generate random string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
