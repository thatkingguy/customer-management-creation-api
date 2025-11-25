package models

import (
	"time"

	"github.com/google/uuid"
)

// CustomerGracePeriod represents the customer_grace_period table
type CustomerGracePeriod struct {
	CustomerGracePeriodID      uuid.UUID `db:"customerGracePeriodId" json:"customerGracePeriodId"`
	InterimApprovalConfigID    *uuid.UUID `db:"interimApprovalConfigId" json:"interimApprovalConfigId"`
	CustomerID                 *uuid.UUID `db:"customerId" json:"customerId"`
	GracePeriod                bool       `db:"gracePeriod" json:"gracePeriod"`
	GracePeriodDate            *string    `db:"gracePeriodDate" json:"gracePeriodDate"`
	NotifyCustomer             bool       `db:"notifyCustomer" json:"notifyCustomer"`
	NotifyCustomerDate         *string    `db:"notifyCustomerDate" json:"notifyCustomerDate"`
	NotifyCustomerChannels     *string    `db:"notifyCustomerChannels" json:"notifyCustomerChannels"`
	NotifyRelationshipTeam     bool       `db:"notifyRelationshipTeam" json:"notifyRelationshipTeam"`
	NotifyRelationshipTeamDate *string    `db:"notifyRelationshipTeamDate" json:"notifyRelationshipTeamDate"`
	RelationshipTeam           JSONB      `db:"relationshipTeam" json:"relationshipTeam"`
	InterimApprovalConfigData  JSONB      `db:"interimApprovalConfigData" json:"interimApprovalConfigData"`
	CreatedAt                  time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt                  time.Time  `db:"updatedAt" json:"updatedAt"`
}




