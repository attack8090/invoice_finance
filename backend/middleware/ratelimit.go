package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter represents a rate limiter for a specific client
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	CleanupInterval   time.Duration
}

// RateLimitManager manages rate limiters for different clients
type RateLimitManager struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	config   RateLimitConfig
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(config RateLimitConfig) *RateLimitManager {
	manager := &RateLimitManager{
		limiters: make(map[string]*RateLimiter),
		config:   config,
	}

	// Start cleanup goroutine
	go manager.cleanupRoutine()

	return manager
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60,  // 60 requests per minute
		BurstSize:         10,  // Allow burst of 10 requests
		CleanupInterval:   time.Minute * 5, // Cleanup every 5 minutes
	}
}

// StrictRateLimitConfig returns strict rate limiting configuration for sensitive endpoints
func StrictRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 10,  // 10 requests per minute
		BurstSize:         3,   // Allow burst of 3 requests
		CleanupInterval:   time.Minute * 5,
	}
}

// AuthRateLimitConfig returns rate limiting configuration for auth endpoints
func AuthRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 5,   // 5 attempts per minute
		BurstSize:         2,   // Allow burst of 2 attempts
		CleanupInterval:   time.Minute * 10,
	}
}

// FileUploadRateLimitConfig returns rate limiting configuration for file uploads
func FileUploadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 20,  // 20 uploads per minute
		BurstSize:         5,   // Allow burst of 5 uploads
		CleanupInterval:   time.Minute * 5,
	}
}

// RateLimit middleware with default configuration
func RateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig())
}

// RateLimitWithConfig creates a rate limiting middleware with custom configuration
func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	manager := NewRateLimitManager(config)

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := getClientIP(c)
		
		if !manager.Allow(clientIP) {
			c.Header("X-Rate-Limit-Limit", strconv.Itoa(config.RequestsPerMinute))
			c.Header("X-Rate-Limit-Remaining", "0")
			c.Header("X-Rate-Limit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Maximum %d requests per minute allowed", config.RequestsPerMinute),
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := manager.GetRemaining(clientIP)
		c.Header("X-Rate-Limit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Header("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
		c.Header("X-Rate-Limit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))

		c.Next()
	})
}

// AuthRateLimit applies strict rate limiting for authentication endpoints
func AuthRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(AuthRateLimitConfig())
}

// StrictRateLimit applies strict rate limiting for sensitive endpoints
func StrictRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(StrictRateLimitConfig())
}

// FileUploadRateLimit applies rate limiting for file upload endpoints
func FileUploadRateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(FileUploadRateLimitConfig())
}

// IPBasedRateLimit applies rate limiting based on IP address with additional security
func IPBasedRateLimit() gin.HandlerFunc {
	manager := NewRateLimitManager(DefaultRateLimitConfig())
	suspiciousIPs := make(map[string]time.Time)
	var suspiciousMu sync.RWMutex

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := getClientIP(c)

		// Check if IP is marked as suspicious
		suspiciousMu.RLock()
		suspiciousTime, isSuspicious := suspiciousIPs[clientIP]
		suspiciousMu.RUnlock()

		if isSuspicious && time.Since(suspiciousTime) < time.Hour {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "IP temporarily blocked",
				"message": "Your IP has been temporarily blocked due to suspicious activity",
				"retry_after": 3600,
			})
			c.Abort()
			return
		}

		if !manager.Allow(clientIP) {
			// Mark IP as suspicious after multiple violations
			suspiciousMu.Lock()
			suspiciousIPs[clientIP] = time.Now()
			suspiciousMu.Unlock()

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. IP temporarily blocked.",
				"retry_after": 3600,
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// Allow checks if a request from the given client should be allowed
func (m *RateLimitManager) Allow(clientID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.limiters[clientID]
	if !exists {
		// Create new limiter for this client
		limiter = &RateLimiter{
			limiter:  rate.NewLimiter(rate.Every(time.Minute/time.Duration(m.config.RequestsPerMinute)), m.config.BurstSize),
			lastSeen: time.Now(),
		}
		m.limiters[clientID] = limiter
	}

	// Update last seen time
	limiter.lastSeen = time.Now()

	// Check if request is allowed
	return limiter.limiter.Allow()
}

// GetRemaining returns the number of remaining requests for a client
func (m *RateLimitManager) GetRemaining(clientID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limiter, exists := m.limiters[clientID]
	if !exists {
		return m.config.RequestsPerMinute
	}

	// Calculate remaining tokens (approximate)
	tokens := limiter.limiter.Tokens()
	remaining := int(tokens)
	if remaining < 0 {
		remaining = 0
	}
	if remaining > m.config.RequestsPerMinute {
		remaining = m.config.RequestsPerMinute
	}

	return remaining
}

// cleanupRoutine periodically removes old limiters
func (m *RateLimitManager) cleanupRoutine() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanup()
	}
}

// cleanup removes limiters that haven't been used recently
func (m *RateLimitManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-m.config.CleanupInterval * 2)
	
	for clientID, limiter := range m.limiters {
		if limiter.lastSeen.Before(cutoff) {
			delete(m.limiters, clientID)
		}
	}
}

// getClientIP extracts the real client IP address
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}

	// Check X-Real-IP header (for nginx)
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Check CF-Connecting-IP header (for Cloudflare)
	cfIP := c.GetHeader("CF-Connecting-IP")
	if cfIP != "" {
		return strings.TrimSpace(cfIP)
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}

// PerUserRateLimit applies rate limiting per authenticated user
func PerUserRateLimit() gin.HandlerFunc {
	manager := NewRateLimitManager(DefaultRateLimitConfig())

	return gin.HandlerFunc(func(c *gin.Context) {
		// Try to get user ID from context (set by auth middleware)
		var userID string
		if uid, exists := c.Get("user_id"); exists {
			userID = fmt.Sprintf("user_%v", uid)
		} else {
			// Fallback to IP-based limiting
			userID = fmt.Sprintf("ip_%s", getClientIP(c))
		}

		if !manager.Allow(userID) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests from your account",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	})
}
