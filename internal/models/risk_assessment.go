package models

import (
	"time"

	"github.com/google/uuid"
)

// RiskAssessment represents the risk_assessment table
type RiskAssessment struct {
	RiskAssessmentID              uuid.UUID `db:"riskAssessmentId" json:"riskAssessmentId"`
	CustomerID                    *uuid.UUID `db:"customerId" json:"customerId"`
	SectionName                   *string    `db:"sectionName" json:"sectionName"`
	Parameter                     string     `db:"parameter" json:"parameter"`
	ImpliedWeight                 int        `db:"impliedWeight" json:"impliedWeight"`
	ParameterOption               string     `db:"parameterOption" json:"parameterOption"`
	AssessmentType                *string    `db:"assessmentType" json:"assessmentType"`
	EscalationFactor              *int       `db:"escalationFactor" json:"escalationFactor"`
	OptionsWeightAllocation       int        `db:"optionsWeightAllocation" json:"optionsWeightAllocation"`
	Score                         int        `db:"score" json:"score"`
	Data                          JSONB      `db:"data" json:"data"`
	TypeOfPoliticalExposure      *string    `db:"typeOfPoliticalExposure" json:"typeOfPoliticalExposure"`
	TypeOfPoliticalExposureDesc  *string    `db:"typeOfPoliticalExposureDesc" json:"typeOfPoliticalExposureDesc"`
	CreatedAt                     time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt                     time.Time  `db:"updatedAt" json:"updatedAt"`
}

