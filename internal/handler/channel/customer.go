package channel

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/dto"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/handler/middleware"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/service"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

type CustomerHandler struct {
	customerService service.CustomerService
}

func NewCustomerHandler(customerService service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// DirectCreateCustomer handles POST /channel/directCreateCustomer
func (h *CustomerHandler) DirectCreateCustomer(c *gin.Context) {
	requestID, _ := c.Get(middleware.RequestIDKey)
	log := logger.Log.WithField("request_id", requestID)
	log.Debug("Received DirectCreateCustomer request")

	var req dto.DirectCreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind JSON request")
		errResp := utils.NewValidationError("Invalid request payload", err)
		c.JSON(http.StatusBadRequest, errResp.ToErrorResponse())
		return
	}
	log.Debug("Request payload parsed successfully")

	// Validate required fields
	if err := validateRequest(&req); err != nil {
		log.WithError(err).Error("Request validation failed")
		c.JSON(err.Code, err.ToErrorResponse())
		return
	}
	log.Debug("Request validation passed")

	// Extract bioData from customerData[0].data
	if len(req.Data.CustomerData) == 0 {
		log.Error("Customer data array is empty")
		err := utils.NewValidationError("Customer data is required", nil)
		c.JSON(http.StatusBadRequest, err.ToErrorResponse())
		return
	}

	bioData, ok := req.Data.CustomerData[0].Data.(map[string]interface{})
	if !ok {
		log.Error("Invalid customer data format - not a map")
		err := utils.NewValidationError("Invalid customer data format", nil)
		c.JSON(http.StatusBadRequest, err.ToErrorResponse())
		return
	}
	log.WithField("bioData_keys", len(bioData)).Debug("BioData extracted successfully")

	// Normalize customer type
	customerType := utils.NormalizeCustomerType(req.Data.CustomerType)
	log.WithField("customer_type", customerType).Debug("Customer type normalized")
	if customerType != "Individual" && customerType != "SME" {
		log.WithField("customer_type", customerType).Error("Invalid customer type")
		err := utils.NewValidationError("Customer type must be 'Individual' or 'SME'", nil)
		c.JSON(http.StatusBadRequest, err.ToErrorResponse())
		return
	}

	// Prepare service request
	serviceReq := &service.CreateCustomerRequest{
		Branch:        req.Data.Branch,
		CustomerType:  customerType,
		BioData:       bioData,
		TenantID:      req.Data.TenantID,
		Department:    req.Data.Department,
		RelatedEntity: req.Data.RelatedEntity,
	}

	if req.Data.FlagCustomer != nil {
		serviceReq.FlagCustomer = &service.FlagCustomer{
			Status: req.Data.FlagCustomer.Status,
		}
	}
	log.WithFields(map[string]interface{}{
		"branch":        serviceReq.Branch,
		"customer_type": serviceReq.CustomerType,
		"department":    serviceReq.Department,
	}).Debug("Service request prepared, calling CreateCustomer")

	// Create customer
	response, err := h.customerService.CreateCustomer(c.Request.Context(), serviceReq)
	if err != nil {
		log.WithError(err).Error("CreateCustomer service call failed")
		handleServiceError(c, err)
		return
	}

	log.WithFields(map[string]interface{}{
		"customer_id": response.CustomerID,
		"request_id":  response.RequestID,
	}).Info("Customer created successfully")

	// Return success response
	resp := dto.DirectCreateCustomerResponse{
		CustomerID: response.CustomerID,
		RequestID:  response.RequestID,
	}

	successResp := utils.NewSuccessResponse("Customer created successfully", resp, http.StatusOK)
	c.JSON(http.StatusCreated, successResp)
}

func validateRequest(req *dto.DirectCreateCustomerRequest) *utils.AppError {
	if req.Data.Branch == "" {
		return utils.NewValidationError("The payload is missing the 'branch' field", nil)
	}

	if req.Data.CustomerType == "" {
		return utils.NewValidationError("Customer type is required", nil)
	}

	if len(req.Data.CustomerData) == 0 {
		return utils.NewValidationError("Customer data is required", nil)
	}

	return nil
}

func handleServiceError(c *gin.Context, err error) {
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
	internalErr := utils.NewInternalError("Failed to create customer", err)
	c.JSON(http.StatusInternalServerError, internalErr.ToErrorResponse())
}
