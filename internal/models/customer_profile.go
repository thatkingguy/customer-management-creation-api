package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CustomerProfile represents the customer_profile table
type CustomerProfile struct {
	CustomerProfileID          uuid.UUID  `db:"customerProfileId" json:"customerProfileId"`
	CustomerID                 uuid.UUID  `db:"customerId" json:"customerId"`
	CustomerNumber             string     `db:"customerNumber" json:"customerNumber"`
	Title                      *string    `db:"title" json:"title"`
	FirstName                  *string    `db:"firstName" json:"firstName"`
	Surname                    *string    `db:"surname" json:"surname"`
	OtherNames                 *string    `db:"otherNames" json:"otherNames"`
	MobileNumber               *string    `db:"mobileNumber" json:"mobileNumber"`
	AlternateMobileNumber      *string    `db:"alternateMobileNumber" json:"alternateMobileNumber"`
	EmailAddress               *string    `db:"emailAddress" json:"emailAddress"`
	DateOfBirth                *time.Time `db:"dateOfBirth" json:"dateOfBirth"`
	CompanyNameBusiness        *string    `db:"companyNameBusiness" json:"companyNameBusiness"`
	BVN                        *string    `db:"bvn" json:"bvn"`
	NIN                        *string    `db:"nin" json:"nin"`
	CertificateOfIncorporation *string    `db:"certificateOfIncorporation" json:"certificateOfIncorporation"`
	TaxIdentificationNumber    *string    `db:"taxIdentificationNumber" json:"taxIdentificationNumber"`
	Nationality                *string    `db:"nationality" json:"nationality"`
	CategoryOfBusiness         *string    `db:"categoryOfBusiness" json:"categoryOfBusiness"`
	DateOfRegistration         *time.Time `db:"dateOfRegistration" json:"dateOfRegistration"`
	Introducer                 *uuid.UUID `db:"introducer" json:"introducer"`
	SectorID                   *int       `db:"sectorId" json:"sectorId"`
	IndustryID                 *int       `db:"industryId" json:"industryId"`
	TargetID                   *int       `db:"targetId" json:"targetId"`
	CustomerProfileData        JSONB      `db:"customerProfileData" json:"customerProfileData"`
	CreatedAt                  time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt                  time.Time  `db:"updatedAt" json:"updatedAt"`
}

// JSONB is a custom type for PostgreSQL JSONB
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), j)
	}
	return json.Unmarshal(bytes, j)
}

// CustomerProfileLookup represents the customer_profile_lookup table
type CustomerProfileLookup struct {
	CustomerProfileLookupID    uuid.UUID  `db:"customerProfileLookupId" json:"customerProfileLookupId"`
	FirstName                  *string    `db:"firstName" json:"firstName"`
	Surname                    *string    `db:"surname" json:"surname"`
	DateOfBirth                *time.Time `db:"dateOfBirth" json:"dateOfBirth"`
	BVN                        *string    `db:"bvn" json:"bvn"`
	NIN                        *string    `db:"nin" json:"nin"`
	CompanyNameBusiness        *string    `db:"companyNameBusiness" json:"companyNameBusiness"`
	CertificateOfIncorporation *string    `db:"certificateOfIncorporation" json:"certificateOfIncorporation"`
	TaxIdentificationNumber    *string    `db:"taxIdentificationNumber" json:"taxIdentificationNumber"`
	DateOfRegistration         *time.Time `db:"dateOfRegistration" json:"dateOfRegistration"`
	CreatedAt                  time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt                  time.Time  `db:"updatedAt" json:"updatedAt"`
}
