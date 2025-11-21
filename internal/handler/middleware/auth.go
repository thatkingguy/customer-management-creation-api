package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

// APIKeyAuthMiddleware validates API key from header
func APIKeyAuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check X-API-Key or x-api-key header (case-insensitive)
		providedKey := c.GetHeader("X-API-Key")
		if providedKey == "" {
			providedKey = c.GetHeader("x-api-key")
		}

		// Try to find in any case variation
		if providedKey == "" {
			for key, values := range c.Request.Header {
				if strings.EqualFold(key, "X-API-Key") {
					if len(values) > 0 {
						providedKey = values[0]
						break
					}
				}
			}
		}

		if providedKey == "" || providedKey != apiKey {
			err := utils.NewUnauthorizedError("Invalid or missing API key")
			c.JSON(http.StatusUnauthorized, err.ToErrorResponse())
			c.Abort()
			return
		}

		c.Next()
	}
}
