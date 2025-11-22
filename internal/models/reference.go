package models

import (
	"time"

	"github.com/google/uuid"
)

// Reference represents the reference table
type Reference struct {
	ReferenceID  uuid.UUID `db:"referenceId" json:"referenceId"`
	CustomerID   uuid.UUID `db:"customerId" json:"customerId"`
	Name         *string   `db:"name" json:"name"`
	PhoneNumber  *string   `db:"phoneNumber" json:"phoneNumber"`
	Email        *string   `db:"email" json:"email"`
	Address      *string   `db:"address" json:"address"`
	Relationship *string   `db:"relationship" json:"relationship"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time `db:"updatedAt" json:"updatedAt"`
}

