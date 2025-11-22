package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/dto"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
	"github.com/sterling-retailcore-team/customer-management-creation-api/pkg/postgres"
)

func (s *requestService) processIndividualCustomerOnApproval(ctx context.Context, request *models.Request, status string, userDetails *UserDetails) (uuid.UUID, error) {
	startTime := time.Now()
	logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Starting")
	var customerID uuid.UUID

	// Check interim approval config BEFORE starting transaction
	// This is a read-only operation and shouldn't block the transaction
	// Only check if status is "Interim Approval" to avoid unnecessary query
	logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Checking interim approval config (before transaction)")
	configStartTime := time.Now()
	var interimConfig *models.InterimApprovalConfig
	if request.CustomerType != nil && status == "Interim Approval" {
		// Use a separate context with short timeout for this read operation
		configCtx, configCancel := context.WithTimeout(context.Background(), 2*time.Second)
		config, err := s.interimApprovalConfigRepo.GetByCustomerType(configCtx, s.db, *request.CustomerType)
		configCancel()
		configDuration := time.Since(configStartTime)
		if err == nil {
			interimConfig = config
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": configDuration.Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: Interim approval config found")
		} else {
			// Log error but don't fail - config is optional
			logger.Log.WithError(err).WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": configDuration.Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: No interim approval config found (continuing without it)")
		}
	} else {
		// Skip query if not Interim Approval
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Skipping interim config check (not Interim Approval)")
	}

	hasInterimConfig := interimConfig != nil
	logger.Log.WithFields(map[string]interface{}{
		"request_id":         request.RequestID,
		"has_interim_config": hasInterimConfig,
	}).Debug("processIndividualCustomerOnApproval: Interim config check completed")

	// Create a background context for the transaction to avoid HTTP request timeout issues
	// Use context.Background() with a reasonable timeout instead of the HTTP request context
	// Increased timeout to 60 seconds to handle slow database operations
	logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Creating transaction context")
	txCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Starting database transaction")
	transactionStartTime := time.Now()
	err := postgres.WithTransaction(txCtx, s.db.DB, func(tx *sql.Tx) error {
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Inside transaction - generating UUIDs")
		// Generate UUIDs
		customerID = s.idGen.GenerateUUID()
		customerProfileID := s.idGen.GenerateUUID()
		logger.Log.WithFields(map[string]interface{}{
			"request_id":          request.RequestID,
			"customer_id":         customerID,
			"customer_profile_id": customerProfileID,
		}).Debug("processIndividualCustomerOnApproval: UUIDs generated")

		// Determine approval status and customer status
		approvalStatusStr := "Interim Approval"
		customerStatus := "Inactive"
		if status == "Approved" {
			approvalStatusStr = "Approved"
		}
		if request.CreationMode != nil && *request.CreationMode == models.CreationModeLegacy {
			customerStatus = "Active"
		}

		// Extract customer data from request
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Extracting customer data")
		customerDataSections, ok := request.Data["customerData"].([]interface{})
		if !ok {
			logger.Log.WithField("request_id", request.RequestID).Error("processIndividualCustomerOnApproval: Invalid customer data format")
			return fmt.Errorf("invalid customer data format")
		}
		logger.Log.WithFields(map[string]interface{}{
			"request_id":             request.RequestID,
			"customer_data_sections": len(customerDataSections),
		}).Debug("processIndividualCustomerOnApproval: Customer data sections extracted")

		// Convert to CustomerDataSection format
		var customerData []dto.CustomerDataSection
		for _, item := range customerDataSections {
			if itemMap, ok := item.(map[string]interface{}); ok {
				section := dto.CustomerDataSection{
					SectionName: getStringFromMap(itemMap, "sectionName"),
					Data:        itemMap["data"],
				}
				customerData = append(customerData, section)
			}
		}

		// Parse customer data
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Parsing customer data")
		customerProfileObj := parseCustomerData(customerData)
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Customer data parsed")

		// Validate phone number exists (for Individual customers)
		// Do this check BEFORE starting the transaction to avoid blocking
		// Use a separate context with shorter timeout for this read operation
		if phoneNumber := getString(customerProfileObj, "mobileNumber", "phoneNumber"); phoneNumber != nil {
			logger.Log.WithFields(map[string]interface{}{
				"request_id":   request.RequestID,
				"phone_number": *phoneNumber,
			}).Debug("processIndividualCustomerOnApproval: Checking phone number existence")
			phoneCheckStartTime := time.Now()
			phoneCheckCtx, phoneCancel := context.WithTimeout(context.Background(), 5*time.Second)
			exists, err := s.validationSvc.CheckPhoneExists(phoneCheckCtx, customerProfileObj, s.profileRepo)
			phoneCancel()
			phoneCheckDuration := time.Since(phoneCheckStartTime)
			if err != nil {
				logger.Log.WithError(err).WithFields(map[string]interface{}{
					"request_id":  request.RequestID,
					"duration_ms": phoneCheckDuration.Milliseconds(),
				}).Error("processIndividualCustomerOnApproval: Phone check failed")
				return fmt.Errorf("failed to check phone number: %w", err)
			}
			if exists {
				logger.Log.WithField("request_id", request.RequestID).Warn("processIndividualCustomerOnApproval: Phone number already exists")
				return utils.NewDuplicateError("Customer with Phone number already exists", nil)
			}
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": phoneCheckDuration.Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: Phone number check passed")
		}

		// Generate customer number and entity ID
		// Use txCtx instead of ctx to avoid HTTP request timeout
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Generating customer number")
		customerNumberStartTime := time.Now()
		customerNumber, err := s.idGen.GenerateCustomerNumber(txCtx, tx, s.profileRepo)
		customerNumberDuration := time.Since(customerNumberStartTime)
		if err != nil {
			logger.Log.WithError(err).WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": customerNumberDuration.Milliseconds(),
			}).Error("processIndividualCustomerOnApproval: Failed to generate customer number")
			return fmt.Errorf("failed to generate customer number: %w", err)
		}
		logger.Log.WithFields(map[string]interface{}{
			"request_id":      request.RequestID,
			"customer_number": customerNumber,
			"duration_ms":     customerNumberDuration.Milliseconds(),
		}).Debug("processIndividualCustomerOnApproval: Customer number generated")

		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Generating customer entity ID")
		entityIDStartTime := time.Now()
		customerEntityID, err := s.idGen.GenerateCustomerEntityID(txCtx, tx)
		entityIDDuration := time.Since(entityIDStartTime)
		if err != nil {
			logger.Log.WithError(err).WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": entityIDDuration.Milliseconds(),
			}).Error("processIndividualCustomerOnApproval: Failed to generate customer entity ID")
			return fmt.Errorf("failed to generate customer entity ID: %w", err)
		}
		logger.Log.WithFields(map[string]interface{}{
			"request_id":         request.RequestID,
			"customer_entity_id": customerEntityID,
			"duration_ms":        entityIDDuration.Milliseconds(),
		}).Debug("processIndividualCustomerOnApproval: Customer entity ID generated")

		// Process industry/sector/target
		// Use txCtx instead of ctx to avoid HTTP request timeout
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Processing industry/sector/target")
		rawIndustry := getField(customerProfileObj, "industry", "industryCode")
		rawSector := getField(customerProfileObj, "sector", "sectorCode")
		rawTarget := getField(customerProfileObj, "target", "targetCode")

		var sectorIDStr, industryIDStr, targetIDStr string
		if rawSector != nil || rawIndustry != nil || rawTarget != nil {
			validateStartTime := time.Now()
			sectorIDStr, industryIDStr, targetIDStr, _, _, _, err = s.validationSvc.ValidateSectorIndustryTarget(
				txCtx, getStringValue(rawSector), getStringValue(rawIndustry), getStringValue(rawTarget),
				s.sectorRepo, s.industryRepo, s.targetRepo,
			)
			validateDuration := time.Since(validateStartTime)
			if err != nil {
				logger.Log.WithError(err).WithFields(map[string]interface{}{
					"request_id":  request.RequestID,
					"duration_ms": validateDuration.Milliseconds(),
				}).Error("processIndividualCustomerOnApproval: Failed to validate sector/industry/target")
				return fmt.Errorf("failed to validate sector/industry/target: %w", err)
			}
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"sector_id":   sectorIDStr,
				"industry_id": industryIDStr,
				"target_id":   targetIDStr,
				"duration_ms": validateDuration.Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: Industry/sector/target validated")
		}

		// Parse customer info
		parsedInfo := parseCustomerInfo(customerProfileObj)

		// Convert sector/industry/target IDs to integers
		var sectorID, industryID, targetID *int
		if sectorIDStr != "" {
			if id, err := strconv.Atoi(sectorIDStr); err == nil {
				sectorID = &id
			}
		}
		if industryIDStr != "" {
			if id, err := strconv.Atoi(industryIDStr); err == nil {
				industryID = &id
			}
		}
		if targetIDStr != "" {
			if id, err := strconv.Atoi(targetIDStr); err == nil {
				targetID = &id
			}
		}

		// Create customer record
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Creating customer record")
		customerCreateStartTime := time.Now()
		approverName := userDetails.Name
		customer := &models.Customer{
			CustomerID:         customerID,
			CustomerType:       string(*request.CustomerType),
			Status:             customerStatus,
			Monitoring:         false,
			ApprovalStatus:     approvalStatusStr,
			TenantID:           nil, // Can be extracted from request data if needed
			Initiator:          request.Initiator,
			InitiatorID:        request.InitiatorID,
			Approver:           approverName,
			ApproverID:         userDetails.ID,
			Branch:             getStringValue(request.Branch),
			ApproverBranch:     userDetails.Branch,
			Department:         getStringValue(request.Department),
			CategoryOfCustomer: stringPtr("Regular"),
			CustomerSubType:    nil,
			RelatedEntity:      nil,
		}

		if err := s.customerRepo.Create(txCtx, tx, customer); err != nil {
			logger.Log.WithError(err).WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": time.Since(customerCreateStartTime).Milliseconds(),
			}).Error("processIndividualCustomerOnApproval: Failed to create customer")
			return fmt.Errorf("failed to create customer: %w", err)
		}
		customerCreateDuration := time.Since(customerCreateStartTime)
		logger.Log.WithFields(map[string]interface{}{
			"request_id":  request.RequestID,
			"customer_id": customerID,
			"duration_ms": customerCreateDuration.Milliseconds(),
		}).Debug("processIndividualCustomerOnApproval: Customer record created")

		// Build customer profile data JSON
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Building customer profile data")
		profileDataStartTime := time.Now()
		customerProfileData := buildCustomerProfileDataJSON(customerProfileObj, customerEntityID, customerNumber)
		logger.Log.WithFields(map[string]interface{}{
			"request_id":  request.RequestID,
			"duration_ms": time.Since(profileDataStartTime).Milliseconds(),
		}).Debug("processIndividualCustomerOnApproval: Customer profile data built")

		// Create customer profile
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Creating customer profile")
		profileCreateStartTime := time.Now()
		profile := &models.CustomerProfile{
			CustomerProfileID:   customerProfileID,
			CustomerID:          customerID,
			CustomerNumber:      customerNumber,
			Title:               parsedInfo.Title,
			FirstName:           parsedInfo.FirstName,
			Surname:             parsedInfo.Surname,
			OtherNames:          parsedInfo.OtherNames,
			MobileNumber:        parsedInfo.MobileNumber,
			EmailAddress:        parsedInfo.EmailAddress,
			DateOfBirth:         parseDateString(parsedInfo.DateOfBirth),
			BVN:                 parsedInfo.BVN,
			NIN:                 parsedInfo.NIN,
			Nationality:         parsedInfo.Nationality,
			SectorID:            sectorID,
			IndustryID:          industryID,
			TargetID:            targetID,
			CustomerProfileData: customerProfileData,
		}

		if err := s.profileRepo.Create(txCtx, tx, profile); err != nil {
			logger.Log.WithError(err).WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": time.Since(profileCreateStartTime).Milliseconds(),
			}).Error("processIndividualCustomerOnApproval: Failed to create customer profile")
			return fmt.Errorf("failed to create customer profile: %w", err)
		}
		profileCreateDuration := time.Since(profileCreateStartTime)
		logger.Log.WithFields(map[string]interface{}{
			"request_id":  request.RequestID,
			"duration_ms": profileCreateDuration.Milliseconds(),
		}).Debug("processIndividualCustomerOnApproval: Customer profile created")

		// Update request with customerId
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Updating request with customer ID")
		if err := s.requestRepo.UpdateCustomerID(txCtx, tx, request.RequestID, customerID); err != nil {
			logger.Log.WithError(err).WithField("request_id", request.RequestID).Error("processIndividualCustomerOnApproval: Failed to update request")
			return fmt.Errorf("failed to update request: %w", err)
		}

		// Update waiver request if exists
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Updating waiver request if exists")
		_ = s.waiverRequestRepo.UpdateCustomerID(txCtx, tx, request.RequestID, customerID)

		// Update activity logs
		// Note: Activity log update would require a new method in repository

		// Create risk assessments
		if riskAssessmentData, ok := request.Data["riskAssessmentData"].([]interface{}); ok {
			logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Creating risk assessments")
			riskAssessmentStartTime := time.Now()
			if err := s.createRiskAssessments(txCtx, tx, customerID, riskAssessmentData); err != nil {
				logger.Log.WithError(err).WithFields(map[string]interface{}{
					"request_id":  request.RequestID,
					"duration_ms": time.Since(riskAssessmentStartTime).Milliseconds(),
				}).Error("processIndividualCustomerOnApproval: Failed to create risk assessments")
				return fmt.Errorf("failed to create risk assessments: %w", err)
			}
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": time.Since(riskAssessmentStartTime).Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: Risk assessments created")
		}

		// Create bank accounts
		if accountData, ok := request.Data["accountData"].([]interface{}); ok {
			logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Creating bank accounts")
			accountsStartTime := time.Now()
			if err := s.createAccounts(txCtx, tx, customerID, customerProfileID, accountData); err != nil {
				logger.Log.WithError(err).WithFields(map[string]interface{}{
					"request_id":  request.RequestID,
					"duration_ms": time.Since(accountsStartTime).Milliseconds(),
				}).Error("processIndividualCustomerOnApproval: Failed to create accounts")
				return fmt.Errorf("failed to create accounts: %w", err)
			}
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": time.Since(accountsStartTime).Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: Bank accounts created")
		}

		// Create signatories
		if signatoryData, ok := request.Data["signatoryData"].([]interface{}); ok {
			logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Creating signatories")
			signatoriesStartTime := time.Now()
			if err := s.createSignatories(txCtx, tx, customerID, request.RequestID, signatoryData); err != nil {
				logger.Log.WithError(err).WithFields(map[string]interface{}{
					"request_id":  request.RequestID,
					"duration_ms": time.Since(signatoriesStartTime).Milliseconds(),
				}).Error("processIndividualCustomerOnApproval: Failed to create signatories")
				return fmt.Errorf("failed to create signatories: %w", err)
			}
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": time.Since(signatoriesStartTime).Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: Signatories created")
		}

		// Create references
		if referenceData, ok := request.Data["referenceData"].([]interface{}); ok {
			logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Creating references")
			referencesStartTime := time.Now()
			if err := s.createReferences(txCtx, tx, customerID, referenceData); err != nil {
				logger.Log.WithError(err).WithFields(map[string]interface{}{
					"request_id":  request.RequestID,
					"duration_ms": time.Since(referencesStartTime).Milliseconds(),
				}).Error("processIndividualCustomerOnApproval: Failed to create references")
				return fmt.Errorf("failed to create references: %w", err)
			}
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": time.Since(referencesStartTime).Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: References created")
		}

		// Create activity log
		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Creating activity log")
		activityLogStartTime := time.Now()
		activityLog := &models.ActivityLog{
			ActivityLogID: s.idGen.GenerateUUID(),
			CustomerID:    &customerID,
			RequestID:     &request.RequestID,
			Description:   fmt.Sprintf("Customer creation request approved by %s", userDetails.Name),
			Reason:        nil,
		}

		if err := s.activityLogRepo.Create(txCtx, tx, activityLog); err != nil {
			logger.Log.WithError(err).WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": time.Since(activityLogStartTime).Milliseconds(),
			}).Warn("processIndividualCustomerOnApproval: Failed to create activity log")
		} else {
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": time.Since(activityLogStartTime).Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: Activity log created")
		}

		// Create grace period data if Interim Approval
		if approvalStatusStr == "Interim Approval" && hasInterimConfig {
			logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Creating grace period data")
			gracePeriodStartTime := time.Now()
			if err := s.createGracePeriodData(txCtx, tx, customerID, interimConfig); err != nil {
				logger.Log.WithError(err).WithFields(map[string]interface{}{
					"request_id":  request.RequestID,
					"duration_ms": time.Since(gracePeriodStartTime).Milliseconds(),
				}).Error("processIndividualCustomerOnApproval: Failed to create grace period data")
				return fmt.Errorf("failed to create grace period data: %w", err)
			}
			logger.Log.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"duration_ms": time.Since(gracePeriodStartTime).Milliseconds(),
			}).Debug("processIndividualCustomerOnApproval: Grace period data created")
		}

		logger.Log.WithField("request_id", request.RequestID).Debug("processIndividualCustomerOnApproval: Transaction completed successfully")
		return nil
	})
	transactionDuration := time.Since(transactionStartTime)

	if err != nil {
		logger.Log.WithError(err).WithFields(map[string]interface{}{
			"request_id":  request.RequestID,
			"duration_ms": transactionDuration.Milliseconds(),
		}).Error("processIndividualCustomerOnApproval: Transaction failed")
		return uuid.Nil, err
	}

	totalDuration := time.Since(startTime)
	logger.Log.WithFields(map[string]interface{}{
		"request_id":              request.RequestID,
		"customer_id":             customerID,
		"total_duration_ms":       totalDuration.Milliseconds(),
		"transaction_duration_ms": transactionDuration.Milliseconds(),
	}).Info("processIndividualCustomerOnApproval: Customer creation completed successfully")

	return customerID, nil
}

func (s *requestService) processSMECustomerOnApproval(ctx context.Context, request *models.Request, status string, userDetails *UserDetails) (uuid.UUID, error) {
	// Similar to Individual but with SME-specific logic
	// For now, reuse Individual logic with adjustments
	return s.processIndividualCustomerOnApproval(ctx, request, status, userDetails)
}

// Helper functions
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseDateString(dateStr *string) *time.Time {
	if dateStr == nil || *dateStr == "" {
		return nil
	}
	if t, err := utils.ParseDate(*dateStr); err == nil {
		return &t
	}
	return nil
}
