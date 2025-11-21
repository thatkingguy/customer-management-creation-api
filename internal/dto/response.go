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
