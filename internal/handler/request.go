package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/dto"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/handler/middleware"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/service"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

type RequestHandler struct {
	requestService service.RequestService
}

func NewRequestHandler(requestService service.RequestService) *RequestHandler {
	return &RequestHandler{
		requestService: requestService,
	}
}

// CreateDraft handles POST /v1/draft?initiatorBranch={branch}
func (h *RequestHandler) CreateDraft(c *gin.Context) {
	requestID, _ := c.Get(middleware.RequestIDKey)
	log := logger.Log.WithField("request_id", requestID)
	log.Debug("Received CreateDraft request")

	// Get user details from context (set by JWT middleware)
	middlewareUserDetails, exists := middleware.GetUserDetails(c)
	if !exists {
		log.Error("User details not found in context")
		err := utils.NewUnauthorizedError("User details not found")
		c.JSON(http.StatusUnauthorized, err.ToErrorResponse())
		return
	}

	// Convert to service UserDetails
	userDetails := &service.UserDetails{
		Name:   middlewareUserDetails.Name,
		ID:     middlewareUserDetails.ID,
		Branch: middlewareUserDetails.Branch,
	}

	// Get branch from query param (override if provided)
	initiatorBranch := c.Query("initiatorBranch")
	if initiatorBranch != "" {
		userDetails.Branch = initiatorBranch
	}

	var req dto.CreateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind JSON request")
		errResp := utils.NewValidationError("Invalid request payload", err)
		c.JSON(http.StatusBadRequest, errResp.ToErrorResponse())
		return
	}
	log.Debug("Request payload parsed successfully")

	// Validate request
	if err := validateDraftRequest(&req); err != nil {
		log.WithError(err).Error("Request validation failed")
		c.JSON(err.Code, err.ToErrorResponse())
		return
	}
	log.Debug("Request validation passed")

	// Create draft request
	response, err := h.requestService.CreateDraftRequest(c.Request.Context(), &req, userDetails)
	if err != nil {
		log.WithError(err).Error("CreateDraftRequest service call failed")
		handleRequestServiceError(c, err)
		return
	}

	log.WithField("request_id", response.RequestID).Info("Draft request created successfully")

	// Return success response
	successResp := utils.NewSuccessResponse("Creation request saved to draft", response, http.StatusOK)
	c.JSON(http.StatusOK, successResp)
}

// ApproveRequest handles PATCH /v1/request/{requestId}?approverBranch={branch}
func (h *RequestHandler) ApproveRequest(c *gin.Context) {
	requestID, _ := c.Get(middleware.RequestIDKey)
	log := logger.Log.WithField("request_id", requestID)
	log.Debug("Received ApproveRequest request")

	// Get request ID from path
	requestIDStr := c.Param("requestId")
	requestUUID, err := uuid.Parse(requestIDStr)
	if err != nil {
		log.WithError(err).Error("Invalid request ID format")
		errResp := utils.NewValidationError("Invalid request ID", err)
		c.JSON(http.StatusBadRequest, errResp.ToErrorResponse())
		return
	}

	// Get user details from context (set by JWT middleware)
	middlewareUserDetails, exists := middleware.GetUserDetails(c)
	if !exists {
		log.Error("User details not found in context")
		err := utils.NewUnauthorizedError("User details not found")
		c.JSON(http.StatusUnauthorized, err.ToErrorResponse())
		return
	}

	// Convert to service UserDetails
	userDetails := &service.UserDetails{
		Name:   middlewareUserDetails.Name,
		ID:     middlewareUserDetails.ID,
		Branch: middlewareUserDetails.Branch,
	}

	// Get branch from query param (override if provided)
	approverBranch := c.Query("approverBranch")
	if approverBranch != "" {
		userDetails.Branch = approverBranch
	}

	var req dto.ApproveRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind JSON request")
		errResp := utils.NewValidationError("Invalid request payload", err)
		c.JSON(http.StatusBadRequest, errResp.ToErrorResponse())
		return
	}
	log.Debug("Request payload parsed successfully")

	// Validate status
	if req.Status != "Approved" && req.Status != "Interim Approval" && req.Status != "Rejected" {
		log.WithField("status", req.Status).Error("Invalid status")
		errResp := utils.NewValidationError("Status must be 'Approved', 'Interim Approval', or 'Rejected'", nil)
		c.JSON(http.StatusBadRequest, errResp.ToErrorResponse())
		return
	}

	// Approve request
	response, err := h.requestService.ApproveRequest(c.Request.Context(), requestUUID, &req, userDetails)
	if err != nil {
		log.WithError(err).Error("ApproveRequest service call failed")
		handleRequestServiceError(c, err)
		return
	}

	// Determine success message based on customer type
	message := "Customer created successfully"
	if response.CustomerID != uuid.Nil {
		log.WithFields(map[string]interface{}{
			"customer_id": response.CustomerID,
			"request_id":  response.RequestID,
		}).Info("Request approved and customer created successfully")
	}

	// Return success response
	successResp := utils.NewSuccessResponse(message, response, http.StatusOK)
	c.JSON(http.StatusOK, successResp)
}

func validateDraftRequest(req *dto.CreateDraftRequest) *utils.AppError {
	if req.CustomerType == "" {
		return utils.NewValidationError("Customer type not specified, Individual or SME", nil)
	}

	if req.CustomerType != "individual" && req.CustomerType != "sme" && req.CustomerType != "Individual" && req.CustomerType != "SME" {
		return utils.NewValidationError("Customer type must be 'Individual' or 'SME'", nil)
	}

	if req.Data.RequestData.RequestType == "" {
		return utils.NewValidationError("Request type not specified", nil)
	}

	return nil
}

func handleRequestServiceError(c *gin.Context, err error) {
	requestID, _ := c.Get(middleware.RequestIDKey)
	log := logger.Log.WithField("request_id", requestID)

	if appErr, ok := err.(*utils.AppError); ok {
		log.WithError(appErr.Err).WithFields(map[string]interface{}{
			"error_code":    appErr.Code,
			"error_message": appErr.Message,
		}).Error("Application error occurred")
		c.JSON(appErr.Code, appErr.ToErrorResponse())
		return
	}

	// Handle other error types
	log.WithError(err).Error("Unexpected error occurred")
	internalErr := utils.NewInternalError("Failed to process request", err)
	c.JSON(http.StatusInternalServerError, internalErr.ToErrorResponse())
}

