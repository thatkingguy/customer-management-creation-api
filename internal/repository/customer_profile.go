package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type CustomerProfileRepository interface {
	Create(ctx context.Context, tx *sql.Tx, profile *models.CustomerProfile) error
	FindDuplicateIndividual(ctx context.Context, tx *sql.Tx, firstName, surname string, dateOfBirth time.Time, bvn, nin *string) (*models.CustomerProfile, error)
	FindDuplicateSME(ctx context.Context, tx *sql.Tx, companyName string, taxID *string, dateOfReg *time.Time) (*models.CustomerProfile, error)
	CheckPhoneExists(ctx context.Context, phone string) (bool, error)
	GetNextCustomerNumber(ctx context.Context, tx *sql.Tx) (string, error)
}

type customerProfileRepository struct {
	db *sqlx.DB
}

func NewCustomerProfileRepository(db *sqlx.DB) CustomerProfileRepository {
	return &customerProfileRepository{db: db}
}

func (r *customerProfileRepository) Create(ctx context.Context, tx *sql.Tx, profile *models.CustomerProfile) error {
	query := `
		INSERT INTO customer_profile (
			"customerProfileId", "customerId", "customerNumber", "title", "firstName",
			"surname", "otherNames", "mobileNumber", "alternateMobileNumber",
			"emailAddress", "dateOfBirth", "companyNameBusiness", "bvn", "nin",
			"certificateOfIncorporation", "taxIdentificationNumber", "nationality",
			"categoryOfBusiness", "dateOfRegistration", "introducer", "sectorId",
			"industryId", "targetId", "customerProfileData", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		)
	`

	now := time.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now

	profileDataJSON, err := profile.CustomerProfileData.Value()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query,
		profile.CustomerProfileID,
		profile.CustomerID,
		profile.CustomerNumber,
		profile.Title,
		profile.FirstName,
		profile.Surname,
		profile.OtherNames,
		profile.MobileNumber,
		profile.AlternateMobileNumber,
		profile.EmailAddress,
		profile.DateOfBirth,
		profile.CompanyNameBusiness,
		profile.BVN,
		profile.NIN,
		profile.CertificateOfIncorporation,
		profile.TaxIdentificationNumber,
		profile.Nationality,
		profile.CategoryOfBusiness,
		profile.DateOfRegistration,
		profile.Introducer,
		profile.SectorID,
		profile.IndustryID,
		profile.TargetID,
		profileDataJSON,
		profile.CreatedAt,
		profile.UpdatedAt,
	)

	return err
}

func (r *customerProfileRepository) FindDuplicateIndividual(ctx context.Context, tx *sql.Tx, firstName, surname string, dateOfBirth time.Time, bvn, nin *string) (*models.CustomerProfile, error) {
	// Optimized query: Check BVN first if available (usually indexed), then NIN, then name match
	var query string
	var args []interface{}

	if bvn != nil && *bvn != "" {
		// If BVN is provided, check BVN first (most unique identifier)
		query = `
			SELECT cp."customerId", cp."firstName", cp."surname"
			FROM customer_profile cp
			WHERE cp."bvn" = $1
				AND cp."firstName" = $2
				AND cp."surname" = $3
				AND cp."dateOfBirth" = $4
			LIMIT 1
		`
		args = []interface{}{*bvn, firstName, surname, dateOfBirth}
	} else if nin != nil && *nin != "" {
		// If NIN is provided, check NIN
		query = `
			SELECT cp."customerId", cp."firstName", cp."surname"
			FROM customer_profile cp
			WHERE cp."nin" = $1
				AND cp."firstName" = $2
				AND cp."surname" = $3
				AND cp."dateOfBirth" = $4
			LIMIT 1
		`
		args = []interface{}{*nin, firstName, surname, dateOfBirth}
	} else {
		// Fallback to name + dateOfBirth only
		query = `
			SELECT cp."customerId", cp."firstName", cp."surname"
			FROM customer_profile cp
			WHERE cp."firstName" = $1
				AND cp."surname" = $2
				AND cp."dateOfBirth" = $3
			LIMIT 1
		`
		args = []interface{}{firstName, surname, dateOfBirth}
	}

	logger.Log.WithFields(map[string]interface{}{
		"firstName":   firstName,
		"surname":     surname,
		"dateOfBirth": dateOfBirth.Format("2006-01-02"),
		"has_bvn":     bvn != nil,
		"has_nin":     nin != nil,
		"query_type": func() string {
			if bvn != nil && *bvn != "" {
				return "BVN-based"
			}
			if nin != nil && *nin != "" {
				return "NIN-based"
			}
			return "name-based"
		}(),
		"using_tx": tx != nil,
	}).Debug("FindDuplicateIndividual: Executing optimized query with transaction")

	// Add timeout to context to prevent hanging queries
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var profile models.CustomerProfile

	// Use transaction connection if provided, otherwise use regular connection
	if tx != nil {
		// Use transaction connection - this is critical to avoid deadlocks
		rows, err := tx.QueryContext(queryCtx, query, args...)
		if err != nil {
			if err == context.DeadlineExceeded {
				logger.Log.WithError(err).Error("FindDuplicateIndividual: Query timeout (10s exceeded)")
				return nil, fmt.Errorf("query timeout: database query took too long: %w", err)
			}
			logger.Log.WithError(err).Error("FindDuplicateIndividual: Query failed")
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			logger.Log.Debug("FindDuplicateIndividual: No duplicate found")
			return nil, nil
		}

		err = rows.Scan(&profile.CustomerID, &profile.FirstName, &profile.Surname)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Log.Debug("FindDuplicateIndividual: No duplicate found")
				return nil, nil
			}
			logger.Log.WithError(err).Error("FindDuplicateIndividual: Scan failed")
			return nil, err
		}
	} else {
		// Fallback to regular connection (shouldn't happen in transaction context)
		err := r.db.GetContext(queryCtx, &profile, query, args...)
		if err == sql.ErrNoRows {
			logger.Log.Debug("FindDuplicateIndividual: No duplicate found")
			return nil, nil
		}
		if err != nil {
			if err == context.DeadlineExceeded {
				logger.Log.WithError(err).Error("FindDuplicateIndividual: Query timeout (10s exceeded)")
				return nil, fmt.Errorf("query timeout: database query took too long: %w", err)
			}
			logger.Log.WithError(err).Error("FindDuplicateIndividual: Query failed")
			return nil, err
		}
	}

	logger.Log.WithField("customer_id", profile.CustomerID).Debug("FindDuplicateIndividual: Duplicate found")
	return &profile, nil
}

func (r *customerProfileRepository) FindDuplicateSME(ctx context.Context, tx *sql.Tx, companyName string, taxID *string, dateOfReg *time.Time) (*models.CustomerProfile, error) {
	var query string
	var args []interface{}

	if taxID != nil && dateOfReg != nil {
		// All fields present
		query = `
			SELECT cp."customerId", cp."companyNameBusiness"
			FROM customer_profile cp
			WHERE LOWER(cp."companyNameBusiness") = LOWER($1)
				AND LOWER(cp."taxIdentificationNumber") = LOWER($2)
				AND cp."dateOfRegistration" = $3
			LIMIT 1
		`
		args = []interface{}{companyName, *taxID, *dateOfReg}
	} else {
		// Only company name
		query = `
			SELECT cp."customerId", cp."companyNameBusiness"
			FROM customer_profile cp
			WHERE LOWER(cp."companyNameBusiness") = LOWER($1)
			LIMIT 1
		`
		args = []interface{}{companyName}
	}

	// Add timeout to context to prevent hanging queries
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var profile models.CustomerProfile

	// Use transaction connection if provided
	if tx != nil {
		rows, err := tx.QueryContext(queryCtx, query, args...)
		if err != nil {
			if err == context.DeadlineExceeded {
				logger.Log.WithError(err).Error("FindDuplicateSME: Query timeout (10s exceeded)")
				return nil, fmt.Errorf("query timeout: %w", err)
			}
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, nil
		}

		err = rows.Scan(&profile.CustomerID, &profile.CompanyNameBusiness)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
	} else {
		err := r.db.GetContext(queryCtx, &profile, query, args...)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			if err == context.DeadlineExceeded {
				logger.Log.WithError(err).Error("FindDuplicateSME: Query timeout (10s exceeded)")
				return nil, fmt.Errorf("query timeout: %w", err)
			}
			return nil, err
		}
	}

	return &profile, nil
}

func (r *customerProfileRepository) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {
	// Optimized query: Only check indexed mobileNumber column (JSONB checks are too slow)
	// The mobileNumber column is indexed and should be the primary source of truth
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM customer_profile cp
			INNER JOIN customer c ON cp."customerId" = c."customerId"
			WHERE c."customerType" = 'Individual'
			AND cp."mobileNumber" = $1
			LIMIT 1
		) as exists
	`

	// Add timeout to prevent hanging queries
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var exists bool
	err := r.db.GetContext(queryCtx, &exists, query, phone)
	if err == context.DeadlineExceeded {
		logger.Log.WithError(err).Error("CheckPhoneExists: Query timeout (3s exceeded)")
		return false, fmt.Errorf("query timeout: phone check took too long: %w", err)
	}
	if err != nil {
		logger.Log.WithError(err).Error("CheckPhoneExists: Query failed")
		return false, err
	}
	return exists, nil
}

func (r *customerProfileRepository) GetNextCustomerNumber(ctx context.Context, tx *sql.Tx) (string, error) {
	var sequenceValue int64
	// Sequence calls are fast, but add timeout for safety
	queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	err := tx.QueryRowContext(queryCtx, `SELECT nextval('customer_number_seq')`).Scan(&sequenceValue)
	if err != nil {
		return "", err
	}

	// Format as 10-digit zero-padded string (e.g., 0000000001, 0000000002, etc.)
	customerNumber := fmt.Sprintf("%010d", sequenceValue)
	return customerNumber, nil
}
