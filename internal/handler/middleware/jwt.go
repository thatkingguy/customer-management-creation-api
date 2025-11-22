package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

const (
	UserDetailsKey = "user_details"
)

// UserDetails represents user information extracted from JWT token
type UserDetails struct {
	Name   string
	ID     uuid.UUID
	Branch string
}

// JWTAuthMiddleware extracts user details from JWT token
// This middleware expects the token in the Authorization header as "Bearer {token}"
// The token should contain claims: name (or username), id (or userId), and branch
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			err := utils.NewUnauthorizedError("Missing authorization header")
			c.JSON(http.StatusUnauthorized, err.ToErrorResponse())
			c.Abort()
			return
		}

		// Extract token from "Bearer {token}" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			err := utils.NewUnauthorizedError("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, err.ToErrorResponse())
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			err := utils.NewUnauthorizedError("Missing token")
			c.JSON(http.StatusUnauthorized, err.ToErrorResponse())
			c.Abort()
			return
		}

		// Extract user details from token
		// Note: In a real implementation, you would parse and validate the JWT token
		// For now, we'll extract from query params as fallback (for testing)
		// In production, parse JWT claims properly
		userDetails, err := extractUserDetailsFromToken(token, c)
		if err != nil {
			err := utils.NewUnauthorizedError("Invalid or expired token")
			c.JSON(http.StatusUnauthorized, err.ToErrorResponse())
			c.Abort()
			return
		}

		// Store user details in context
		c.Set(UserDetailsKey, userDetails)
		c.Next()
	}
}

// extractUserDetailsFromToken extracts user details from JWT token
// SECURITY WARNING: This is a placeholder implementation for development/testing only.
// In production, this MUST be replaced with proper JWT token parsing and validation.
// The current implementation ignores the token and uses query parameters, which is a security vulnerability.
func extractUserDetailsFromToken(token string, c *gin.Context) (*UserDetails, error) {
	// SECURITY: The token parameter is currently ignored - this is a critical security vulnerability
	// TODO: Implement proper JWT parsing using github.com/golang-jwt/jwt/v5 or similar library
	// The token should be parsed, validated (signature, expiration, issuer), and claims extracted

	// Check environment - fail in production if JWT parsing is not implemented
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("NODE_ENV")
	}
	envLower := strings.ToLower(env)
	isProduction := envLower == "production" || envLower == "prod"

	if isProduction {
		// In production, JWT token MUST be parsed and validated
		// This fallback implementation is not secure and should not be used
		return nil, utils.NewUnauthorizedError("JWT token parsing not implemented - authentication disabled in production")
	}

	// For development/testing: fallback to query params
	// WARNING: This allows authentication bypass - only use in development
	name := c.Query("initiatorName")
	if name == "" {
		name = c.Query("approverName")
	}
	if name == "" {
		return nil, utils.NewUnauthorizedError("Missing user identification - provide initiatorName or approverName query param")
	}

	idStr := c.Query("initiatorId")
	if idStr == "" {
		idStr = c.Query("approverId")
	}
	var id uuid.UUID
	var err error
	if idStr != "" {
		id, err = uuid.Parse(idStr)
		if err != nil {
			return nil, utils.NewUnauthorizedError("Invalid user ID format")
		}
	} else {
		return nil, utils.NewUnauthorizedError("Missing user ID - provide initiatorId or approverId query param")
	}

	branch := c.Query("initiatorBranch")
	if branch == "" {
		branch = c.Query("approverBranch")
	}
	if branch == "" {
		return nil, utils.NewUnauthorizedError("Missing branch information - provide initiatorBranch or approverBranch query param")
	}

	// Explicitly acknowledge token is unused (security warning)
	// TODO: Remove this fallback and implement proper JWT parsing
	_ = token

	return &UserDetails{
		Name:   name,
		ID:     id,
		Branch: branch,
	}, nil
}

// GetUserDetails extracts user details from gin context
func GetUserDetails(c *gin.Context) (*UserDetails, bool) {
	userDetails, exists := c.Get(UserDetailsKey)
	if !exists {
		return nil, false
	}

	details, ok := userDetails.(*UserDetails)
	return details, ok
}
