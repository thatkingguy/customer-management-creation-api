package models

import (
	"time"

	"github.com/google/uuid"
)

// ConfigStatus represents the status of an interim approval config
type ConfigStatus string

const (
	ConfigStatusApproved ConfigStatus = "Approved"
	ConfigStatusPending  ConfigStatus = "Pending"
	ConfigStatusRejected ConfigStatus = "Rejected"
)

// DurationUnit represents the unit for duration
type DurationUnit string

const (
	DurationUnitDays   DurationUnit = "Days"
	DurationUnitHours  DurationUnit = "Hours"
	DurationUnitMonths DurationUnit = "Months"
	DurationUnitWeeks  DurationUnit = "Weeks"
)

// GracePeriodExpiryAction represents the action to take when grace period expires
type GracePeriodExpiryAction string

const (
	GracePeriodExpiryActionDeactivateCustomer GracePeriodExpiryAction = "Deactivate Customer"
	GracePeriodExpiryActionPostNoDebit        GracePeriodExpiryAction = "Post No Debit"
)

// InterimApprovalConfig represents the interim_approval_config table
type InterimApprovalConfig struct {
	InterimApprovalConfigID                    uuid.UUID                `db:"interimApprovalConfigId" json:"interimApprovalConfigId"`
	Status                                     ConfigStatus             `db:"status" json:"status"`
	CustomerType                               CustomerType             `db:"customerType" json:"customerType"`
	GracePeriod                                bool                     `db:"gracePeriod" json:"gracePeriod"`
	GracePeriodBeforeAction                    *int                     `db:"gracePeriodBeforeAction" json:"gracePeriodBeforeAction"`
	GraceDurationBeforeAction                  *DurationUnit            `db:"graceDurationBeforeAction" json:"graceDurationBeforeAction"`
	GracePeriodExpiryAction                    *GracePeriodExpiryAction `db:"gracePeriodExpiryAction" json:"gracePeriodExpiryAction"`
	NotifyCustomer                             bool                     `db:"notifyCustomer" json:"notifyCustomer"`
	NotifyCustomerPeriodBeforeAction           *int                     `db:"notifyCustomerPeriodBeforeAction" json:"notifyCustomerPeriodBeforeAction"`
	NotifyCustomerDurationBeforeAction         *DurationUnit            `db:"notifyCustomerDurationBeforeAction" json:"notifyCustomerDurationBeforeAction"`
	NotifyCustomerChannels                     *string                  `db:"notifyCustomerChannels" json:"notifyCustomerChannels"`
	NotifyCustomerMessageTemplate              *string                  `db:"notifyCustomerMessageTemplate" json:"notifyCustomerMessageTemplate"`
	NotifyRelationshipTeam                     bool                     `db:"notifyRelationshipTeam" json:"notifyRelationshipTeam"`
	NotifyRelationshipTeamPeriodBeforeAction   *int                     `db:"notifyRelationshipTeamPeriodBeforeAction" json:"notifyRelationshipTeamPeriodBeforeAction"`
	NotifyRelationshipTeamDurationBeforeAction *string                  `db:"notifyRelationshipTeamDurationBeforeAction" json:"notifyRelationshipTeamDurationBeforeAction"`
	NotifyRelationshipTeamChannels             *string                  `db:"notifyRelationshipTeamChannels" json:"notifyRelationshipTeamChannels"`
	NotifyRelationshipTeamMessageTemplate      *string                  `db:"notifyRelationshipTeamMessageTemplate" json:"notifyRelationshipTeamMessageTemplate"`
	DepositProducts                            JSONB                    `db:"depositProducts" json:"depositProducts"`
	PaymentProducts                            JSONB                    `db:"paymentProducts" json:"paymentProducts"`
	Initiator                                  string                   `db:"initiator" json:"initiator"`
	InitiatorID                                uuid.UUID                `db:"initiatorId" json:"initiatorId"`
	Approver                                   *string                  `db:"approver" json:"approver"`
	ApproverID                                 *uuid.UUID               `db:"approverId" json:"approverId"`
	RelationshipTeam                           JSONB                    `db:"relationshipTeam" json:"relationshipTeam"`
	CreatedAt                                  time.Time                `db:"createdAt" json:"createdAt"`
	UpdatedAt                                  time.Time                `db:"updatedAt" json:"updatedAt"`
}
