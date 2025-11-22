package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

type TargetRepository interface {
	FindByCodeOrDescription(ctx context.Context, code string) (*models.Target, error)
}

type targetRepository struct {
	db *sqlx.DB
}

func NewTargetRepository(db *sqlx.DB) TargetRepository {
	return &targetRepository{db: db}
}

func (r *targetRepository) FindByCodeOrDescription(ctx context.Context, code string) (*models.Target, error) {
	// Extract numeric code (handles "6000" or "6000-ABC" formats)
	code = utils.ExtractNumericCode(code)

	var targetID int
	if _, err := fmt.Sscanf(code, "%d", &targetID); err != nil {
		logger.Log.WithError(err).WithField("code", code).Error("FindByCodeOrDescription: Invalid target code format")
		return nil, fmt.Errorf("invalid target code format: %s", code)
	}

	// Primary key lookup - should be very fast, no need for nested timeout
	query := `
		SELECT "targetId", description, "shortName"
		FROM target
		WHERE "targetId" = $1
		LIMIT 1
	`
	
	var target models.Target
	err := r.db.GetContext(ctx, &target, query, targetID)
	if err == sql.ErrNoRows {
		logger.Log.WithField("target_id", targetID).Debug("FindByCodeOrDescription: Target not found")
		return nil, nil
	}
	if err != nil {
		logger.Log.WithError(err).WithField("target_id", targetID).Error("FindByCodeOrDescription: Query failed")
		return nil, err
	}

	logger.Log.WithField("target_id", target.TargetID).Debug("FindByCodeOrDescription: Target found")
	return &target, nil
}

