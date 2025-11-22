package models

import (
	"time"

	"github.com/google/uuid"
)

// RequestStatus represents the status of a request
type RequestStatus string

const (
	RequestStatusApproved        RequestStatus = "Approved"
	RequestStatusDraft           RequestStatus = "Draft"
	RequestStatusInIssue         RequestStatus = "In Issue"
	RequestStatusInReview        RequestStatus = "In-Review"
	RequestStatusInterimApproval RequestStatus = "Interim Approval"
	RequestStatusPending         RequestStatus = "Pending"
	RequestStatusRejected        RequestStatus = "Rejected"
	RequestStatusProcessing      RequestStatus = "Processing"
	RequestStatusFailed          RequestStatus = "Failed"
)

// ApprovalStatus represents the approval status
type ApprovalStatus string

const (
	ApprovalStatusApproved ApprovalStatus = "Approved"
	ApprovalStatusPending  ApprovalStatus = "Pending"
	ApprovalStatusRejected ApprovalStatus = "Rejected"
	ApprovalStatusFailed   ApprovalStatus = "Failed"
)

// RequestType represents the type of request
type RequestType string

const (
	RequestTypeCreation            RequestType = "Creation"
	RequestTypeDeactivation        RequestType = "Deactivation"
	RequestTypeModification        RequestType = "Modification"
	RequestTypeReactivation        RequestType = "Reactivation"
	RequestTypeContractModification RequestType = "ContractModification"
)

// CustomerType represents the customer type
type CustomerType string

const (
	CustomerTypeIndividual CustomerType = "Individual"
	CustomerTypeSME        CustomerType = "SME"
)

// CreationMode represents the creation mode
type CreationMode string

const (
	CreationModeAccelerated CreationMode = "Accelerated"
	CreationModeBulk         CreationMode = "Bulk"
	CreationModeLegacy       CreationMode = "Legacy"
)

// Request represents the request table
type Request struct {
	RequestID              uuid.UUID   `db:"requestId" json:"requestId"`
	CustomerID             *uuid.UUID  `db:"customerId" json:"customerId"`
	RequestTitle           string      `db:"requestTitle" json:"requestTitle"`
	RequestType            RequestType `db:"requestType" json:"requestType"`
	RequestSubType         *string     `db:"requestSubType" json:"requestSubType"`
	AccountNumber          *string     `db:"accountNumber" json:"accountNumber"`
	Justification          *string     `db:"justification" json:"justification"`
	Initiator              string      `db:"initiator" json:"initiator"`
	InitiatorID            uuid.UUID   `db:"initiatorId" json:"initiatorId"`
	Status                 RequestStatus `db:"status" json:"status"`
	ApprovalStatus         ApprovalStatus `db:"approvalStatus" json:"approvalStatus"`
	Approver               *string     `db:"approver" json:"approver"`
	ApproverID             *uuid.UUID  `db:"approverId" json:"approverId"`
	Data                   JSONB       `db:"data" json:"data"`
	CustomerType           *CustomerType `db:"customerType" json:"customerType"`
	CreationMode           *CreationMode  `db:"creationMode" json:"creationMode"`
	Branch                 *string     `db:"branch" json:"branch"`
	ApproverBranch         *string     `db:"approverBranch" json:"approverBranch"`
	Department             *string     `db:"department" json:"department"`
	Withdrawn              bool        `db:"withdrawn" json:"withdrawn"`
	IsDeleted              bool        `db:"isDeleted" json:"isDeleted"`
	IsProduct              bool        `db:"isProduct" json:"isProduct"`
	DeletedOn              *time.Time  `db:"deletedOn" json:"deletedOn"`
	RejectionDocument      JSONB       `db:"rejectionDocument" json:"rejectionDocument"`
	RejectionReason        *string     `db:"rejectionReason" json:"rejectionReason"`
	HasCollectionProduct   bool        `db:"hasCollectionProduct" json:"hasCollectionProduct"`
	BulkReferenceID       *string     `db:"bulkReferenceId" json:"bulkReferenceId"`
	CreatedAt              time.Time   `db:"createdAt" json:"createdAt"`
	UpdatedAt              time.Time   `db:"updatedAt" json:"updatedAt"`
}

