package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type CustomerGracePeriodRepository interface {
	Create(ctx context.Context, tx *sql.Tx, gracePeriod *models.CustomerGracePeriod) error
}

type customerGracePeriodRepository struct {
	db *sqlx.DB
}

func NewCustomerGracePeriodRepository(db *sqlx.DB) CustomerGracePeriodRepository {
	return &customerGracePeriodRepository{db: db}
}

func (r *customerGracePeriodRepository) Create(ctx context.Context, tx *sql.Tx, gracePeriod *models.CustomerGracePeriod) error {
	query := `
		INSERT INTO customer_grace_period (
			"customerGracePeriodId", "interimApprovalConfigId", "customerId", "gracePeriod",
			"gracePeriodDate", "notifyCustomer", "notifyCustomerDate", "notifyCustomerChannels",
			"notifyRelationshipTeam", "notifyRelationshipTeamDate", "relationshipTeam",
			"interimApprovalConfigData", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	var relationshipTeamJSON interface{}
	if gracePeriod.RelationshipTeam != nil {
		relationshipTeamJSON = gracePeriod.RelationshipTeam
	}

	var configDataJSON interface{}
	if gracePeriod.InterimApprovalConfigData != nil {
		configDataJSON = gracePeriod.InterimApprovalConfigData
	}

	_, err := tx.ExecContext(ctx, query,
		gracePeriod.CustomerGracePeriodID,
		gracePeriod.InterimApprovalConfigID,
		gracePeriod.CustomerID,
		gracePeriod.GracePeriod,
		gracePeriod.GracePeriodDate,
		gracePeriod.NotifyCustomer,
		gracePeriod.NotifyCustomerDate,
		gracePeriod.NotifyCustomerChannels,
		gracePeriod.NotifyRelationshipTeam,
		gracePeriod.NotifyRelationshipTeamDate,
		relationshipTeamJSON,
		configDataJSON,
		gracePeriod.CreatedAt,
		gracePeriod.UpdatedAt,
	)

	return err
}

