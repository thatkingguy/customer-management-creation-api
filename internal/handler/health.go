package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/database"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/dto"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

// HealthHandler handles health check endpoints
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	resp := dto.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  "unknown",
	}

	c.JSON(http.StatusOK, resp)
}

// Ready handles GET /ready
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dbStatus := "connected"
	if err := database.HealthCheck(ctx); err != nil {
		dbStatus = "disconnected"
		errResp := utils.NewInternalError("Database connection failed", err)
		c.JSON(http.StatusServiceUnavailable, errResp.ToErrorResponse())
		return
	}

	resp := dto.HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  dbStatus,
	}

	c.JSON(http.StatusOK, resp)
}
