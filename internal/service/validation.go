package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/repository"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

type ValidationService interface {
	ValidateCustomerBioData(ctx context.Context, bioData map[string]interface{}, customerType string, isChannelRequest bool) error
	CheckPhoneExists(ctx context.Context, bioData map[string]interface{}, profileRepo repository.CustomerProfileRepository) (bool, error)
	LookupCustomer(ctx context.Context, bioData map[string]interface{}, customerType string, profileRepo repository.CustomerProfileRepository, lookupRepo repository.CustomerProfileLookupRepository, tx *sql.Tx) error
	ValidateSectorIndustryTarget(ctx context.Context, sectorCode, industryCode, targetCode string, sectorRepo repository.SectorRepository, industryRepo repository.IndustryRepository, targetRepo repository.TargetRepository) (sectorID, industryID, targetID string, sectorDesc, industryDesc, targetDesc string, err error)
}

type validationService struct {
	idGen IDGeneratorService
}

func NewValidationService(idGen IDGeneratorService) ValidationService {
	return &validationService{idGen: idGen}
}

// Field name aliases map for case-insensitive matching
var fieldAliases = map[string]string{
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
	"id number":                  "idNumber",
	"certificateofincorporation": "certificateOfIncorporation",
	"rcnumber":                   "certificateOfIncorporation",
}

// getFieldValue extracts field value from bioData with case-insensitive matching
func getFieldValue(bioData map[string]interface{}, fieldName string) (interface{}, bool) {
	// Try exact match first
	if val, ok := bioData[fieldName]; ok {
		return val, true
	}

	// Try case-insensitive match
	fieldLower := strings.ToLower(fieldName)
	for k, v := range bioData {
		if strings.ToLower(k) == fieldLower {
			return v, true
		}
	}

	// Try aliases
	if alias, ok := fieldAliases[fieldLower]; ok {
		if val, ok := bioData[alias]; ok {
			return val, true
		}
		// Try case-insensitive alias match
		for k, v := range bioData {
			if strings.EqualFold(k, alias) {
				return v, true
			}
		}
	}

	return nil, false
}

// getFieldValueFromBioData extracts field value from bioData with case-insensitive matching
func getFieldValueFromBioData(bioData map[string]interface{}, fieldName string) (interface{}, bool) {
	return getFieldValue(bioData, fieldName)
}

func (s *validationService) ValidateCustomerBioData(ctx context.Context, bioData map[string]interface{}, customerType string, isChannelRequest bool) error {
	// Validate BVN (if present)
	if bvnVal, ok := getFieldValueFromBioData(bioData, "bvn"); ok && bvnVal != nil {
		bvn, ok := bvnVal.(string)
		if !ok {
			return utils.NewValidationError("BVN must be a string.", nil)
		}
		if !utils.IsNumeric(bvn) {
			return utils.NewValidationError("BVN must contain only numbers.", nil)
		}
		if utils.ContainsAsterisk(bvn) {
			return utils.NewValidationError("BVN must not be redacted (should not contain '*').", nil)
		}
	}

	// Validate NIN (if present)
	if ninVal, ok := getFieldValueFromBioData(bioData, "nin"); ok && ninVal != nil {
		nin, ok := ninVal.(string)
		if !ok {
			return utils.NewValidationError("NIN must be a string.", nil)
		}
		if !utils.IsNumeric(nin) {
			return utils.NewValidationError("NIN must contain only numbers.", nil)
		}
		if utils.ContainsAsterisk(nin) {
			return utils.NewValidationError("NIN must not be redacted (should not contain '*').", nil)
		}
	}

	// Validate Tax Identification Number (if present)
	if tinVal, ok := getFieldValueFromBioData(bioData, "taxIdentificationNumber"); ok && tinVal != nil {
		tin, ok := tinVal.(string)
		if !ok {
			return utils.NewValidationError("Tax Identification Number must be a string.", nil)
		}
		if utils.ContainsAsterisk(tin) {
			return utils.NewValidationError("Tax Identification Number must not be redacted (should not contain '*').", nil)
		}
	}

	// Validate Certificate of Incorporation/RC Number (if present)
	if certVal, ok := getFieldValueFromBioData(bioData, "certificateOfIncorporation"); ok && certVal != nil {
		cert, ok := certVal.(string)
		if !ok {
			// Try rcNumber
			if rcVal, ok := getFieldValueFromBioData(bioData, "rcNumber"); ok && rcVal != nil {
				cert, ok = rcVal.(string)
			}
		}
		if ok && utils.ContainsAsterisk(cert) {
			return utils.NewValidationError("Certificate of Incorporation must not be redacted (should not contain '*').", nil)
		}
	}

	// Validate Date of Birth (if present)
	if dobVal, ok := getFieldValueFromBioData(bioData, "dateOfBirth"); ok && dobVal != nil {
		dobStr, ok := dobVal.(string)
		if !ok {
			return utils.NewValidationError("Date must be a valid date string in ISO 8601 or RFC 2822 format.", nil)
		}
		_, err := utils.ParseDate(dobStr)
		if err != nil {
			return utils.NewValidationError("Date must be a valid date string in ISO 8601 or RFC 2822 format.", err)
		}
	}

	// Phone number duplicate check (if phone exists and age > 18)
	// Note: For channel requests, we skip throwing error (just return)
	if phoneVal, ok := getFieldValueFromBioData(bioData, "mobileNumber"); ok && phoneVal != nil {
		if dobVal, ok := getFieldValueFromBioData(bioData, "dateOfBirth"); ok && dobVal != nil {
			dobStr, ok := dobVal.(string)
			if ok {
				dob, err := utils.ParseDate(dobStr)
				if err == nil {
					age := utils.CalculateAge(dob)
					if age > 18 && !isChannelRequest {
						// This will be checked separately in CheckPhoneExists
						// For channel requests, we don't throw error here
					}
				}
			}
		}
	}

	return nil
}

func (s *validationService) CheckPhoneExists(ctx context.Context, bioData map[string]interface{}, profileRepo repository.CustomerProfileRepository) (bool, error) {
	// Extract phone number
	var phone string
	phoneKeys := []string{"phoneNumber", "mobile", "mobileNumber", "contactNumber"}
	for _, key := range phoneKeys {
		if val, ok := getFieldValueFromBioData(bioData, key); ok && val != nil {
			if phoneStr, ok := val.(string); ok && phoneStr != "" {
				phone = phoneStr
				break
			}
		}
	}

	if phone == "" {
		return false, nil
	}

	// Check date of birth
	dobVal, ok := getFieldValueFromBioData(bioData, "dateOfBirth")
	if !ok || dobVal == nil {
		return false, nil // Skip check if no DOB
	}

	dobStr, ok := dobVal.(string)
	if !ok {
		return false, nil
	}

	dob, err := utils.ParseDate(dobStr)
	if err != nil {
		return false, nil // Skip check if invalid date
	}

	age := utils.CalculateAge(dob)
	if age <= 18 {
		return false, nil // Skip check for minors
	}

	// Check if phone exists
	exists, err := profileRepo.CheckPhoneExists(ctx, phone)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *validationService) LookupCustomer(ctx context.Context, bioData map[string]interface{}, customerType string, profileRepo repository.CustomerProfileRepository, lookupRepo repository.CustomerProfileLookupRepository, tx *sql.Tx) error {
	normalizedType := utils.NormalizeCustomerType(customerType)

	if normalizedType == "Individual" {
		return s.lookupIndividualCustomer(ctx, bioData, lookupRepo, tx)
	} else if normalizedType == "SME" {
		return s.lookupSMECustomer(ctx, bioData, lookupRepo, tx)
	}

	return nil
}

func (s *validationService) lookupIndividualCustomer(ctx context.Context, bioData map[string]interface{}, lookupRepo repository.CustomerProfileLookupRepository, tx *sql.Tx) error {
	// Extract fields
	firstNameVal, _ := getFieldValueFromBioData(bioData, "firstName")
	surnameVal, _ := getFieldValueFromBioData(bioData, "surname")
	dobVal, _ := getFieldValueFromBioData(bioData, "dateOfBirth")
	bvnVal, _ := getFieldValueFromBioData(bioData, "bvn")
	ninVal, _ := getFieldValueFromBioData(bioData, "nin")

	firstName, _ := firstNameVal.(string)
	surname, _ := surnameVal.(string)

	if firstName == "" || surname == "" {
		return utils.NewValidationError("firstName and surname are required for Individual customers", nil)
	}

	var dob time.Time
	if dobStr, ok := dobVal.(string); ok {
		parsed, err := utils.ParseDate(dobStr)
		if err != nil {
			return utils.NewValidationError("Invalid dateOfBirth", err)
		}
		dob = parsed
	}

	var bvn, nin *string
	if bvnStr, ok := bvnVal.(string); ok && bvnStr != "" {
		bvn = &bvnStr
	}
	if ninStr, ok := ninVal.(string); ok && ninStr != "" {
		nin = &ninStr
	}

	// Insert directly into lookup table - database unique constraint will catch duplicates
	lookupID := uuid.New()
	lookup := &models.CustomerProfileLookup{
		CustomerProfileLookupID: lookupID,
		FirstName:               &firstName,
		Surname:                 &surname,
		DateOfBirth:             &dob,
		BVN:                     bvn,
		NIN:                     nin,
	}

	err := lookupRepo.Create(ctx, tx, lookup)
	if err != nil {
		// Check if it's a unique constraint violation
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "duplicate key") ||
			strings.Contains(errStr, "unique constraint") ||
			strings.Contains(errStr, "violates unique constraint") {
			// Determine which field caused the duplicate
			if bvn != nil {
				return utils.NewDuplicateError("Customer with BVN already exists", nil)
			}
			if nin != nil {
				return utils.NewDuplicateError("Customer with NIN already exists", nil)
			}
			return utils.NewDuplicateError("Customer with details passed already exists", nil)
		}
		return utils.NewDatabaseError("Failed to create customer lookup", err)
	}

	return nil
}

func (s *validationService) lookupSMECustomer(ctx context.Context, bioData map[string]interface{}, lookupRepo repository.CustomerProfileLookupRepository, tx *sql.Tx) error {
	// Extract fields
	companyNameVal, _ := getFieldValueFromBioData(bioData, "companyNameBusiness")
	taxIDVal, _ := getFieldValueFromBioData(bioData, "taxIdentificationNumber")
	dateRegVal, _ := getFieldValueFromBioData(bioData, "dateOfRegistration")
	certVal, _ := getFieldValueFromBioData(bioData, "certificateOfIncorporation")

	companyName, _ := companyNameVal.(string)
	if companyName == "" {
		return utils.NewValidationError("companyNameBusiness is required for SME customers", nil)
	}

	var taxID *string
	if taxIDStr, ok := taxIDVal.(string); ok && taxIDStr != "" {
		taxID = &taxIDStr
	}

	var dateReg *time.Time
	if dateRegStr, ok := dateRegVal.(string); ok && dateRegStr != "" {
		parsed, err := utils.ParseDate(dateRegStr)
		if err != nil {
			return utils.NewValidationError("Invalid dateOfRegistration", err)
		}
		dateReg = &parsed
	}

	var cert *string
	if certStr, ok := certVal.(string); ok && certStr != "" {
		cert = &certStr
	} else if rcVal, ok := getFieldValueFromBioData(bioData, "rcNumber"); ok {
		if rcStr, ok := rcVal.(string); ok && rcStr != "" {
			cert = &rcStr
		}
	}

	// Insert directly into lookup table - database unique constraint will catch duplicates
	lookupID := uuid.New()
	lookup := &models.CustomerProfileLookup{
		CustomerProfileLookupID:    lookupID,
		CompanyNameBusiness:        &companyName,
		TaxIdentificationNumber:    taxID,
		DateOfRegistration:         dateReg,
		CertificateOfIncorporation: cert,
	}

	err := lookupRepo.Create(ctx, tx, lookup)
	if err != nil {
		// Check if it's a unique constraint violation
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "duplicate key") ||
			strings.Contains(errStr, "unique constraint") ||
			strings.Contains(errStr, "violates unique constraint") {
			return utils.NewDuplicateError("Customer with companyNameBusiness already exists", nil)
		}
		return utils.NewDatabaseError("Failed to create customer lookup", err)
	}

	return nil
}

func (s *validationService) ValidateSectorIndustryTarget(ctx context.Context, sectorCode, industryCode, targetCode string, sectorRepo repository.SectorRepository, industryRepo repository.IndustryRepository, targetRepo repository.TargetRepository) (sectorID, industryID, targetID string, sectorDesc, industryDesc, targetDesc string, err error) {
	// Extract numeric codes (handles "4200" or "4200-ABC" formats)
	// Repositories will handle validation and return errors for invalid formats
	sectorCode = utils.ExtractNumericCode(sectorCode)
	industryCode = utils.ExtractNumericCode(industryCode)
	targetCode = utils.ExtractNumericCode(targetCode)

	// Parallelize lookups for better performance
	type sectorResult struct {
		sector *models.Sector
		err    error
	}
	type industryResult struct {
		industry *models.Industry
		err      error
	}
	type targetResult struct {
		target *models.Target
		err    error
	}

	sectorChan := make(chan sectorResult, 1)
	industryChan := make(chan industryResult, 1)
	targetChan := make(chan targetResult, 1)

	// Start parallel lookups
	go func() {
		sector, err := sectorRepo.FindByCodeOrDescription(ctx, sectorCode)
		sectorChan <- sectorResult{sector: sector, err: err}
	}()

	go func() {
		industry, err := industryRepo.FindByCodeOrDescription(ctx, industryCode)
		industryChan <- industryResult{industry: industry, err: err}
	}()

	go func() {
		target, err := targetRepo.FindByCodeOrDescription(ctx, targetCode)
		targetChan <- targetResult{target: target, err: err}
	}()

	// Collect results
	sectorRes := <-sectorChan
	if sectorRes.err != nil {
		logger.Log.WithError(sectorRes.err).Error("ValidateSectorIndustryTarget: Sector lookup failed")
		return "", "", "", "", "", "", utils.NewDatabaseError("Failed to find sector", sectorRes.err)
	}
	if sectorRes.sector == nil {
		logger.Log.WithField("sector_code", sectorCode).Error("ValidateSectorIndustryTarget: Sector not found")
		return "", "", "", "", "", "", utils.NewNotFoundError(fmt.Sprintf("Sector not found: %s", sectorCode), nil)
	}
	sectorID = fmt.Sprintf("%d", sectorRes.sector.SectorID)
	sectorDesc = sectorRes.sector.Description
	logger.Log.WithField("sector_id", sectorID).Debug("ValidateSectorIndustryTarget: Sector found")

	industryRes := <-industryChan
	if industryRes.err != nil {
		logger.Log.WithError(industryRes.err).Error("ValidateSectorIndustryTarget: Industry lookup failed")
		return "", "", "", "", "", "", utils.NewDatabaseError("Failed to find industry", industryRes.err)
	}
	if industryRes.industry == nil {
		logger.Log.WithField("industry_code", industryCode).Error("ValidateSectorIndustryTarget: Industry not found")
		return "", "", "", "", "", "", utils.NewNotFoundError(fmt.Sprintf("Industry not found: %s", industryCode), nil)
	}
	industryID = fmt.Sprintf("%d", industryRes.industry.IndustryID)
	industryDesc = industryRes.industry.Description
	logger.Log.WithField("industry_id", industryID).Debug("ValidateSectorIndustryTarget: Industry found")

	// Validate that industry belongs to sector
	if sectorRes.sector.SectorID != industryRes.industry.SectorID {
		logger.Log.WithFields(map[string]interface{}{
			"sector_id":   sectorID,
			"industry_id": industryID,
		}).Error("ValidateSectorIndustryTarget: Industry does not belong to sector")
		return "", "", "", "", "", "", utils.NewValidationError(fmt.Sprintf("Industry %s does not belong to sector %s", industryCode, sectorCode), nil)
	}

	targetRes := <-targetChan
	if targetRes.err != nil {
		logger.Log.WithError(targetRes.err).Error("ValidateSectorIndustryTarget: Target lookup failed")
		return "", "", "", "", "", "", utils.NewDatabaseError("Failed to find target", targetRes.err)
	}
	if targetRes.target == nil {
		logger.Log.WithField("target_code", targetCode).Error("ValidateSectorIndustryTarget: Target not found")
		return "", "", "", "", "", "", utils.NewNotFoundError(fmt.Sprintf("Target not found: %s", targetCode), nil)
	}
	targetID = fmt.Sprintf("%d", targetRes.target.TargetID)
	if targetRes.target.Description != nil && *targetRes.target.Description != "" {
		targetDesc = *targetRes.target.Description
	} else {
		targetDesc = targetRes.target.ShortName
	}
	logger.Log.WithField("target_id", targetID).Debug("ValidateSectorIndustryTarget: Target found")
	logger.Log.Debug("ValidateSectorIndustryTarget: Validation completed successfully")

	return sectorID, industryID, targetID, sectorDesc, industryDesc, targetDesc, nil
}
