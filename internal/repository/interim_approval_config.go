package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type InterimApprovalConfigRepository interface {
	GetByCustomerType(ctx context.Context, db *sqlx.DB, customerType models.CustomerType) (*models.InterimApprovalConfig, error)
}

type interimApprovalConfigRepository struct {
	db *sqlx.DB
}

func NewInterimApprovalConfigRepository(db *sqlx.DB) InterimApprovalConfigRepository {
	return &interimApprovalConfigRepository{db: db}
}

func (r *interimApprovalConfigRepository) GetByCustomerType(ctx context.Context, db *sqlx.DB, customerType models.CustomerType) (*models.InterimApprovalConfig, error) {
	// Optimized query - use composite index-friendly WHERE clause
	// The composite index (customerType, gracePeriod, status, createdAt DESC) should make this fast
	// Only select fields we actually need - skip JSONB fields that cause unmarshal issues
	query := `
		SELECT 
			"interimApprovalConfigId", "status", "customerType", "gracePeriod",
			"gracePeriodBeforeAction", "graceDurationBeforeAction", "gracePeriodExpiryAction",
			"notifyCustomer", "notifyCustomerPeriodBeforeAction", "notifyCustomerDurationBeforeAction",
			"notifyCustomerChannels", "notifyCustomerMessageTemplate", "notifyRelationshipTeam",
			"notifyRelationshipTeamPeriodBeforeAction", "notifyRelationshipTeamDurationBeforeAction",
			"notifyRelationshipTeamChannels", "notifyRelationshipTeamMessageTemplate",
			"initiator", "initiatorId", "approver", "approverId", "createdAt", "updatedAt",
			'{}'::jsonb as "depositProducts",
			'{}'::jsonb as "paymentProducts",
			'{}'::jsonb as "relationshipTeam"
		FROM interim_approval_config
		WHERE "customerType" = $1 AND "gracePeriod" = true AND "status" = $2
		ORDER BY "createdAt" DESC
		LIMIT 1
	`

	var config models.InterimApprovalConfig
	err := db.GetContext(ctx, &config, query, customerType, models.ConfigStatusApproved)
	if err != nil {
		// Don't log "record not found" as error - it's expected if no config exists
		return nil, err
	}

	return &config, nil
}
