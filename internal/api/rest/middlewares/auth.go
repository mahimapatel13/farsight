// internal/api/rest/middlewares/auth.go
package middlewares

import (
	"slices"
	"strings"

	"budget-planner/internal/common/errors"
	"budget-planner/internal/infrastructure/auth"
	"budget-planner/pkg/logger"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtProvider   *auth.JWTProvider
	apiKeyManager *auth.APIKeyManager
	logger        *logger.Logger
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(
	jwtProvider *auth.JWTProvider,
	apiKeyManager *auth.APIKeyManager,
	logger *logger.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtProvider:   jwtProvider,
		apiKeyManager: apiKeyManager,
		logger:        logger,
	}
}

// ===================================
// ✅ JWT Authentication Middleware
// ===================================
func (m *AuthMiddleware) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			m.handleUnauthorized(c, errors.Unauthorized("missing or invalid JWT token"))
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := m.validateJWT(tokenString, false) // Not a refresh token
		if err != nil {
			m.handleUnauthorized(c, errors.Unauthorized("invalid or expired token"))
			return
		}

		// Store claims in context
		c.Set("userID", claims.UserID)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

// validateJWT parses and validates JWT token (access or refresh)
func (m *AuthMiddleware) validateJWT(tokenString string, isRefresh bool) (*auth.CustomClaims, error) {
	return m.jwtProvider.ValidateToken(tokenString, isRefresh)
}

// ===================================
// 🔐 API Key Authentication Middleware
// ===================================
func (m *AuthMiddleware) APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "ApiKey ") {
			m.handleUnauthorized(c, errors.Unauthorized("missing or invalid API key"))
			return
		}

		apiKey := strings.TrimPrefix(authHeader, "ApiKey ")
		keyInfo, err := m.apiKeyManager.ValidateKey(c.Request.Context(), apiKey)
		if err != nil {
			m.handleUnauthorized(c, errors.Unauthorized("invalid API key"))
			return
		}

		// Store API key info in context
		c.Set("clientID", keyInfo.ClientID)
		c.Set("keyScopes", keyInfo.Scopes)
		c.Next()
	}
}

// ===================================
// 🎟️ Require Roles Middleware
// ===================================
func (m *AuthMiddleware) RequireRoles(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRolesValue, exists := c.Get("roles")

		if !exists {
			m.handleForbidden(c, errors.Forbidden("user not authenticated or roles missing"))
			return
		}

		userRoles, ok := userRolesValue.([]string)
		if !ok {
			m.handleForbidden(c, errors.Forbidden("invalid roles data"))
			return
		}

		if !m.hasAnyRequiredRole(userRoles, requiredRoles) {
			m.handleForbidden(c, errors.Forbidden("insufficient permissions"))
			return
		}

		c.Next()
	}
}

// ===================================
// 🔒 Require API Key Scopes Middleware
// ===================================
func (m *AuthMiddleware) RequireScopes(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyScopesValue, exists := c.Get("keyScopes")
		if !exists {
			m.handleForbidden(c, errors.Forbidden("no API key scopes found"))
			return
		}

		keyScopes, ok := keyScopesValue.([]string)
		if !ok {
			m.handleForbidden(c, errors.Forbidden("invalid key scopes data"))
			return
		}

		if !m.hasRequiredScope(keyScopes, requiredScopes) {
			m.handleForbidden(c, errors.Forbidden("insufficient API key permissions"))
			return
		}

		c.Next()
	}
}

// Check if API key has required scopes
func (m *AuthMiddleware) hasRequiredScope(keyScopes, requiredScopes []string) bool {
	for _, required := range requiredScopes {
		if slices.Contains(keyScopes, required) {
			return true
		}
	}
	return false
}

// ===================================
// ⚠️ Error Handling Helpers
// ===================================
func (m *AuthMiddleware) handleUnauthorized(c *gin.Context, apiErr *errors.APIError) {
	m.logger.WithError(apiErr).Warn("Unauthorized access attempt")
	apiErr.RespondWithError(c)
	c.Abort()
}

func (m *AuthMiddleware) handleForbidden(c *gin.Context, apiErr *errors.APIError) {
	m.logger.WithError(apiErr).Warn("Forbidden access attempt")
	apiErr.RespondWithError(c)
	c.Abort()
}

func (m *AuthMiddleware) hasAnyRequiredRole(userRoles, requiredRoles []string) bool {
	for _, userRole := range userRoles {
		if slices.Contains(requiredRoles, userRole) {
			return true
		}
	}
	return false
}

