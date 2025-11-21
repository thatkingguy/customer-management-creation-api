package models

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents the customer table
type Customer struct {
	CustomerID         uuid.UUID  `db:"customerId" json:"customerId"`
	CustomerType       string     `db:"customerType" json:"customerType"`
	Status             string     `db:"status" json:"status"`
	Monitoring         bool       `db:"monitoring" json:"monitoring"`
	ApprovalStatus     string     `db:"approvalStatus" json:"approvalStatus"`
	TenantID           *uuid.UUID `db:"tenantId" json:"tenantId"`
	Initiator          string     `db:"initiator" json:"initiator"`
	InitiatorID        uuid.UUID  `db:"initiatorId" json:"initiatorId"`
	Approver           string     `db:"approver" json:"approver"`
	ApproverID         uuid.UUID  `db:"approverId" json:"approverId"`
	Branch             string     `db:"branch" json:"branch"`
	ApproverBranch     string     `db:"approverBranch" json:"approverBranch"`
	Department         string     `db:"department" json:"department"`
	CategoryOfCustomer *string    `db:"categoryOfCustomer" json:"categoryOfCustomer"`
	CustomerSubType    *int       `db:"customerSubType" json:"customerSubType"`
	RelatedEntity      *uuid.UUID `db:"relatedEntity" json:"relatedEntity"`
	CreatedAt          time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time  `db:"updatedAt" json:"updatedAt"`
}
