package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type CustomerProfileLookupRepository interface {
	Create(ctx context.Context, tx *sql.Tx, lookup *models.CustomerProfileLookup) error
}

type customerProfileLookupRepository struct {
	db *sqlx.DB
}

func NewCustomerProfileLookupRepository(db *sqlx.DB) CustomerProfileLookupRepository {
	return &customerProfileLookupRepository{db: db}
}

func (r *customerProfileLookupRepository) Create(ctx context.Context, tx *sql.Tx, lookup *models.CustomerProfileLookup) error {
	query := `
		INSERT INTO customer_profile_lookup (
			"customerProfileLookupId", "firstName", "surname", "dateOfBirth",
			"bvn", "nin", "companyNameBusiness", "certificateOfIncorporation",
			"taxIdentificationNumber", "dateOfRegistration", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	now := time.Now()
	lookup.CreatedAt = now
	lookup.UpdatedAt = now

	_, err := tx.ExecContext(ctx, query,
		lookup.CustomerProfileLookupID,
		lookup.FirstName,
		lookup.Surname,
		lookup.DateOfBirth,
		lookup.BVN,
		lookup.NIN,
		lookup.CompanyNameBusiness,
		lookup.CertificateOfIncorporation,
		lookup.TaxIdentificationNumber,
		lookup.DateOfRegistration,
		lookup.CreatedAt,
		lookup.UpdatedAt,
	)

	return err
}
