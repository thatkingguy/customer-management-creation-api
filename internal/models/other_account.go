package models

import (
	"time"

	"github.com/google/uuid"
)

// OtherAccount represents the other_account table
type OtherAccount struct {
	OtherAccountID    uuid.UUID `db:"otherAccountId" json:"otherAccountId"`
	CustomerID        uuid.UUID `db:"customerId" json:"customerId"`
	CustomerProfileID uuid.UUID `db:"customerProfileId" json:"customerProfileId"`
	BankName          *string   `db:"bankName" json:"bankName"`
	AccountNumber     *string   `db:"accountNumber" json:"accountNumber"`
	AccountType       *string   `db:"accountType" json:"accountType"`
	CreatedAt         time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time `db:"updatedAt" json:"updatedAt"`
}

