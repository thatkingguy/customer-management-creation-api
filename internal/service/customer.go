package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/repository"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
	"github.com/sterling-retailcore-team/customer-management-creation-api/pkg/postgres"
)

// CustomerSubTypeMap maps customer sub-type strings to integers
var CustomerSubTypeMap = map[string]int{
	"Individual":      1,
	"PrivateBankInd":  10,
	"Youth Corper":    11,
	"Salary Customer": 12,
	"Minor":           13,
	"Student":         14,
	"Not for Profit":  15,
	"Joint Customers": 16,
	"PrivateBankCorp": 17,
	"MSME":            18,
	"Corporate":       2,
	"EMBASSY/HIGH CO": 22,
	"Govt./MDA (LG)":  23,
	"EX-Staff":        24,
	"Enterprise":      3,
	"Govt./MDA (FGN)": 4,
	"Internal":        5,
	"Ind. Staff":      6,
	"Govt./MDA (ST)":  7,
	"DSE":             8,
	"Trainee":         9,
}

type CustomerService interface {
	CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CreateCustomerResponse, error)
}

type customerService struct {
	db              *sqlx.DB
	customerRepo    repository.CustomerRepository
	profileRepo     repository.CustomerProfileRepository
	lookupRepo      repository.CustomerProfileLookupRepository
	activityLogRepo repository.ActivityLogRepository
	sectorRepo      repository.SectorRepository
	industryRepo    repository.IndustryRepository
	targetRepo      repository.TargetRepository
	validationSvc   ValidationService
	idGen           IDGeneratorService
}

type CreateCustomerRequest struct {
	Branch        string
	CustomerType  string
	BioData       map[string]interface{}
	TenantID      *uuid.UUID
	Department    string
	RelatedEntity *uuid.UUID
	FlagCustomer  *FlagCustomer
}

type FlagCustomer struct {
	Status bool
}

type CreateCustomerResponse struct {
	CustomerID uuid.UUID
	RequestID  uuid.UUID
}

func NewCustomerService(
	db *sqlx.DB,
	customerRepo repository.CustomerRepository,
	profileRepo repository.CustomerProfileRepository,
	lookupRepo repository.CustomerProfileLookupRepository,
	activityLogRepo repository.ActivityLogRepository,
	sectorRepo repository.SectorRepository,
	industryRepo repository.IndustryRepository,
	targetRepo repository.TargetRepository,
	validationSvc ValidationService,
	idGen IDGeneratorService,
) CustomerService {
	return &customerService{
		db:              db,
		customerRepo:    customerRepo,
		profileRepo:     profileRepo,
		lookupRepo:      lookupRepo,
		activityLogRepo: activityLogRepo,
		sectorRepo:      sectorRepo,
		industryRepo:    industryRepo,
		targetRepo:      targetRepo,
		validationSvc:   validationSvc,
		idGen:           idGen,
	}
}

func (s *customerService) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CreateCustomerResponse, error) {
	// Normalize customer type
	customerType := utils.NormalizeCustomerType(req.CustomerType)
	if customerType != "Individual" && customerType != "SME" {
		return nil, utils.NewValidationError("Customer type must be 'Individual' or 'SME'", nil)
	}

	// Normalize and clean bioData
	bioData := s.normalizeBioData(req.BioData)

	// Extract branch (priority: bioData.branch > req.Branch)
	branch := req.Branch
	if bioBranch, ok := bioData["branch"].(string); ok && bioBranch != "" {
		branch = bioBranch
	}
	if branch == "" {
		return nil, utils.NewValidationError("The payload is missing the 'branch' field", nil)
	}

	// Validate bioData
	if err := s.validationSvc.ValidateCustomerBioData(ctx, bioData, customerType, true); err != nil {
		logger.Log.WithError(err).Error("CreateCustomer: BioData validation failed")
		return nil, err
	}

	// Check phone exists (for channel requests, this doesn't throw error)
	phoneExists, err := s.validationSvc.CheckPhoneExists(ctx, bioData, s.profileRepo)
	if err != nil {
		logger.Log.WithError(err).Error("CreateCustomer: Phone check failed")
		return nil, utils.NewDatabaseError("Failed to check phone number", err)
	}
	if phoneExists {
		return nil, utils.NewDuplicateError("Customer with Phone number already exists", nil)
	}

	// Validate sector/industry/target BEFORE transaction (read-only lookups, avoids potential deadlocks)
	sectorCode, _ := getFieldValueFromMap(bioData, "sectorCode")
	industryCode, _ := getFieldValueFromMap(bioData, "industryCode")
	targetCode, _ := getFieldValueFromMap(bioData, "targetCode")

	// Convert to strings efficiently
	var sectorCodeStr, industryCodeStr, targetCodeStr string
	if sectorCode != nil {
		sectorCodeStr = fmt.Sprintf("%v", sectorCode)
	}
	if industryCode != nil {
		industryCodeStr = fmt.Sprintf("%v", industryCode)
	}
	if targetCode != nil {
		targetCodeStr = fmt.Sprintf("%v", targetCode)
	}

	sectorID, industryID, targetID, sectorDesc, industryDesc, targetDesc, err := s.validationSvc.ValidateSectorIndustryTarget(
		ctx, sectorCodeStr, industryCodeStr, targetCodeStr,
		s.sectorRepo, s.industryRepo, s.targetRepo,
	)
	if err != nil {
		logger.Log.WithError(err).Error("CreateCustomer: Sector/industry/target validation failed")
		return nil, err
	}

	// Start transaction
	var customerID uuid.UUID
	var requestID uuid.UUID

	err = postgres.WithTransaction(ctx, s.db.DB, func(tx *sql.Tx) error {
		// Lookup customer (duplicate check and insert into lookup table)
		if err := s.validationSvc.LookupCustomer(ctx, bioData, customerType, s.profileRepo, s.lookupRepo, tx); err != nil {
			logger.Log.WithError(err).Error("CreateCustomer: Lookup customer failed")
			return err
		}

		// Generate IDs
		customerID = s.idGen.GenerateUUID()
		customerProfileID := s.idGen.GenerateUUID()

		customerNumber, err := s.idGen.GenerateCustomerNumber(ctx, tx, s.profileRepo)
		if err != nil {
			logger.Log.WithError(err).Error("CreateCustomer: Failed to generate customer number")
			return fmt.Errorf("failed to generate customer number: %w", err)
		}

		customerEntityID, err := s.idGen.GenerateCustomerEntityID(ctx, tx)
		if err != nil {
			logger.Log.WithError(err).Error("CreateCustomer: Failed to generate customer entity ID")
			return fmt.Errorf("failed to generate customer entity ID: %w", err)
		}

		// Parse IDs as integers (sector/industry/target use integer IDs, not UUIDs)
		// Use strconv for better performance than fmt.Sscanf
		var sectorIDInt, industryIDInt, targetIDInt int
		sectorIDInt, _ = strconv.Atoi(sectorID)
		industryIDInt, _ = strconv.Atoi(industryID)
		targetIDInt, _ = strconv.Atoi(targetID)

		// Convert to integer pointers for customer_profile
		sectorIDPtr := &sectorIDInt
		industryIDPtr := &industryIDInt
		targetIDPtr := &targetIDInt

		// Prepare customer record
		initiatorID := s.idGen.GenerateUUID()
		tenantID := req.TenantID
		if tenantID == nil {
			generatedTenantID := s.idGen.GenerateUUID()
			tenantID = &generatedTenantID
		}

		department := req.Department
		if department == "" {
			department = "SYSTEM_DEPARTMENT"
		}

		monitoring := false
		if req.FlagCustomer != nil {
			monitoring = req.FlagCustomer.Status
		}

		categoryOfCustomer := "Regular"
		if cat, ok := getFieldValueFromMap(bioData, "categoryOfCustomer"); ok {
			if catStr, ok := cat.(string); ok && catStr != "" {
				categoryOfCustomer = catStr
			}
		}

		var customerSubType *int
		if subTypeVal, ok := getFieldValueFromMap(bioData, "customerSubType"); ok {
			if subTypeStr, ok := subTypeVal.(string); ok {
				if subTypeInt, ok := CustomerSubTypeMap[subTypeStr]; ok {
					customerSubType = &subTypeInt
				}
			}
		}

		customer := &models.Customer{
			CustomerID:         customerID,
			CustomerType:       customerType,
			Status:             "Active",
			Monitoring:         monitoring,
			ApprovalStatus:     "Approved",
			TenantID:           tenantID,
			Initiator:          "SYSTEM_USER",
			InitiatorID:        initiatorID,
			Approver:           "SYSTEM_USER",
			ApproverID:         initiatorID,
			Branch:             branch,
			ApproverBranch:     branch,
			Department:         department,
			CategoryOfCustomer: &categoryOfCustomer,
			CustomerSubType:    customerSubType,
			RelatedEntity:      req.RelatedEntity,
		}

		// Create customer
		if err := s.customerRepo.Create(ctx, tx, customer); err != nil {
			logger.Log.WithError(err).Error("CreateCustomer: Failed to create customer record")
			return fmt.Errorf("failed to create customer: %w", err)
		}

		// Prepare customer profile
		profile := s.prepareCustomerProfile(
			customerProfileID, customerID, customerNumber, customerEntityID,
			bioData, customerType, sectorIDPtr, industryIDPtr, targetIDPtr,
			sectorDesc, industryDesc, targetDesc,
		)

		// Create customer profile
		if err := s.profileRepo.Create(ctx, tx, profile); err != nil {
			logger.Log.WithError(err).Error("CreateCustomer: Failed to create customer profile record")
			return fmt.Errorf("failed to create customer profile: %w", err)
		}

		// Create activity log
		activityLog := &models.ActivityLog{
			ActivityLogID: s.idGen.GenerateUUID(),
			Description:   "Customer created via channel direct flow by SYSTEM_USER",
			CustomerID:    customerID,
		}

		if err := s.activityLogRepo.Create(ctx, tx, activityLog); err != nil {
			logger.Log.WithError(err).Error("CreateCustomer: Failed to create activity log")
			return fmt.Errorf("failed to create activity log: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Log.WithError(err).Error("CreateCustomer: Transaction failed")
		return nil, err
	}

	requestID = s.idGen.GenerateUUID()
	logger.Log.WithFields(map[string]interface{}{
		"customer_id": customerID,
		"request_id":  requestID,
	}).Info("CreateCustomer: Customer creation completed successfully")

	return &CreateCustomerResponse{
		CustomerID: customerID,
		RequestID:  requestID,
	}, nil
}

func (s *customerService) normalizeBioData(bioData map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{}, len(bioData))

	// Field name mapping (pre-computed lowercase keys for O(1) lookup)
	fieldMap := map[string]string{
		"bvn":                        "bvn",
		"nin":                        "nin",
		"email":                      "emailAddress",
		"emailaddress":               "emailAddress",
		"mobilenumber":               "mobileNumber",
		"mobile":                     "mobileNumber",
		"phonenumber":                "mobileNumber",
		"contactnumber":              "mobileNumber",
		"chooseanid":                 "chooseAnId",
		"idnumber":                   "idNumber",
		"certificateofincorporation": "certificateOfIncorporation",
		"rcnumber":                   "certificateOfIncorporation",
	}

	for k, v := range bioData {
		// Skip null/empty values (fast path)
		if v == nil {
			continue
		}
		if str, ok := v.(string); ok {
			if len(str) == 0 {
				continue
			}
			// Only trim if string is not empty (avoid unnecessary allocation)
			if strings.TrimSpace(str) == "" {
				continue
			}
		}

		// Normalize field names
		keyLower := strings.ToLower(k)
		if normalizedKey, ok := fieldMap[keyLower]; ok {
			normalized[normalizedKey] = v
		} else {
			normalized[k] = v
		}
	}

	return normalized
}

func (s *customerService) prepareCustomerProfile(
	customerProfileID, customerID uuid.UUID,
	customerNumber, customerEntityID string,
	bioData map[string]interface{},
	customerType string,
	sectorID, industryID, targetID *int,
	sectorDesc, industryDesc, targetDesc string,
) *models.CustomerProfile {
	// Pre-allocate JSONB map with estimated size
	profileData := make(models.JSONB, len(bioData)+4) // +4 for customerEntityId, sector, industry, target

	// Copy all bioData fields to customerProfileData (single pass)
	for k, v := range bioData {
		profileData[k] = v
	}

	// Add specific fields
	profileData["customerEntityId"] = customerEntityID
	if sectorCode, ok := bioData["sectorCode"]; ok {
		profileData["sector"] = sectorCode
	} else if sectorCode, ok := getFieldValueFromMap(bioData, "sectorCode"); ok {
		profileData["sector"] = sectorCode
	}
	if industryCode, ok := bioData["industryCode"]; ok {
		profileData["industry"] = industryCode
	} else if industryCode, ok := getFieldValueFromMap(bioData, "industryCode"); ok {
		profileData["industry"] = industryCode
	}
	if targetCode, ok := bioData["targetCode"]; ok {
		profileData["target"] = targetCode
	} else if targetCode, ok := getFieldValueFromMap(bioData, "targetCode"); ok {
		profileData["target"] = targetCode
	}

	profile := &models.CustomerProfile{
		CustomerProfileID:   customerProfileID,
		CustomerID:          customerID,
		CustomerNumber:      customerNumber,
		SectorID:            sectorID,
		IndustryID:          industryID,
		TargetID:            targetID,
		CustomerProfileData: profileData,
	}

	// Set fields based on customer type
	if customerType == "Individual" {
		if val, ok := getFieldValueFromMap(bioData, "title"); ok {
			if str, ok := val.(string); ok {
				profile.Title = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "firstName"); ok {
			if str, ok := val.(string); ok {
				profile.FirstName = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "surname"); ok {
			if str, ok := val.(string); ok {
				profile.Surname = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "otherNames"); ok {
			if str, ok := val.(string); ok {
				profile.OtherNames = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "mobileNumber"); ok {
			if str, ok := val.(string); ok {
				profile.MobileNumber = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "emailAddress"); ok {
			if str, ok := val.(string); ok {
				profile.EmailAddress = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "dateOfBirth"); ok {
			if str, ok := val.(string); ok {
				if dob, err := utils.ParseDate(str); err == nil {
					profile.DateOfBirth = &dob
				}
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "bvn"); ok {
			if str, ok := val.(string); ok {
				profile.BVN = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "nin"); ok {
			if str, ok := val.(string); ok {
				profile.NIN = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "nationality"); ok {
			if str, ok := val.(string); ok {
				profile.Nationality = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "introducer"); ok {
			if str, ok := val.(string); ok {
				if introUUID, err := uuid.Parse(str); err == nil {
					profile.Introducer = &introUUID
				}
			}
		}
	} else if customerType == "SME" {
		if val, ok := getFieldValueFromMap(bioData, "companyNameBusiness"); ok {
			if str, ok := val.(string); ok {
				profile.CompanyNameBusiness = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "taxIdentificationNumber"); ok {
			if str, ok := val.(string); ok {
				profile.TaxIdentificationNumber = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "certificateOfIncorporation"); ok {
			if str, ok := val.(string); ok {
				profile.CertificateOfIncorporation = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "dateOfRegistration"); ok {
			if str, ok := val.(string); ok {
				if dor, err := utils.ParseDate(str); err == nil {
					profile.DateOfRegistration = &dor
				}
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "categoryOfBusiness"); ok {
			if str, ok := val.(string); ok {
				profile.CategoryOfBusiness = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "mobileNumber"); ok {
			if str, ok := val.(string); ok {
				profile.MobileNumber = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "emailAddress"); ok {
			if str, ok := val.(string); ok {
				profile.EmailAddress = &str
			}
		}
		if val, ok := getFieldValueFromMap(bioData, "introducer"); ok {
			if str, ok := val.(string); ok {
				if introUUID, err := uuid.Parse(str); err == nil {
					profile.Introducer = &introUUID
				}
			}
		}
	}

	return profile
}

// getFieldValueFromMap is a helper to extract field value with case-insensitive matching
// Optimized: Build lowercase map once per bioData map
func getFieldValueFromMap(bioData map[string]interface{}, fieldName string) (interface{}, bool) {
	// Try exact match first (fast path - O(1))
	if val, ok := bioData[fieldName]; ok {
		return val, true
	}

	// Try case-insensitive match (O(n) but only if exact match fails)
	fieldLower := strings.ToLower(fieldName)
	for k, v := range bioData {
		if strings.ToLower(k) == fieldLower {
			return v, true
		}
	}

	return nil, false
}
