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
