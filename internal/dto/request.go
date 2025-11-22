package dto

import "github.com/google/uuid"

// DirectCreateCustomerRequest represents the request payload
type DirectCreateCustomerRequest struct {
	Data RequestData `json:"data" binding:"required"`
}

type RequestData struct {
	Branch        string                `json:"branch" binding:"required"`
	CustomerType  string                `json:"customerType" binding:"required"`
	CustomerData  []CustomerDataSection `json:"customerData" binding:"required,min=1"`
	FormInfo      FormInformation       `json:"formInformation" binding:"required"`
	RequestInfo   RequestInformation    `json:"requestData" binding:"required"`
	TenantID      *uuid.UUID            `json:"tenantId"`
	Department    string                `json:"department"`
	RelatedEntity *uuid.UUID            `json:"relatedEntity"`
	FlagCustomer  *FlagCustomer         `json:"flagCustomer"`
}

type CustomerDataSection struct {
	SectionName string      `json:"sectionName" binding:"required"`
	Data        interface{} `json:"data" binding:"required"`
}

type FormInformation struct {
	FormType string `json:"formType" binding:"required"`
}

type RequestInformation struct {
	RequestType string `json:"requestType" binding:"required"`
}

type FlagCustomer struct {
	Status bool `json:"status"`
}

// CreateDraftRequest represents the request payload for creating a draft
type CreateDraftRequest struct {
	CustomerType string           `json:"customerType" binding:"required"`
	Data         DraftRequestData `json:"data" binding:"required"`
}

// DraftRequestData represents the data section of draft request
type DraftRequestData struct {
	Branch             string                  `json:"branch" binding:"required"`
	CustomerData       []CustomerDataSection   `json:"customerData" binding:"required,min=1"`
	FormInformation    FormInformation         `json:"formInformation" binding:"required"`
	RequestData        RequestInformation      `json:"requestData" binding:"required"`
	WaiverData         []interface{}           `json:"waiverData"`
	SignatoryData      []SignatorySection      `json:"signatoryData"`
	ExecutiveData      []ExecutiveSection      `json:"executiveData"`
	AccountData        []AccountSection        `json:"accountData"`
	RiskAssessmentData []RiskAssessmentSection `json:"riskAssessmentData"`
	RiskData           *RiskData               `json:"riskData"`
	ReferenceData      []ReferenceSection      `json:"referenceData"`
}

// SignatorySection represents a signatory data section
type SignatorySection struct {
	SectionName string        `json:"sectionName"`
	Data        []interface{} `json:"data"`
}

// ExecutiveSection represents executive data section
type ExecutiveSection struct {
	Data []interface{} `json:"data"`
}

// AccountSection represents account data section
type AccountSection struct {
	Data []AccountItem `json:"data"`
}

// AccountItem represents a single account item
type AccountItem struct {
	BankName      *string `json:"bankName"`
	AccountNumber *string `json:"accountNumber"`
	AccountType   *string `json:"accountType"`
}

// RiskAssessmentSection represents a risk assessment section
type RiskAssessmentSection struct {
	SectionName string               `json:"sectionName"`
	Data        []RiskAssessmentItem `json:"data"`
}

// RiskAssessmentItem represents a single risk assessment item
type RiskAssessmentItem struct {
	Parameter                   string  `json:"parameter"`
	ImpliedWeight               int     `json:"impliedWeight"`
	ParameterOption             string  `json:"parameterOption"`
	AssessmentType              *string `json:"assessmentType"`
	EscalationFactor            *int    `json:"escalationFactor"`
	OptionsWeightAllocation     int     `json:"optionsWeightAllocation"`
	Score                       int     `json:"score"`
	TypeOfPoliticalExposure     *string `json:"typeOfPoliticalExposure"`
	TypeOfPoliticalExposureDesc *string `json:"typeOfPoliticalExposureDesc"`
}

// RiskData represents risk data
type RiskData struct {
	RiskScore  float64 `json:"riskScore"`
	RiskStatus string  `json:"riskStatus"`
}

// ReferenceSection represents reference data section
type ReferenceSection struct {
	SectionName string          `json:"sectionName"`
	Data        []ReferenceItem `json:"data"`
}

// ReferenceItem represents a single reference item
type ReferenceItem struct {
	Name         *string `json:"name"`
	PhoneNumber  *string `json:"phoneNumber"`
	Email        *string `json:"email"`
	Address      *string `json:"address"`
	Relationship *string `json:"relationship"`
}

// ApproveRequestPayload represents the request payload for approving a request
type ApproveRequestPayload struct {
	Status string `json:"status" binding:"required"` // "Approved", "Interim Approval", or "Rejected"
}
