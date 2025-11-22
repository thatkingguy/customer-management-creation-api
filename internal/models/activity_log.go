package models

import (
	"time"

	"github.com/google/uuid"
)

// ActivityLog represents the activity_log table
type ActivityLog struct {
	ActivityLogID uuid.UUID  `db:"activityLogId" json:"activityLogId"`
	CustomerID    *uuid.UUID `db:"customerId" json:"customerId"`
	RequestID     *uuid.UUID `db:"requestId" json:"requestId"`
	Description   string     `db:"description" json:"description"`
	Reason        JSONB      `db:"reason" json:"reason"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time  `db:"updatedAt" json:"updatedAt"`
}
