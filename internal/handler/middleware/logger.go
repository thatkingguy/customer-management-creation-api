package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
)

// LoggerMiddleware logs incoming requests and responses
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		status := c.Writer.Status()
		requestID, _ := c.Get(RequestIDKey)

		entry := logger.Log.WithFields(map[string]interface{}{
			"method":     method,
			"path":       path,
			"status":     status,
			"latency":    latency.String(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})

		if requestID != nil {
			entry = entry.WithField("request_id", requestID)
		}

		if status >= 500 {
			entry.Error("Request completed with error")
		} else if status >= 400 {
			entry.Warn("Request completed with client error")
		} else {
			entry.Info("Request completed successfully")
		}
	}
}

