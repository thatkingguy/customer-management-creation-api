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

type SectorRepository interface {
	FindByCodeOrDescription(ctx context.Context, code string) (*models.Sector, error)
}

type sectorRepository struct {
	db *sqlx.DB
}

func NewSectorRepository(db *sqlx.DB) SectorRepository {
	return &sectorRepository{db: db}
}

func (r *sectorRepository) FindByCodeOrDescription(ctx context.Context, code string) (*models.Sector, error) {
	// Extract numeric code (handles "4200" or "4200-ABC" formats)
	code = utils.ExtractNumericCode(code)

	var sectorID int
	if _, err := fmt.Sscanf(code, "%d", &sectorID); err != nil {
		logger.Log.WithError(err).WithField("code", code).Error("FindByCodeOrDescription: Invalid sector code format")
		return nil, fmt.Errorf("invalid sector code format: %s", code)
	}

	// Primary key lookup - should be very fast, no need for nested timeout
	query := `
		SELECT "sectorId", description
		FROM sector
		WHERE "sectorId" = $1
		LIMIT 1
	`
	
	var sector models.Sector
	err := r.db.GetContext(ctx, &sector, query, sectorID)
	if err == sql.ErrNoRows {
		logger.Log.WithField("sector_id", sectorID).Debug("FindByCodeOrDescription: Sector not found")
		return nil, nil
	}
	if err != nil {
		logger.Log.WithError(err).WithField("sector_id", sectorID).Error("FindByCodeOrDescription: Query failed")
		return nil, err
	}
	
	logger.Log.WithField("sector_id", sector.SectorID).Debug("FindByCodeOrDescription: Sector found")
	return &sector, nil
}

