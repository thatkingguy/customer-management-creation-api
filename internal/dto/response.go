package dto

import "github.com/google/uuid"

// DirectCreateCustomerResponse represents the success response
type DirectCreateCustomerResponse struct {
	CustomerID uuid.UUID `json:"customerId"`
	RequestID  uuid.UUID `json:"requestId"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
}

// CreateDraftResponse represents the response for draft creation
type CreateDraftResponse struct {
	RequestID uuid.UUID `json:"requestId"`
}

// ApproveRequestResponse represents the response for request approval
type ApproveRequestResponse struct {
	CustomerID uuid.UUID  `json:"customerId"`
	RequestID  uuid.UUID  `json:"requestId"`
	SignatoryID *uuid.UUID `json:"signatoryId"`
	Product     *string    `json:"product"`
}

