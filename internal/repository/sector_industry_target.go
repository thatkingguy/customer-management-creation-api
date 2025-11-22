package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/utils"
)

type SectorIndustryTargetRepository interface {
	FindAllByCodes(ctx context.Context, sectorCode, industryCode, targetCode string) (*models.Sector, *models.Industry, *models.Target, error)
}

type sectorIndustryTargetRepository struct {
	db *sqlx.DB
}

func NewSectorIndustryTargetRepository(db *sqlx.DB) SectorIndustryTargetRepository {
	return &sectorIndustryTargetRepository{db: db}
}

// FindAllByCodes fetches sector, industry, and target in a single query
// This reduces network round trips from 3 to 1, significantly improving performance
func (r *sectorIndustryTargetRepository) FindAllByCodes(ctx context.Context, sectorCode, industryCode, targetCode string) (*models.Sector, *models.Industry, *models.Target, error) {
	// Extract numeric codes
	sectorCode = utils.ExtractNumericCode(sectorCode)
	industryCode = utils.ExtractNumericCode(industryCode)
	targetCode = utils.ExtractNumericCode(targetCode)

	// Parse IDs
	var sectorID, industryID, targetID int
	var err error

	if sectorCode != "" {
		if _, err = fmt.Sscanf(sectorCode, "%d", &sectorID); err != nil {
			return nil, nil, nil, fmt.Errorf("invalid sector code format: %s", sectorCode)
		}
	}
	if industryCode != "" {
		if _, err = fmt.Sscanf(industryCode, "%d", &industryID); err != nil {
			return nil, nil, nil, fmt.Errorf("invalid industry code format: %s", industryCode)
		}
	}
	if targetCode != "" {
		if _, err = fmt.Sscanf(targetCode, "%d", &targetID); err != nil {
			return nil, nil, nil, fmt.Errorf("invalid target code format: %s", targetCode)
		}
	}

	// Combined query using CTEs - fetches all three in a single round trip
	// Using CTEs with separate queries wrapped in subqueries for proper LIMIT handling
	query := `
		WITH sector_data AS (
			SELECT 
				'sector' as entity_type,
				"sectorId"::text as id,
				description as description,
				NULL::text as sector_id,
				NULL::text as short_name
			FROM sector
			WHERE "sectorId" = $1
			LIMIT 1
		),
		industry_data AS (
			SELECT 
				'industry' as entity_type,
				"industryId"::text as id,
				description as description,
				"sectorId"::text as sector_id,
				NULL::text as short_name
			FROM industry
			WHERE "industryId" = $2
			LIMIT 1
		),
		target_data AS (
			SELECT 
				'target' as entity_type,
				"targetId"::text as id,
				COALESCE(description, '') as description,
				NULL::text as sector_id,
				COALESCE("shortName", '') as short_name
			FROM target
			WHERE "targetId" = $3
			LIMIT 1
		)
		SELECT * FROM sector_data
		UNION ALL
		SELECT * FROM industry_data
		UNION ALL
		SELECT * FROM target_data
	`

	type resultRow struct {
		EntityType  string  `db:"entity_type"`
		ID          string  `db:"id"`
		Description string  `db:"description"`
		SectorID    *string `db:"sector_id"`
		ShortName   *string `db:"short_name"`
	}

	rows, err := r.db.QueryxContext(ctx, query, sectorID, industryID, targetID)
	if err != nil {
		logger.Log.WithError(err).Error("FindAllByCodes: Query failed")
		return nil, nil, nil, err
	}
	defer rows.Close()

	var sector *models.Sector
	var industry *models.Industry
	var target *models.Target

	for rows.Next() {
		var row resultRow
		if err := rows.StructScan(&row); err != nil {
			logger.Log.WithError(err).Error("FindAllByCodes: Failed to scan row")
			continue
		}

		switch row.EntityType {
		case "sector":
			var id int
			if _, err := fmt.Sscanf(row.ID, "%d", &id); err == nil {
				sector = &models.Sector{
					SectorID:    id,
					Description: row.Description,
				}
				logger.Log.WithField("sector_id", sector.SectorID).Debug("FindAllByCodes: Sector found")
			}
		case "industry":
			var id int
			var indSectorID int
			if _, err := fmt.Sscanf(row.ID, "%d", &id); err == nil {
				if row.SectorID != nil {
					if _, parseErr := fmt.Sscanf(*row.SectorID, "%d", &indSectorID); parseErr != nil {
						logger.Log.WithError(parseErr).WithField("sector_id", *row.SectorID).Warn("FindAllByCodes: Failed to parse industry SectorID, using 0")
						indSectorID = 0
					}
				}
				industry = &models.Industry{
					IndustryID:  id,
					SectorID:    indSectorID,
					Description: row.Description,
				}
				logger.Log.WithField("industry_id", industry.IndustryID).Debug("FindAllByCodes: Industry found")
			}
		case "target":
			var id int
			if _, err := fmt.Sscanf(row.ID, "%d", &id); err == nil {
				targetDesc := row.Description
				targetShortName := ""
				if row.ShortName != nil {
					targetShortName = *row.ShortName
				}
				target = &models.Target{
					TargetID:    id,
					Description: &targetDesc,
					ShortName:   targetShortName,
				}
				logger.Log.WithField("target_id", target.TargetID).Debug("FindAllByCodes: Target found")
			}
		}
	}

	if err = rows.Err(); err != nil {
		logger.Log.WithError(err).Error("FindAllByCodes: Error iterating rows")
		return nil, nil, nil, err
	}

	return sector, industry, target, nil
}
