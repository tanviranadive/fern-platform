// Package middleware provides HTTP middleware components for the Fern Platform
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/guidewire-oss/fern-platform/pkg/config"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	config *config.AuthConfig
	logger *logging.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(cfg *config.AuthConfig, logger *logging.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config: cfg,
		logger: logger,
	}
}

// RequireAuth middleware validates JWT tokens
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		token := m.extractToken(c)
		if token == "" {
			m.logger.WithRequest(c.GetString("request_id"), c.Request.Method, c.Request.URL.Path).
				Warn("Missing authorization token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims, err := m.validateToken(token)
		if err != nil {
			m.logger.WithRequest(c.GetString("request_id"), c.Request.Method, c.Request.URL.Path).
				WithError(err).Error("Token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims["sub"])
		c.Set("user_claims", claims)

		entry := m.logger.WithRequest(c.GetString("request_id"), c.Request.Method, c.Request.URL.Path)
		logging.WithUser(entry, claims["sub"].(string)).Debug("User authenticated")

		c.Next()
	}
}

// OptionalAuth middleware validates JWT tokens but allows unauthenticated requests
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.validateToken(token)
		if err != nil {
			m.logger.WithRequest(c.GetString("request_id"), c.Request.Method, c.Request.URL.Path).
				WithError(err).Debug("Optional auth token validation failed")
			c.Next()
			return
		}

		// Set user context
		c.Set("user_id", claims["sub"])
		c.Set("user_claims", claims)

		c.Next()
	}
}

// extractToken extracts the JWT token from the Authorization header
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// validateToken validates a JWT token and returns the claims
func (m *AuthMiddleware) validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	// Validate issuer if configured
	if m.config.Issuer != "" {
		if iss, ok := claims["iss"]; !ok || iss != m.config.Issuer {
			return nil, fmt.Errorf("invalid issuer")
		}
	}

	// Validate audience if configured
	if m.config.Audience != "" {
		if aud, ok := claims["aud"]; !ok || aud != m.config.Audience {
			return nil, fmt.Errorf("invalid audience")
		}
	}

	return claims, nil
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// GetUserClaims extracts the user claims from the Gin context
func GetUserClaims(c *gin.Context) (jwt.MapClaims, bool) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(jwt.MapClaims)
	return userClaims, ok
}

// ContextKey type for context keys
type ContextKey string

const (
	UserIDKey     ContextKey = "user_id"
	UserClaimsKey ContextKey = "user_claims"
)

// SetUserContext sets user information in the context
func SetUserContext(ctx context.Context, userID string, claims jwt.MapClaims) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, UserClaimsKey, claims)
	return ctx
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetUserClaimsFromContext extracts user claims from context
func GetUserClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(jwt.MapClaims)
	return claims, ok
}
