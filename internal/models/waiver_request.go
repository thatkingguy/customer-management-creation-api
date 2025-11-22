package models

import (
	"time"

	"github.com/google/uuid"
)

// WaiverRequest represents the waiver_request table
type WaiverRequest struct {
	WaiverRequestID uuid.UUID  `db:"waiverRequestId" json:"waiverRequestId"`
	RequestID        *uuid.UUID `db:"requestId" json:"requestId"`
	CustomerID       *uuid.UUID `db:"customerId" json:"customerId"`
	Status           *string    `db:"status" json:"status"`
	Approver         *string    `db:"approver" json:"approver"`
	ApproverID       *uuid.UUID `db:"approverId" json:"approverId"`
	Justification    *string    `db:"justification" json:"justification"`
	CreatedAt        time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time  `db:"updatedAt" json:"updatedAt"`
}

