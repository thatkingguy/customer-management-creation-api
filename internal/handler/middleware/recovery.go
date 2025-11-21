package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

// RecoveryMiddleware handles panics and returns 500 error
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get(RequestIDKey)
		log := logger.Log.WithField("request_id", requestID)
		
		panicErr := fmt.Errorf("panic recovered: %v", recovered)
		log.WithError(panicErr).
			WithField("method", c.Request.Method).
			WithField("path", c.Request.URL.Path).
			WithField("panic_value", fmt.Sprintf("%+v", recovered)).
			Error("Panic recovered - detailed error")
		
		// Log stack trace if available
		if err, ok := recovered.(error); ok {
			log.WithError(err).Error("Panic error details")
		}
		
		err := utils.NewInternalError("Internal server error", panicErr)
		c.JSON(http.StatusInternalServerError, err.ToErrorResponse())
		c.Abort()
	})
}

