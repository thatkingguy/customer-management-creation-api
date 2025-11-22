package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/dto"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/repository"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

type RequestService interface {
	CreateDraftRequest(ctx context.Context, payload *dto.CreateDraftRequest, userDetails *UserDetails) (*dto.CreateDraftResponse, error)
	ApproveRequest(ctx context.Context, requestID uuid.UUID, payload *dto.ApproveRequestPayload, userDetails *UserDetails) (*dto.ApproveRequestResponse, error)
}

type requestService struct {
	db                        *sqlx.DB
	requestRepo               repository.RequestRepository
	customerRepo              repository.CustomerRepository
	profileRepo               repository.CustomerProfileRepository
	lookupRepo                repository.CustomerProfileLookupRepository
	activityLogRepo           repository.ActivityLogRepository
	sectorRepo                repository.SectorRepository
	industryRepo              repository.IndustryRepository
	targetRepo                repository.TargetRepository
	riskAssessmentRepo        repository.RiskAssessmentRepository
	signatoryRepo             repository.SignatoryRepository
	otherAccountRepo          repository.OtherAccountRepository
	referenceRepo             repository.ReferenceRepository
	interimApprovalConfigRepo repository.InterimApprovalConfigRepository
	gracePeriodRepo           repository.CustomerGracePeriodRepository
	waiverRequestRepo         repository.WaiverRequestRepository
	validationSvc             ValidationService
	idGen                     IDGeneratorService
}

// UserDetails represents user information (matching middleware)
type UserDetails struct {
	Name   string
	ID     uuid.UUID
	Branch string
}

func NewRequestService(
	db *sqlx.DB,
	requestRepo repository.RequestRepository,
	customerRepo repository.CustomerRepository,
	profileRepo repository.CustomerProfileRepository,
	lookupRepo repository.CustomerProfileLookupRepository,
	activityLogRepo repository.ActivityLogRepository,
	sectorRepo repository.SectorRepository,
	industryRepo repository.IndustryRepository,
	targetRepo repository.TargetRepository,
	riskAssessmentRepo repository.RiskAssessmentRepository,
	signatoryRepo repository.SignatoryRepository,
	otherAccountRepo repository.OtherAccountRepository,
	referenceRepo repository.ReferenceRepository,
	interimApprovalConfigRepo repository.InterimApprovalConfigRepository,
	gracePeriodRepo repository.CustomerGracePeriodRepository,
	waiverRequestRepo repository.WaiverRequestRepository,
	validationSvc ValidationService,
	idGen IDGeneratorService,
) RequestService {
	return &requestService{
		db:                        db,
		requestRepo:               requestRepo,
		customerRepo:              customerRepo,
		profileRepo:               profileRepo,
		lookupRepo:                lookupRepo,
		activityLogRepo:           activityLogRepo,
		sectorRepo:                sectorRepo,
		industryRepo:              industryRepo,
		targetRepo:                targetRepo,
		riskAssessmentRepo:        riskAssessmentRepo,
		signatoryRepo:             signatoryRepo,
		otherAccountRepo:          otherAccountRepo,
		referenceRepo:             referenceRepo,
		interimApprovalConfigRepo: interimApprovalConfigRepo,
		gracePeriodRepo:           gracePeriodRepo,
		waiverRequestRepo:         waiverRequestRepo,
		validationSvc:             validationSvc,
		idGen:                     idGen,
	}
}

func (s *requestService) CreateDraftRequest(ctx context.Context, payload *dto.CreateDraftRequest, userDetails *UserDetails) (*dto.CreateDraftResponse, error) {
	// Validate customer type
	if payload.CustomerType == "" {
		return nil, utils.NewValidationError("Customer type not specified, Individual or SME", nil)
	}

	customerType := mapCustomerType(payload.CustomerType)
	if customerType == nil {
		return nil, utils.NewValidationError("Invalid customer type", nil)
	}

	// Generate request ID
	requestID := s.idGen.GenerateUUID()

	// Determine request type
	requestType := mapRequestType(payload.Data.RequestData.RequestType)

	// Determine creation mode
	creationMode := mapCreationMode(payload.Data.FormInformation.FormType)

	// Build request title
	requestTitle := buildRequestTitle(payload, requestType)

	// Convert data to JSONB
	dataJSON := models.JSONB{
		"branch":             payload.Data.Branch,
		"customerData":       payload.Data.CustomerData,
		"formInformation":    payload.Data.FormInformation,
		"requestData":        payload.Data.RequestData,
		"waiverData":         payload.Data.WaiverData,
		"signatoryData":      payload.Data.SignatoryData,
		"executiveData":      payload.Data.ExecutiveData,
		"accountData":        payload.Data.AccountData,
		"riskAssessmentData": payload.Data.RiskAssessmentData,
		"riskData":           payload.Data.RiskData,
		"referenceData":      payload.Data.ReferenceData,
	}

	// Create request record
	request := &models.Request{
		RequestID:      requestID,
		RequestTitle:   requestTitle,
		RequestType:    requestType,
		CustomerType:   customerType,
		CreationMode:   &creationMode,
		Status:         models.RequestStatusDraft,
		ApprovalStatus: models.ApprovalStatusPending,
		Initiator:      userDetails.Name,
		InitiatorID:    userDetails.ID,
		Branch:         stringPtr(userDetails.Branch),
		Data:           dataJSON,
		Withdrawn:      false,
		IsDeleted:      false,
		IsProduct:      false,
	}

	// Create request in database
	err := s.requestRepo.Create(ctx, nil, request)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to create draft request")
		return nil, utils.NewDatabaseError("Failed to create draft request", err)
	}

	// Create activity log
	activityLog := &models.ActivityLog{
		ActivityLogID: s.idGen.GenerateUUID(),
		RequestID:     &requestID,
		Description:   fmt.Sprintf("Customer %s request saved to draft by %s", requestType, userDetails.Name),
		Reason:        nil,
	}

	err = s.activityLogRepo.Create(ctx, nil, activityLog)
	if err != nil {
		logger.Log.WithError(err).Warn("Failed to create activity log for draft request")
		// Don't fail the request if activity log fails
	}

	return &dto.CreateDraftResponse{
		RequestID: requestID,
	}, nil
}

func (s *requestService) ApproveRequest(ctx context.Context, requestID uuid.UUID, payload *dto.ApproveRequestPayload, userDetails *UserDetails) (*dto.ApproveRequestResponse, error) {
	startTime := time.Now()
	logger.Log.WithFields(map[string]interface{}{
		"request_id": requestID,
		"status":     payload.Status,
		"approver":   userDetails.Name,
	}).Debug("ApproveRequest: Starting approval process")

	// Fetch request
	logger.Log.WithField("request_id", requestID).Debug("ApproveRequest: Fetching request from database")
	fetchStartTime := time.Now()
	request, err := s.requestRepo.GetByID(ctx, s.db, requestID)
	fetchDuration := time.Since(fetchStartTime)
	if err != nil {
		logger.Log.WithError(err).Error("ApproveRequest: Failed to fetch request")
		return nil, utils.NewNotFoundError("Request not found", err)
	}
	logger.Log.WithFields(map[string]interface{}{
		"request_id":    requestID,
		"request_type":  request.RequestType,
		"customer_type": request.CustomerType,
		"status":        request.Status,
		"duration_ms":   fetchDuration.Milliseconds(),
	}).Debug("ApproveRequest: Request fetched successfully")

	// Validate request is in correct state
	if request.Status != models.RequestStatusDraft && request.Status != models.RequestStatusPending {
		logger.Log.WithFields(map[string]interface{}{
			"request_id": requestID,
			"status":     request.Status,
		}).Warn("ApproveRequest: Request is not in a state that can be approved")
		return nil, utils.NewValidationError("Request is not in a state that can be approved", nil)
	}

	// Normalize status
	status := strings.TrimSpace(payload.Status)
	if status == "" {
		logger.Log.Warn("ApproveRequest: Status is empty")
		return nil, utils.NewValidationError("Status is required", nil)
	}
	logger.Log.WithFields(map[string]interface{}{
		"request_id": requestID,
		"status":     status,
	}).Debug("ApproveRequest: Status normalized")

	// Process approval based on request type and status
	if request.RequestType == models.RequestTypeCreation {
		logger.Log.WithField("request_id", requestID).Debug("ApproveRequest: Processing Creation request type")
		if status == "Interim Approval" || status == "Approved" {
			logger.Log.WithFields(map[string]interface{}{
				"request_id":    requestID,
				"status":        status,
				"customer_type": request.CustomerType,
			}).Debug("ApproveRequest: Starting customer creation process")
			// Process customer creation
			var customerID uuid.UUID
			if request.CustomerType != nil && *request.CustomerType == models.CustomerTypeIndividual {
				logger.Log.WithField("request_id", requestID).Debug("ApproveRequest: Processing Individual customer")
				customerID, err = s.processIndividualCustomerOnApproval(ctx, request, status, userDetails)
			} else if request.CustomerType != nil && *request.CustomerType == models.CustomerTypeSME {
				logger.Log.WithField("request_id", requestID).Debug("ApproveRequest: Processing SME customer")
				customerID, err = s.processSMECustomerOnApproval(ctx, request, status, userDetails)
			} else {
				logger.Log.WithField("customer_type", request.CustomerType).Warn("ApproveRequest: Invalid customer type")
				return nil, utils.NewValidationError("Invalid customer type", nil)
			}

			if err != nil {
				logger.Log.WithError(err).WithFields(map[string]interface{}{
					"request_id":  requestID,
					"customer_id": customerID,
				}).Error("ApproveRequest: Customer creation failed")
				return nil, err
			}
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  requestID,
				"customer_id": customerID,
			}).Debug("ApproveRequest: Customer created successfully")

			// Update request status
			logger.Log.WithField("request_id", requestID).Debug("ApproveRequest: Updating request status")
			updateStatusStartTime := time.Now()
			approvalStatus := models.ApprovalStatusApproved
			if status == "Interim Approval" {
				approvalStatus = models.ApprovalStatusPending // Interim Approval maps to Pending approval status
			}

			requestStatus := models.RequestStatusApproved
			if status == "Interim Approval" {
				requestStatus = models.RequestStatusInterimApproval
			}

			err = s.requestRepo.UpdateStatus(ctx, nil, requestID, requestStatus, approvalStatus, &userDetails.Name, &userDetails.ID, stringPtr(userDetails.Branch))
			updateStatusDuration := time.Since(updateStatusStartTime)
			if err != nil {
				logger.Log.WithError(err).WithFields(map[string]interface{}{
					"request_id":  requestID,
					"duration_ms": updateStatusDuration.Milliseconds(),
				}).Error("ApproveRequest: Failed to update request status")
				return nil, utils.NewDatabaseError("Failed to update request status", err)
			}
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  requestID,
				"duration_ms": updateStatusDuration.Milliseconds(),
			}).Debug("ApproveRequest: Request status updated")

			totalDuration := time.Since(startTime)
			logger.Log.WithFields(map[string]interface{}{
				"request_id":        requestID,
				"customer_id":       customerID,
				"total_duration_ms": totalDuration.Milliseconds(),
			}).Info("ApproveRequest: Approval completed successfully")

			return &dto.ApproveRequestResponse{
				CustomerID:  customerID,
				RequestID:   requestID,
				SignatoryID: nil,
				Product:     nil,
			}, nil
		} else if status == "Rejected" {
			// Handle rejection
			approvalStatus := models.ApprovalStatusRejected
			requestStatus := models.RequestStatusRejected

			err = s.requestRepo.UpdateStatus(ctx, nil, requestID, requestStatus, approvalStatus, &userDetails.Name, &userDetails.ID, stringPtr(userDetails.Branch))
			if err != nil {
				logger.Log.WithError(err).Error("Failed to update request status")
				return nil, utils.NewDatabaseError("Failed to update request status", err)
			}

			return nil, utils.NewValidationError("Request rejected", nil)
		}
	}

	return nil, utils.NewValidationError("Invalid status or request type", nil)
}
