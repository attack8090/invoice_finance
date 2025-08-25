package api

import (
	"net/http"
	"strings"
	"time"

	"invoice-financing-platform/internal/models"
	"invoice-financing-platform/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email       string           `json:"email" binding:"required,email"`
	Password    string           `json:"password" binding:"required,min=8"`
	FirstName   string           `json:"first_name" binding:"required"`
	LastName    string           `json:"last_name" binding:"required"`
	Role        models.UserRole  `json:"role" binding:"required"`
	CompanyName string           `json:"company_name"`
	TaxID       string           `json:"tax_id"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         models.User `json:"user"`
	ExpiresIn    int64       `json:"expires_in"`
}

func (s *Server) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	existingUser, _ := s.userService.GetByEmail(req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         req.Role,
		CompanyName:  req.CompanyName,
		TaxID:        req.TaxID,
	}

	if err := s.userService.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate tokens
	token, refreshToken, err := s.generateTokens(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Remove password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         *user,
		ExpiresIn:    time.Now().Add(24 * time.Hour).Unix(),
	})
}

func (s *Server) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by email
	user, err := s.userService.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	token, refreshToken, err := s.generateTokens(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Remove password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusOK, AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         *user,
		ExpiresIn:    time.Now().Add(24 * time.Hour).Unix(),
	})
}

func (s *Server) refreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("Authorization")
	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token required"})
		return
	}

	refreshToken = strings.TrimPrefix(refreshToken, "Bearer ")
	
	claims, err := auth.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	userID, _ := uuid.Parse(claims["user_id"].(string))
	email := claims["email"].(string)
	role := claims["role"].(string)

	// Generate new tokens
	token, newRefreshToken, err := s.generateTokens(userID, email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         token,
		"refresh_token": newRefreshToken,
		"expires_in":    time.Now().Add(24 * time.Hour).Unix(),
	})
}

func (s *Server) logout(c *gin.Context) {
	// In a production system, you would invalidate the token
	// by adding it to a blacklist or removing it from a whitelist
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (s *Server) generateTokens(userID uuid.UUID, email, role string) (string, string, error) {
	// Generate access token (24 hours)
	token, err := auth.GenerateToken(userID.String(), email, role, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token (7 days)
	refreshToken, err := auth.GenerateToken(userID.String(), email, role, s.jwtSecret, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

// AuthMiddleware validates JWT tokens
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		
		claims, err := auth.ValidateToken(tokenString, s.jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims["user_id"])
		c.Set("user_email", claims["email"])
		c.Set("user_role", claims["role"])
		
		c.Next()
	}
}

// AdminMiddleware checks if user has admin role
func (s *Server) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != string(models.RoleAdmin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}
