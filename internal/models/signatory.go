package models

import (
	"time"

	"github.com/google/uuid"
)

// Signatory represents the signatory table
type Signatory struct {
	SignatoryID                    uuid.UUID  `db:"signatoryId" json:"signatoryId"`
	Title                          *string    `db:"title" json:"title"`
	FirstName                      *string    `db:"firstName" json:"firstName"`
	Surname                        *string    `db:"surname" json:"surname"`
	OtherNames                     *string    `db:"otherNames" json:"otherNames"`
	MotherMaidenName               *string    `db:"motherMaidenName" json:"motherMaidenName"`
	Gender                         *string    `db:"gender" json:"gender"`
	DateOfBirth                    *string    `db:"dateOfBirth" json:"dateOfBirth"`
	MaritalStatus                  *string    `db:"maritalStatus" json:"maritalStatus"`
	Country                        *string    `db:"country" json:"country"`
	City                           *string    `db:"city" json:"city"`
	StateOfOrigin                  *string    `db:"stateOfOrigin" json:"stateOfOrigin"`
	LGA                            *string    `db:"lga" json:"lga"`
	OtherIDType                    *string    `db:"otherIdType" json:"otherIdType"`
	IDNumber                       *string    `db:"idNumber" json:"idNumber"`
	IDIssueDate                    *string    `db:"idIssueDate" json:"idIssueDate"`
	IDExpiryDate                   *string    `db:"idExpiryDate" json:"idExpiryDate"`
	MobileNumber                   *string    `db:"mobileNumber" json:"mobileNumber"`
	AlternativeMobileNumber        *string    `db:"alternativeMobileNumber" json:"alternativeMobileNumber"`
	EmailAddress                   *string    `db:"emailAddress" json:"emailAddress"`
	ResidentialAddress             *string    `db:"residentialAddress" json:"residentialAddress"`
	ResidentialAddressDetailedDesc *string    `db:"residentialAddressDetailedDesc" json:"residentialAddressDetailedDesc"`
	MeansOfIdentification          *string    `db:"meansOfIdentification" json:"meansOfIdentification"`
	JobTitle                       *string    `db:"jobTitle" json:"jobTitle"`
	Occupation                     *string    `db:"occupation" json:"occupation"`
	Signature                      *string    `db:"signature" json:"signature"`
	Date                           *string    `db:"date" json:"date"`
	CityOrTown                     *string    `db:"cityOrTown" json:"cityOrTown"`
	Nationality                    *string    `db:"nationality" json:"nationality"`
	SignatoryClass                 *string    `db:"signatoryClass" json:"signatoryClass"`
	Position                       *string    `db:"position" json:"position"`
	CustomerID                     *uuid.UUID `db:"customerId" json:"customerId"`
	EmploymentStatus               *string    `db:"employmentStatus" json:"employmentStatus"`
	NatureOfBusiness               *string    `db:"natureOfBusiness" json:"natureOfBusiness"`
	PassportPhotograph             *string    `db:"passportPhotograph" json:"passportPhotograph"`
	ProofOfIdentity                *string    `db:"proofOfIdentity" json:"proofOfIdentity"`
	ProofOfAddress                 *string    `db:"proofOfAddress" json:"proofOfAddress"`
	CustomerSignature              *string    `db:"customerSignature" json:"customerSignature"`
	PrimarySignatory               bool       `db:"primarySignatory" json:"primarySignatory"`
	BVN                            *string    `db:"bvn" json:"bvn"`
	NIN                            *string    `db:"nin" json:"nin"`
	CreatedAt                      time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt                      time.Time  `db:"updatedAt" json:"updatedAt"`
}

