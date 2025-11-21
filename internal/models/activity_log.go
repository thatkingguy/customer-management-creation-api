package models

import (
	"time"

	"github.com/google/uuid"
)

// ActivityLog represents the activity_log table
type ActivityLog struct {
	ActivityLogID uuid.UUID `db:"activityLogId" json:"activityLogId"`
	Description   string    `db:"description" json:"description"`
	CustomerID    uuid.UUID `db:"customerId" json:"customerId"`
	CreatedAt     time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time `db:"updatedAt" json:"updatedAt"`
}
