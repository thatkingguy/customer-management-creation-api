package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/dto"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

// mapRequestType maps string request type to RequestType enum
func mapRequestType(requestType string) models.RequestType {
	switch strings.ToLower(strings.TrimSpace(requestType)) {
	case "creation":
		return models.RequestTypeCreation
	case "deactivation":
		return models.RequestTypeDeactivation
	case "reactivation":
		return models.RequestTypeReactivation
	case "modification":
		return models.RequestTypeModification
	case "contractmodification":
		return models.RequestTypeContractModification
	default:
		return models.RequestTypeCreation
	}
}

// mapCreationMode maps form type string to CreationMode enum
func mapCreationMode(formType string) models.CreationMode {
	switch strings.ToLower(strings.TrimSpace(formType)) {
	case "accelerated":
		return models.CreationModeAccelerated
	case "bulk":
		return models.CreationModeBulk
	case "legacy":
		return models.CreationModeLegacy
	default:
		return models.CreationModeLegacy
	}
}

// mapCustomerType maps customer type string to CustomerType enum
func mapCustomerType(customerType string) *models.CustomerType {
	if strings.ToLower(customerType) == "sme" {
		ct := models.CustomerTypeSME
		return &ct
	}
	ct := models.CustomerTypeIndividual
	return &ct
}

// buildRequestTitle builds a request title from customer data
func buildRequestTitle(payload *dto.CreateDraftRequest, requestType models.RequestType) string {
	var firstName, surname string

	// Extract customer name from customerData
	for _, section := range payload.Data.CustomerData {
		if section.SectionName == "bioData" {
			if dataMap, ok := section.Data.(map[string]interface{}); ok {
				if val, ok := dataMap["firstName"].(string); ok {
					firstName = val
				}
				if val, ok := dataMap["surname"].(string); ok {
					surname = val
				}
			}
		}
	}

	if firstName != "" || surname != "" {
		name := strings.TrimSpace(fmt.Sprintf("%s %s", firstName, surname))
		return fmt.Sprintf("%s of %s", requestType, name)
	}

	return fmt.Sprintf("%s request", requestType)
}

// parseCustomerData parses customer data sections into a flat map
func parseCustomerData(customerData []dto.CustomerDataSection) map[string]interface{} {
	result := make(map[string]interface{})

	for _, section := range customerData {
		if dataMap, ok := section.Data.(map[string]interface{}); ok {
			for key, value := range dataMap {
				// Handle different field name variations
				normalizedKey := normalizeFieldName(key)
				result[normalizedKey] = value
			}
		}
	}

	return result
}

// normalizeFieldName normalizes field names to standard format
func normalizeFieldName(key string) string {
	keyLower := strings.ToLower(key)

	// Field name mapping
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

	if normalized, ok := fieldMap[keyLower]; ok {
		return normalized
	}

	// Preserve original key if no mapping found
	return key
}

// getString extracts string value from map with multiple possible keys
func getString(m map[string]interface{}, keys ...string) *string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				return &str
			}
		}
		// Try case-insensitive match
		keyLower := strings.ToLower(key)
		for k, v := range m {
			if strings.ToLower(k) == keyLower {
				if str, ok := v.(string); ok && str != "" {
					return &str
				}
			}
		}
	}
	return nil
}

// getField extracts field value from map (alias for getString for consistency)
func getField(m map[string]interface{}, keys ...string) *string {
	return getString(m, keys...)
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// calculateDate calculates a date by adding period to startDate based on duration unit
func calculateDate(startDate time.Time, period *int, duration *models.DurationUnit) string {
	if period == nil || duration == nil {
		return ""
	}

	var result time.Time
	switch *duration {
	case models.DurationUnitDays:
		result = startDate.AddDate(0, 0, *period)
	case models.DurationUnitWeeks:
		result = startDate.AddDate(0, 0, *period*7)
	case models.DurationUnitMonths:
		result = startDate.AddDate(0, *period, 0)
	case models.DurationUnitHours:
		result = startDate.Add(time.Duration(*period) * time.Hour)
	default:
		result = startDate
	}

	return result.Format("2006-01-02 15:04:05")
}

// parseCustomerInfo extracts customer information from parsed customer data
type ParsedCustomerInfo struct {
	FirstName    *string
	Surname      *string
	OtherNames   *string
	Title        *string
	DateOfBirth  *string
	MobileNumber *string
	EmailAddress *string
	BVN          *string
	NIN          *string
	Nationality  *string
}

func parseCustomerInfo(customerProfileObj map[string]interface{}) *ParsedCustomerInfo {
	return &ParsedCustomerInfo{
		FirstName:    getString(customerProfileObj, "firstName"),
		Surname:      getString(customerProfileObj, "surname"),
		OtherNames:   getString(customerProfileObj, "otherNames"),
		Title:        getString(customerProfileObj, "title"),
		DateOfBirth:  getString(customerProfileObj, "dateOfBirth"),
		MobileNumber: getString(customerProfileObj, "mobileNumber", "phoneNumber"),
		EmailAddress: getString(customerProfileObj, "emailAddress", "email"),
		BVN:          getString(customerProfileObj, "bvn", "bVN"),
		NIN:          getString(customerProfileObj, "nin"),
		Nationality:  getString(customerProfileObj, "nationality"),
	}
}

// buildCustomerProfileDataJSON builds the customerProfileData JSONB field
func buildCustomerProfileDataJSON(
	customerProfileObj map[string]interface{},
	customerEntityID string,
	customerNumber string,
) models.JSONB {
	data := make(models.JSONB)

	// Copy all fields from customerProfileObj
	for k, v := range customerProfileObj {
		data[k] = v
	}

	// Add computed fields
	data["customerEntityId"] = customerEntityID
	data["customerNumber"] = customerNumber

	// Build full name
	firstName := getString(customerProfileObj, "firstName")
	surname := getString(customerProfileObj, "surname")
	if firstName != nil && surname != nil {
		data["fullName"] = fmt.Sprintf("%s %s", *firstName, *surname)
	} else if firstName != nil {
		data["fullName"] = *firstName
	} else if surname != nil {
		data["fullName"] = *surname
	}

	// Normalize field names
	if bvn := getString(customerProfileObj, "bvn", "bVN"); bvn != nil {
		data["bvn"] = *bvn
	}
	if idType := getString(customerProfileObj, "chooseAnId", "chooseAnID"); idType != nil {
		data["idType"] = *idType
		data["chooseAnId"] = *idType
	}
	if idNumber := getString(customerProfileObj, "idNumber", "iDNumber", "ID Number"); idNumber != nil {
		data["idNumber"] = *idNumber
	}
	if email := getString(customerProfileObj, "email", "emailAddress"); email != nil {
		data["email"] = *email
		data["emailAddress"] = *email
	}

	return data
}
