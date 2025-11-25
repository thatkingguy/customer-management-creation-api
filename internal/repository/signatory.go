package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type SignatoryRepository interface {
	Create(ctx context.Context, tx *sql.Tx, signatory *models.Signatory) error
	CreateBatch(ctx context.Context, tx *sql.Tx, signatories []*models.Signatory) error
}

type signatoryRepository struct {
	db *sqlx.DB
}

func NewSignatoryRepository(db *sqlx.DB) SignatoryRepository {
	return &signatoryRepository{db: db}
}

func (r *signatoryRepository) Create(ctx context.Context, tx *sql.Tx, signatory *models.Signatory) error {
	query := `
		INSERT INTO signatory (
			"signatoryId", "title", "firstName", "surname", "otherNames", "motherMaidenName",
			"gender", "dateOfBirth", "maritalStatus", "country", "city", "stateOfOrigin",
			"lga", "otherIdType", "idNumber", "idIssueDate", "idExpiryDate", "mobileNumber",
			"alternativeMobileNumber", "emailAddress", "residentialAddress", "residentialAddressDetailedDesc",
			"meansOfIdentification", "jobTitle", "occupation", "signature", "date", "cityOrTown",
			"nationality", "signatoryClass", "position", "customerId", "employmentStatus",
			"natureOfBusiness", "passportPhotograph", "proofOfIdentity", "proofOfAddress",
			"customerSignature", "primarySignatory", "bvn", "nin", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42
		)
	`

	_, err := tx.ExecContext(ctx, query,
		signatory.SignatoryID,
		signatory.Title,
		signatory.FirstName,
		signatory.Surname,
		signatory.OtherNames,
		signatory.MotherMaidenName,
		signatory.Gender,
		signatory.DateOfBirth,
		signatory.MaritalStatus,
		signatory.Country,
		signatory.City,
		signatory.StateOfOrigin,
		signatory.LGA,
		signatory.OtherIDType,
		signatory.IDNumber,
		signatory.IDIssueDate,
		signatory.IDExpiryDate,
		signatory.MobileNumber,
		signatory.AlternativeMobileNumber,
		signatory.EmailAddress,
		signatory.ResidentialAddress,
		signatory.ResidentialAddressDetailedDesc,
		signatory.MeansOfIdentification,
		signatory.JobTitle,
		signatory.Occupation,
		signatory.Signature,
		signatory.Date,
		signatory.CityOrTown,
		signatory.Nationality,
		signatory.SignatoryClass,
		signatory.Position,
		signatory.CustomerID,
		signatory.EmploymentStatus,
		signatory.NatureOfBusiness,
		signatory.PassportPhotograph,
		signatory.ProofOfIdentity,
		signatory.ProofOfAddress,
		signatory.CustomerSignature,
		signatory.PrimarySignatory,
		signatory.BVN,
		signatory.NIN,
		signatory.CreatedAt,
		signatory.UpdatedAt,
	)

	return err
}

func (r *signatoryRepository) CreateBatch(ctx context.Context, tx *sql.Tx, signatories []*models.Signatory) error {
	if len(signatories) == 0 {
		return nil
	}

	for _, signatory := range signatories {
		if err := r.Create(ctx, tx, signatory); err != nil {
			return err
		}
	}

	return nil
}




