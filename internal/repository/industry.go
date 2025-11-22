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

type IndustryRepository interface {
	FindByCodeOrDescription(ctx context.Context, code string) (*models.Industry, error)
}

type industryRepository struct {
	db *sqlx.DB
}

func NewIndustryRepository(db *sqlx.DB) IndustryRepository {
	return &industryRepository{db: db}
}

func (r *industryRepository) FindByCodeOrDescription(ctx context.Context, code string) (*models.Industry, error) {
	// Extract numeric code (handles "5000" or "5000-ABC" formats)
	code = utils.ExtractNumericCode(code)

	var industryID int
	if _, err := fmt.Sscanf(code, "%d", &industryID); err != nil {
		logger.Log.WithError(err).WithField("code", code).Error("FindByCodeOrDescription: Invalid industry code format")
		return nil, fmt.Errorf("invalid industry code format: %s", code)
	}

	// Primary key lookup - should be very fast, no need for nested timeout
	query := `
		SELECT "industryId", "sectorId", description
		FROM industry
		WHERE "industryId" = $1
		LIMIT 1
	`
	
	var industry models.Industry
	err := r.db.GetContext(ctx, &industry, query, industryID)
	if err == sql.ErrNoRows {
		logger.Log.WithField("industry_id", industryID).Debug("FindByCodeOrDescription: Industry not found")
		return nil, nil
	}
	if err != nil {
		logger.Log.WithError(err).WithField("industry_id", industryID).Error("FindByCodeOrDescription: Query failed")
		return nil, err
	}

	logger.Log.WithField("industry_id", industry.IndustryID).Debug("FindByCodeOrDescription: Industry found")
	return &industry, nil
}

