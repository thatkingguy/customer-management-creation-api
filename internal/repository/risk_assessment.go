package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type RiskAssessmentRepository interface {
	Create(ctx context.Context, tx *sql.Tx, assessment *models.RiskAssessment) error
	CreateBatch(ctx context.Context, tx *sql.Tx, assessments []*models.RiskAssessment) error
}

type riskAssessmentRepository struct {
	db *sqlx.DB
}

func NewRiskAssessmentRepository(db *sqlx.DB) RiskAssessmentRepository {
	return &riskAssessmentRepository{db: db}
}

func (r *riskAssessmentRepository) Create(ctx context.Context, tx *sql.Tx, assessment *models.RiskAssessment) error {
	query := `
		INSERT INTO risk_assessment (
			"riskAssessmentId", "customerId", "sectionName", "parameter", "impliedWeight",
			"parameterOption", "assessmentType", "escalationFactor", "optionsWeightAllocation",
			"score", "data", "typeOfPoliticalExposure", "typeOfPoliticalExposureDesc",
			"createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	var dataJSON interface{}
	if assessment.Data != nil {
		dataJSON = assessment.Data
	}

	_, err := tx.ExecContext(ctx, query,
		assessment.RiskAssessmentID,
		assessment.CustomerID,
		assessment.SectionName,
		assessment.Parameter,
		assessment.ImpliedWeight,
		assessment.ParameterOption,
		assessment.AssessmentType,
		assessment.EscalationFactor,
		assessment.OptionsWeightAllocation,
		assessment.Score,
		dataJSON,
		assessment.TypeOfPoliticalExposure,
		assessment.TypeOfPoliticalExposureDesc,
		assessment.CreatedAt,
		assessment.UpdatedAt,
	)

	return err
}

func (r *riskAssessmentRepository) CreateBatch(ctx context.Context, tx *sql.Tx, assessments []*models.RiskAssessment) error {
	if len(assessments) == 0 {
		return nil
	}

	// Use batch insert for better performance
	// Build query with multiple value sets
	query := `
		INSERT INTO risk_assessment (
			"riskAssessmentId", "customerId", "sectionName", "parameter", "impliedWeight",
			"parameterOption", "assessmentType", "escalationFactor", "optionsWeightAllocation",
			"score", "data", "typeOfPoliticalExposure", "typeOfPoliticalExposureDesc",
			"createdAt", "updatedAt"
		) VALUES `
	
	// Build values and args
	values := make([]string, 0, len(assessments))
	args := make([]interface{}, 0, len(assessments)*15)
	argIndex := 1

	for _, assessment := range assessments {
		values = append(values, `($`+strconv.Itoa(argIndex)+`,$`+strconv.Itoa(argIndex+1)+`,$`+strconv.Itoa(argIndex+2)+`,$`+strconv.Itoa(argIndex+3)+`,$`+strconv.Itoa(argIndex+4)+`,$`+strconv.Itoa(argIndex+5)+`,$`+strconv.Itoa(argIndex+6)+`,$`+strconv.Itoa(argIndex+7)+`,$`+strconv.Itoa(argIndex+8)+`,$`+strconv.Itoa(argIndex+9)+`,$`+strconv.Itoa(argIndex+10)+`,$`+strconv.Itoa(argIndex+11)+`,$`+strconv.Itoa(argIndex+12)+`,$`+strconv.Itoa(argIndex+13)+`,$`+strconv.Itoa(argIndex+14)+`)`)
		
		var dataJSON interface{}
		if assessment.Data != nil {
			dataJSON = assessment.Data
		}

		args = append(args,
			assessment.RiskAssessmentID,
			assessment.CustomerID,
			assessment.SectionName,
			assessment.Parameter,
			assessment.ImpliedWeight,
			assessment.ParameterOption,
			assessment.AssessmentType,
			assessment.EscalationFactor,
			assessment.OptionsWeightAllocation,
			assessment.Score,
			dataJSON,
			assessment.TypeOfPoliticalExposure,
			assessment.TypeOfPoliticalExposureDesc,
			assessment.CreatedAt,
			assessment.UpdatedAt,
		)
		argIndex += 15
	}

	query += strings.Join(values, ",")
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

