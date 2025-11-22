package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

func (s *requestService) createRiskAssessments(ctx context.Context, tx *sql.Tx, customerID uuid.UUID, riskAssessmentData []interface{}) error {
	if len(riskAssessmentData) == 0 {
		return nil
	}

	var assessments []*models.RiskAssessment

	for _, sectionItem := range riskAssessmentData {
		sectionMap, ok := sectionItem.(map[string]interface{})
		if !ok {
			continue
		}

		sectionName, _ := sectionMap["sectionName"].(string)
		dataItems, ok := sectionMap["data"].([]interface{})
		if !ok {
			continue
		}

		for _, item := range dataItems {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			assessment := &models.RiskAssessment{
				RiskAssessmentID:       s.idGen.GenerateUUID(),
				CustomerID:             &customerID,
				SectionName:             stringPtr(sectionName),
				Parameter:               getStringValueFromMap(itemMap, "parameter"),
				ImpliedWeight:           getIntValueFromMap(itemMap, "impliedWeight"),
				ParameterOption:         getStringValueFromMap(itemMap, "parameterOption"),
				AssessmentType:          getStringPtrFromMap(itemMap, "assessmentType"),
				EscalationFactor:        getIntPtrFromMap(itemMap, "escalationFactor"),
				OptionsWeightAllocation: getIntValueFromMap(itemMap, "optionsWeightAllocation"),
				Score:                   getIntValueFromMap(itemMap, "score"),
				TypeOfPoliticalExposure: getStringPtrFromMap(itemMap, "typeOfPoliticalExposure"),
				TypeOfPoliticalExposureDesc: getStringPtrFromMap(itemMap, "typeOfPoliticalExposureDesc"),
				Data:                   models.JSONB(itemMap),
				CreatedAt:               time.Now(),
				UpdatedAt:               time.Now(),
			}

			assessments = append(assessments, assessment)
		}
	}

	if len(assessments) > 0 {
		return s.riskAssessmentRepo.CreateBatch(ctx, tx, assessments)
	}

	return nil
}

func (s *requestService) createAccounts(ctx context.Context, tx *sql.Tx, customerID uuid.UUID, customerProfileID uuid.UUID, accountData []interface{}) error {
	if len(accountData) == 0 {
		return nil
	}

	var accounts []*models.OtherAccount

	for _, sectionItem := range accountData {
		sectionMap, ok := sectionItem.(map[string]interface{})
		if !ok {
			continue
		}

		dataItems, ok := sectionMap["data"].([]interface{})
		if !ok {
			continue
		}

		for _, item := range dataItems {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			account := &models.OtherAccount{
				OtherAccountID:    s.idGen.GenerateUUID(),
				CustomerID:        customerID,
				CustomerProfileID: customerProfileID,
				BankName:          getStringPtrFromMap(itemMap, "bankName"),
				AccountNumber:     getStringPtrFromMap(itemMap, "accountNumber"),
				AccountType:       getStringPtrFromMap(itemMap, "accountType"),
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}

			accounts = append(accounts, account)
		}
	}

	if len(accounts) > 0 {
		return s.otherAccountRepo.CreateBatch(ctx, tx, accounts)
	}

	return nil
}

func (s *requestService) createSignatories(ctx context.Context, tx *sql.Tx, customerID uuid.UUID, requestID uuid.UUID, signatoryData []interface{}) error {
	if len(signatoryData) == 0 {
		return nil
	}

	var signatories []*models.Signatory

	for _, sectionItem := range signatoryData {
		sectionMap, ok := sectionItem.(map[string]interface{})
		if !ok {
			continue
		}

		dataItems, ok := sectionMap["data"].([]interface{})
		if !ok {
			continue
		}

		for _, item := range dataItems {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			signatory := &models.Signatory{
				SignatoryID:     s.idGen.GenerateUUID(),
				CustomerID:      &customerID,
				Title:           getStringPtrFromMap(itemMap, "title"),
				FirstName:       getStringPtrFromMap(itemMap, "firstName"),
				Surname:         getStringPtrFromMap(itemMap, "surname"),
				OtherNames:      getStringPtrFromMap(itemMap, "otherNames"),
				DateOfBirth:     getStringPtrFromMap(itemMap, "dateOfBirth"),
				MobileNumber:    getStringPtrFromMap(itemMap, "mobileNumber"),
				EmailAddress:    getStringPtrFromMap(itemMap, "emailAddress"),
				PrimarySignatory: getBoolValueFromMap(itemMap, "primarySignatory"),
				BVN:             getStringPtrFromMap(itemMap, "bvn"),
				NIN:             getStringPtrFromMap(itemMap, "nin"),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			signatories = append(signatories, signatory)
		}
	}

	if len(signatories) > 0 {
		return s.signatoryRepo.CreateBatch(ctx, tx, signatories)
	}

	return nil
}

func (s *requestService) createReferences(ctx context.Context, tx *sql.Tx, customerID uuid.UUID, referenceData []interface{}) error {
	if len(referenceData) == 0 {
		return nil
	}

	var references []*models.Reference

	for _, sectionItem := range referenceData {
		sectionMap, ok := sectionItem.(map[string]interface{})
		if !ok {
			continue
		}

		dataItems, ok := sectionMap["data"].([]interface{})
		if !ok {
			continue
		}

		for _, item := range dataItems {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			reference := &models.Reference{
				ReferenceID:  s.idGen.GenerateUUID(),
				CustomerID:   customerID,
				Name:         getStringPtrFromMap(itemMap, "name"),
				PhoneNumber:  getStringPtrFromMap(itemMap, "phoneNumber"),
				Email:        getStringPtrFromMap(itemMap, "email"),
				Address:      getStringPtrFromMap(itemMap, "address"),
				Relationship: getStringPtrFromMap(itemMap, "relationship"),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			references = append(references, reference)
		}
	}

	if len(references) > 0 {
		return s.referenceRepo.CreateBatch(ctx, tx, references)
	}

	return nil
}

func (s *requestService) createGracePeriodData(ctx context.Context, tx *sql.Tx, customerID uuid.UUID, interimConfig *models.InterimApprovalConfig) error {
	if interimConfig == nil {
		return nil
	}

	customerGracePeriodID := s.idGen.GenerateUUID()
	requestDate := time.Now()

	// Calculate grace period date
	gracePeriodDate := calculateDate(
		requestDate,
		interimConfig.GracePeriodBeforeAction,
		interimConfig.GraceDurationBeforeAction,
	)

	// Calculate customer notification date
	var notifyCustomerDate *string
	if interimConfig.NotifyCustomer {
		date := calculateDate(
			requestDate,
			interimConfig.NotifyCustomerPeriodBeforeAction,
			interimConfig.NotifyCustomerDurationBeforeAction,
		)
		notifyCustomerDate = &date
	}

	// Calculate relationship team notification date
	var notifyRelationshipTeamDate *string
	if interimConfig.NotifyRelationshipTeam {
		// Note: NotifyRelationshipTeamDurationBeforeAction is a string, not DurationUnit
		// This would need proper parsing/conversion
		// For now, we'll use Days as default
		duration := models.DurationUnitDays
		date := calculateDate(
			requestDate,
			interimConfig.NotifyRelationshipTeamPeriodBeforeAction,
			&duration,
		)
		notifyRelationshipTeamDate = &date
	}

	gracePeriod := &models.CustomerGracePeriod{
		CustomerGracePeriodID:      customerGracePeriodID,
		InterimApprovalConfigID:     &interimConfig.InterimApprovalConfigID,
		CustomerID:                  &customerID,
		GracePeriod:                 interimConfig.GracePeriod,
		GracePeriodDate:             &gracePeriodDate,
		NotifyCustomer:              interimConfig.NotifyCustomer,
		NotifyCustomerDate:          notifyCustomerDate,
		NotifyCustomerChannels:      interimConfig.NotifyCustomerChannels,
		NotifyRelationshipTeam:      interimConfig.NotifyRelationshipTeam,
		NotifyRelationshipTeamDate:   notifyRelationshipTeamDate,
		RelationshipTeam:            interimConfig.RelationshipTeam,
		InterimApprovalConfigData:   models.JSONB{}, // Store config as JSONB
		CreatedAt:                   requestDate,
		UpdatedAt:                    requestDate,
	}

	// Convert config to JSONB
	if configJSON, err := convertToJSONB(interimConfig); err == nil {
		gracePeriod.InterimApprovalConfigData = configJSON
	}

	return s.gracePeriodRepo.Create(ctx, tx, gracePeriod)
}

// Helper functions for map value extraction
func getStringValueFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringPtrFromMap(m map[string]interface{}, key string) *string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok && str != "" {
			return &str
		}
	}
	return nil
}

func getIntValueFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

func getIntPtrFromMap(m map[string]interface{}, key string) *int {
	if val, ok := m[key]; ok {
		var i int
		switch v := val.(type) {
		case int:
			i = v
		case float64:
			i = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				i = parsed
			} else {
				return nil
			}
		default:
			return nil
		}
		return &i
	}
	return nil
}

func getBoolValueFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func convertToJSONB(v interface{}) (models.JSONB, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var result models.JSONB
	err = json.Unmarshal(jsonBytes, &result)
	return result, err
}

