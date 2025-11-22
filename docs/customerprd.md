# Customer Creation Flow - Golang Implementation Guide

## Table of Contents
1. [Overview](#overview)
2. [Flow Diagram](#flow-diagram)
3. [Endpoints](#endpoints)
4. [Database Models](#database-models)
5. [External Services](#external-services)
6. [Step-by-Step Implementation](#step-by-step-implementation)
7. [Helper Functions](#helper-functions)
8. [Error Handling](#error-handling)

---

## Overview

This document outlines the complete customer creation flow from the dashboard. The flow consists of two main phases:

1. **Draft Creation Phase**: Maker (initiator) creates a draft request
2. **Approval Phase**: Approver reviews and approves the draft, which triggers customer creation

### Key Characteristics
- **Request Type**: `Creation`
- **Customer Types**: `Individual` or `SME`
- **Creation Modes**: `Accelerated`, `Legacy`, or `Bulk`
- **Approval Statuses**: `Approved`, `Interim Approval`, `Pending`, `Rejected`
- **Customer Status**: `Active`, `Inactive`, or `Dormant`

---

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    DRAFT CREATION PHASE                         │
└─────────────────────────────────────────────────────────────────┘

1. POST /v1/draft?initiatorBranch={branch}
   └─> Save draft request to database
   └─> Create activity log entry
   └─> Return requestId

┌─────────────────────────────────────────────────────────────────┐
│                    APPROVAL PHASE                               │
└─────────────────────────────────────────────────────────────────┘

2. PATCH /v1/request/{requestId}?approverBranch={branch}
   └─> Fetch request from database
   └─> Validate approver permissions
   └─> Process approval (Interim Approval or Approved)
       │
       ├─> Generate UUIDs (customerId, customerProfileId)
       ├─> Check interim approval config
       ├─> Create customer record
       ├─> Generate customer number and entity ID
       ├─> Process industry/sector/target data
       ├─> Create customer profile
       ├─> Update request with customerId
       ├─> Create risk assessments
       ├─> Create bank accounts (if any)
       ├─> Create signatories (if any)
       ├─> Create references (if any)
       ├─> Create activity log
       └─> Create grace period data (if Interim Approval)
```

---

## Endpoints

### 1. Create Draft Request

**Endpoint**: `POST /v1/draft?initiatorBranch={branch}`

**Request Headers**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Query Parameters**:
- `initiatorBranch` (required): Branch code of the initiator

**Request Body**:
```json
{
  "customerType": "individual",  // or "sme"
  "data": {
    "branch": "HQ",
    "customerData": [
      {
        "sectionName": "accountServices-SECTIONLESS",
        "data": {
          "phoneNumber": "",
          "surname": "Doe",
          "firstName": "John",
          "dateOfBirth": "1988-04-04",
          "bvn": "32322323232",
          "title": "Mr.",
          "notificationChannel": "SMS",
          "notificationRule": "Debit Event",
          "statementChannel": "In-branch",
          "frequency": "Monthly"
        },
        "pageId": "17371060192262010000",
        "sectionId": null
      },
      {
        "sectionName": "bioData",
        "data": {
          "surname": "Doe",
          "firstName": "John",
          "dateOfBirth": "1988-04-04",
          "maritalStatus": "Single",
          "dualCitizenship": "Yes",
          "title": "Mr",
          "gender": "Male"
        },
        "pageId": "17526727716631420000",
        "sectionId": "17526727953111770000"
      },
      {
        "sectionName": "contactInformation",
        "data": {
          "location": "",
          "phoneNumber": ""
        },
        "pageId": "17526727716631420000",
        "sectionId": "17538112756395500000"
      },
      {
        "sectionName": "identityVerification",
        "data": {
          "bvn": "32322323232"
        },
        "pageId": "17526727716631420000",
        "sectionId": "17526730342452200000"
      }
    ],
    "formInformation": {
      "formType": "accelerated",  // or "legacy" or "bulk"
      "formId": "691b8abf57202143c2660135"
    },
    "requestData": {
      "initiator": "Admin Name",
      "initiatorId": "03921ddf-2607-41dd-a6d9-091c286f3dd2",
      "requestType": "creation"
    },
    "waiverData": [],
    "signatoryData": [
      {
        "sectionName": "Signatory details",
        "data": []
      }
    ],
    "executiveData": [
      {
        "data": []
      }
    ],
    "accountData": [
      {
        "data": []
      }
    ],
    "riskAssessmentData": [
      {
        "sectionName": "Customer's Identity",
        "data": [
          {
            "parameter": "Status of Customer identity verification",
            "impliedWeight": 100,
            "parameterOption": "Not verified",
            "assessmentType": "",
            "escalationFactor": 1,
            "optionsWeightAllocation": 10,
            "score": 1000
          }
        ]
      }
    ],
    "riskData": {
      "riskScore": 72.05,
      "riskStatus": "HIGH"
    }
  }
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "message": "Creation request saved to draft",
  "data": {
    "requestId": "1b09c409-2769-4dd3-b9d7-425d13e6121e"
  },
  "responseCode": 200
}
```

**Implementation Steps**:
1. Extract user details from JWT token (`initiator`, `initiatorId`, `branch`)
2. Validate `customerType` is provided
3. Generate `requestId` (UUID v4)
4. Determine `requestType` from `data.requestData.requestType` (map to enum: `Creation`)
5. Determine `creationMode` from `formInformation.formType`:
   - `accelerated` → `Accelerated`
   - `legacy` → `Legacy`
   - `bulk` → `Bulk`
6. Build request title from customer name or default
7. Create request record with status `Draft`
8. Create activity log entry
9. Return `requestId`

---

### 2. Approve Request (Interim Approval)

**Endpoint**: `PATCH /v1/request/{requestId}?approverBranch={branch}`

**Request Headers**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**Path Parameters**:
- `requestId` (required): UUID of the request to approve

**Query Parameters**:
- `approverBranch` (required): Branch code of the approver

**Request Body**:
```json
{
  "status": "Interim Approval"  // or "Approved" or "Rejected"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "message": "Individual customer created successfully",
  "data": {
    "customerId": "e2fdfa98-2d9c-4790-9947-c9213077d4a4",
    "requestId": "1b09c409-2769-4dd3-b9d7-425d13e6121e",
    "signatoryId": null,
    "product": null
  },
  "responseCode": 200
}
```

**Implementation Steps**:
1. Extract user details from JWT token (`approver`, `approverId`, `branch`)
2. Fetch request by `requestId`
3. Validate request exists and is in correct state
4. Normalize status: `Interim Approval` → `Interim Approval`
5. If `requestType == "Creation"` and `status == "Interim Approval"`:
   - Call `processIndividualCustomerOnApproval` or `processSMECustomerOnApproval`
6. Update request status and approval status
7. Return response with `customerId`

---

## Database Models

### Go Struct Definitions

#### 1. Request Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type RequestStatus string
const (
    RequestStatusApproved        RequestStatus = "Approved"
    RequestStatusDraft           RequestStatus = "Draft"
    RequestStatusInIssue         RequestStatus = "In Issue"
    RequestStatusInReview        RequestStatus = "In-Review"
    RequestStatusInterimApproval RequestStatus = "Interim Approval"
    RequestStatusPending         RequestStatus = "Pending"
    RequestStatusRejected         RequestStatus = "Rejected"
    RequestStatusProcessing      RequestStatus = "Processing"
    RequestStatusFailed          RequestStatus = "Failed"
)

type ApprovalStatus string
const (
    ApprovalStatusApproved ApprovalStatus = "Approved"
    ApprovalStatusPending  ApprovalStatus = "Pending"
    ApprovalStatusRejected ApprovalStatus = "Rejected"
    ApprovalStatusFailed   ApprovalStatus = "Failed"
)

type RequestType string
const (
    RequestTypeCreation            RequestType = "Creation"
    RequestTypeDeactivation         RequestType = "Deactivation"
    RequestTypeModification         RequestType = "Modification"
    RequestTypeReactivation         RequestType = "Reactivation"
    RequestTypeContractModification RequestType = "ContractModification"
)

type CustomerType string
const (
    CustomerTypeIndividual CustomerType = "Individual"
    CustomerTypeSME        CustomerType = "SME"
)

type CreationMode string
const (
    CreationModeAccelerated CreationMode = "Accelerated"
    CreationModeBulk         CreationMode = "Bulk"
    CreationModeLegacy       CreationMode = "Legacy"
)

type Request struct {
    RequestID       uuid.UUID      `gorm:"type:uuid;primary_key;column:requestId" json:"requestId"`
    CustomerID      *uuid.UUID     `gorm:"type:uuid;column:customerId" json:"customerId"`
    RequestTitle    string         `gorm:"type:varchar(255);not null;column:requestTitle" json:"requestTitle"`
    RequestType     RequestType    `gorm:"type:varchar(50);not null;column:requestType" json:"requestType"`
    RequestSubType  *string        `gorm:"type:varchar(255);column:requestSubType" json:"requestSubType"`
    AccountNumber   *string        `gorm:"type:varchar(255);column:accountNumber" json:"accountNumber"`
    Justification   *string        `gorm:"type:varchar(255);column:justification" json:"justification"`
    Initiator       string         `gorm:"type:varchar(255);not null" json:"initiator"`
    InitiatorID     uuid.UUID      `gorm:"type:uuid;not null;column:initiatorId" json:"initiatorId"`
    Status          RequestStatus  `gorm:"type:varchar(50);not null;default:'Pending'" json:"status"`
    ApprovalStatus  ApprovalStatus `gorm:"type:varchar(50);not null;default:'Pending';column:approvalStatus" json:"approvalStatus"`
    Approver        *string        `gorm:"type:varchar(255)" json:"approver"`
    ApproverID      *uuid.UUID     `gorm:"type:uuid;column:approverId" json:"approverId"`
    Data            datatypes.JSON `gorm:"type:jsonb" json:"data"`
    CustomerType    *CustomerType  `gorm:"type:varchar(50);default:'Individual';column:customerType" json:"customerType"`
    CreationMode    *CreationMode  `gorm:"type:varchar(50);default:'Legacy';column:creationMode" json:"creationMode"`
    Branch          *string        `gorm:"type:varchar(255)" json:"branch"`
    ApproverBranch  *string        `gorm:"type:varchar(255);column:approverBranch" json:"approverBranch"`
    Department      *string        `gorm:"type:varchar(255)" json:"department"`
    Withdrawn       bool           `gorm:"default:false" json:"withdrawn"`
    IsDeleted       bool           `gorm:"default:false;column:isDeleted" json:"isDeleted"`
    IsProduct       bool           `gorm:"default:false;column:isProduct" json:"isProduct"`
    DeletedOn       *time.Time     `gorm:"column:deletedOn" json:"deletedOn"`
    RejectionDocument *datatypes.JSON `gorm:"type:jsonb;column:rejectionDocument" json:"rejectionDocument"`
    RejectionReason   *string      `gorm:"type:text;column:rejectionReason" json:"rejectionReason"`
    HasCollectionProduct bool       `gorm:"default:false;column:hasCollectionProduct" json:"hasCollectionProduct"`
    BulkReferenceID    *string     `gorm:"type:text;column:bulkReferenceId" json:"bulkReferenceId"`
    CreatedAt       time.Time      `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt       time.Time      `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Request) TableName() string {
    return "request"
}
```

#### 2. Customer Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type CustomerStatus string
const (
    CustomerStatusActive   CustomerStatus = "Active"
    CustomerStatusInactive CustomerStatus = "Inactive"
    CustomerStatusDormant   CustomerStatus = "Dormant"
)

type Customer struct {
    CustomerID              uuid.UUID      `gorm:"type:uuid;primary_key;column:customerId" json:"customerId"`
    CustomerType            CustomerType    `gorm:"type:varchar(50);not null;default:'Individual';column:customerType" json:"customerType"`
    Status                  CustomerStatus  `gorm:"type:varchar(50);not null;default:'Inactive'" json:"status"`
    Monitoring              bool            `gorm:"default:false" json:"monitoring"`
    IsTerminated            bool            `gorm:"default:false;column:isTerminated" json:"isTerminated"`
    ApprovalStatus          *string         `gorm:"type:varchar(255);column:approvalStatus" json:"approvalStatus"`
    TenantID                *uuid.UUID      `gorm:"type:uuid;column:tenantId" json:"tenantId"`
    Initiator               string          `gorm:"type:varchar(255);not null" json:"initiator"`
    InitiatorID             uuid.UUID       `gorm:"type:uuid;not null;column:initiatorId" json:"initiatorId"`
    Approver                *string         `gorm:"type:varchar(255)" json:"approver"`
    ApproverID              *uuid.UUID      `gorm:"type:uuid;column:approverId" json:"approverId"`
    Branch                  *string         `gorm:"type:varchar(255)" json:"branch"`
    ApproverBranch          *string          `gorm:"type:varchar(255);column:approverBranch" json:"approverBranch"`
    Department              *string         `gorm:"type:varchar(255)" json:"department"`
    RequiresRegularization  bool            `gorm:"default:false;column:requiresRegularization" json:"requiresRegularization"`
    IsAutoReactivationOnly  bool            `gorm:"default:false;column:isAutoReactivationOnly" json:"isAutoReactivationOnly"`
    CategoryOfCustomer      *string         `gorm:"type:varchar(255);default:'Regular';column:categoryOfCustomer" json:"categoryOfCustomer"`
    CategoryOfBusiness      *string         `gorm:"type:varchar(255);column:categoryOfBusiness" json:"categoryOfBusiness"`
    RelatedEntity           *uuid.UUID      `gorm:"type:uuid;column:relatedEntity" json:"relatedEntity"`
    CustomerSubType         *int            `gorm:"column:customerSubType" json:"customerSubType"`
    CreatedAt               time.Time       `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt               time.Time       `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Customer) TableName() string {
    return "customer"
}
```

#### 3. Customer Profile Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type CustomerProfile struct {
    CustomerProfileID       uuid.UUID      `gorm:"type:uuid;primary_key;column:customerProfileId" json:"customerProfileId"`
    CustomerID              uuid.UUID      `gorm:"type:uuid;not null;column:customerId" json:"customerId"`
    Title                   *string        `gorm:"type:text" json:"title"`
    FirstName               *string        `gorm:"type:text;column:firstName" json:"firstName"`
    Surname                 *string        `gorm:"type:text" json:"surname"`
    OtherNames              *string        `gorm:"type:text;column:otherNames" json:"otherNames"`
    CustomerNumber          *string        `gorm:"type:text;column:customerNumber" json:"customerNumber"`
    MobileNumber            *string        `gorm:"type:text;column:mobileNumber" json:"mobileNumber"`
    AlternateMobileNumber   *string        `gorm:"type:text;column:alternateMobileNumber" json:"alternateMobileNumber"`
    EmailAddress            *string        `gorm:"type:text;column:emailAddress" json:"emailAddress"`
    DateOfBirth             *string        `gorm:"type:text;column:dateOfBirth" json:"dateOfBirth"`
    CompanyNameBusiness     *string        `gorm:"type:text;column:companyNameBusiness" json:"companyNameBusiness"`
    BVN                     *string        `gorm:"type:text;column:bvn" json:"bvn"`
    NIN                     *string        `gorm:"type:text;column:nin" json:"nin"`
    CertificateOfIncorporation *string     `gorm:"type:text;column:certificateOfIncorporation" json:"certificateOfIncorporation"`
    TaxIdentificationNumber *string         `gorm:"type:text;column:taxIdentificationNumber" json:"taxIdentificationNumber"`
    Nationality             *string         `gorm:"type:text" json:"nationality"`
    CategoryOfBusiness      *string         `gorm:"type:text;column:categoryOfBusiness" json:"categoryOfBusiness"`
    DateOfRegistration      *string         `gorm:"type:text;column:dateOfRegistration" json:"dateOfRegistration"`
    Introducer              *string         `gorm:"type:text" json:"introducer"`
    SearchText              *string         `gorm:"type:text;column:searchText" json:"searchText"`
    IndustryID              *string         `gorm:"type:varchar(255);column:industryId" json:"industryId"`
    SectorID                *string         `gorm:"type:varchar(255);column:sectorId" json:"sectorId"`
    TargetID                *string         `gorm:"type:varchar(255);column:targetId" json:"targetId"`
    CustomerProfileData     datatypes.JSON  `gorm:"type:jsonb;column:customerProfileData" json:"customerProfileData"`
    CreatedAt               time.Time       `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt               time.Time       `gorm:"column:updatedAt" json:"updatedAt"`
}

func (CustomerProfile) TableName() string {
    return "customer_profile"
}
```

#### 4. Activity Log Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type ActivityLog struct {
    ActivityLogID   uuid.UUID      `gorm:"type:uuid;primary_key;column:activityLogId" json:"activityLogId"`
    CustomerID      *uuid.UUID     `gorm:"type:uuid;column:customerId" json:"customerId"`
    RequestID       *uuid.UUID     `gorm:"type:uuid;column:requestId" json:"requestId"`
    Description     string         `gorm:"type:text;not null" json:"description"`
    Reason          datatypes.JSON `gorm:"type:jsonb" json:"reason"`
    CreatedAt       time.Time      `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt       time.Time      `gorm:"column:updatedAt" json:"updatedAt"`
}

func (ActivityLog) TableName() string {
    return "activity_log"
}
```

#### 5. Risk Assessment Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type RiskAssessment struct {
    RiskAssessmentID           uuid.UUID      `gorm:"type:uuid;primary_key;column:riskAssessmentId" json:"riskAssessmentId"`
    CustomerID                  *uuid.UUID     `gorm:"type:uuid;column:customerId" json:"customerId"`
    SectionName                 *string        `gorm:"type:varchar(1000);column:sectionName" json:"sectionName"`
    Parameter                   string         `gorm:"type:varchar(1000);not null" json:"parameter"`
    ImpliedWeight               int            `gorm:"not null;column:impliedWeight" json:"impliedWeight"`
    ParameterOption             string         `gorm:"type:varchar(255);not null;column:parameterOption" json:"parameterOption"`
    AssessmentType              *string        `gorm:"type:varchar(255);column:assessmentType" json:"assessmentType"`
    EscalationFactor            *int           `gorm:"column:escalationFactor" json:"escalationFactor"`
    OptionsWeightAllocation     int            `gorm:"not null;column:optionsWeightAllocation" json:"optionsWeightAllocation"`
    Score                       int            `gorm:"not null" json:"score"`
    Data                        datatypes.JSON `gorm:"type:jsonb" json:"data"`
    TypeOfPoliticalExposure     *string       `gorm:"type:text;column:typeOfPoliticalExposure" json:"typeOfPoliticalExposure"`
    TypeOfPoliticalExposureDesc *string        `gorm:"type:text;column:typeOfPoliticalExposureDesc" json:"typeOfPoliticalExposureDesc"`
    CreatedAt                   time.Time      `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt                   time.Time      `gorm:"column:updatedAt" json:"updatedAt"`
}

func (RiskAssessment) TableName() string {
    return "risk_assessment"
}
```

#### 6. Signatory Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Signatory struct {
    SignatoryID                    uuid.UUID  `gorm:"type:uuid;primary_key;column:signatoryId" json:"signatoryId"`
    Title                          *string    `gorm:"type:varchar(255)" json:"title"`
    FirstName                      *string    `gorm:"type:varchar(255);column:firstName" json:"firstName"`
    Surname                        *string    `gorm:"type:varchar(255)" json:"surname"`
    OtherNames                     *string    `gorm:"type:varchar(255);column:otherNames" json:"otherNames"`
    MotherMaidenName               *string    `gorm:"type:varchar(255);column:motherMaidenName" json:"motherMaidenName"`
    Gender                         *string    `gorm:"type:varchar(255)" json:"gender"`
    DateOfBirth                    *string    `gorm:"type:varchar(255);column:dateOfBirth" json:"dateOfBirth"`
    MaritalStatus                  *string    `gorm:"type:varchar(255);column:maritalStatus" json:"maritalStatus"`
    Country                        *string    `gorm:"type:varchar(255)" json:"country"`
    City                           *string    `gorm:"type:varchar(255)" json:"city"`
    StateOfOrigin                  *string    `gorm:"type:varchar(255);column:stateOfOrigin" json:"stateOfOrigin"`
    LGA                            *string    `gorm:"type:varchar(255)" json:"lga"`
    OtherIDType                    *string    `gorm:"type:varchar(255);column:otherIdType" json:"otherIdType"`
    IDNumber                       *string    `gorm:"type:varchar(255);column:idNumber" json:"idNumber"`
    IDIssueDate                    *string    `gorm:"type:text;column:idIssueDate" json:"idIssueDate"`
    IDExpiryDate                   *string    `gorm:"type:text;column:idExpiryDate" json:"idExpiryDate"`
    MobileNumber                   *string    `gorm:"type:varchar(255);column:mobileNumber" json:"mobileNumber"`
    AlternativeMobileNumber        *string    `gorm:"type:varchar(255);column:alternativeMobileNumber" json:"alternativeMobileNumber"`
    EmailAddress                   *string    `gorm:"type:varchar(255);column:emailAddress" json:"emailAddress"`
    ResidentialAddress             *string    `gorm:"type:varchar(255);column:residentialAddress" json:"residentialAddress"`
    ResidentialAddressDetailedDesc *string   `gorm:"type:varchar(255);column:residentialAddressDetailedDesc" json:"residentialAddressDetailedDesc"`
    MeansOfIdentification          *string    `gorm:"type:varchar(255);column:meansOfIdentification" json:"meansOfIdentification"`
    JobTitle                       *string    `gorm:"type:varchar(255);column:jobTitle" json:"jobTitle"`
    Occupation                     *string    `gorm:"type:varchar(255)" json:"occupation"`
    Signature                      *string    `gorm:"type:varchar(255)" json:"signature"`
    Date                           *string    `gorm:"type:varchar(255)" json:"date"`
    CityOrTown                     *string    `gorm:"type:varchar(255);column:cityOrTown" json:"cityOrTown"`
    Nationality                    *string    `gorm:"type:varchar(255)" json:"nationality"`
    SignatoryClass                 *string    `gorm:"type:varchar(255);column:signatoryClass" json:"signatoryClass"`
    Position                       *string    `gorm:"type:varchar(255)" json:"position"`
    CustomerID                     *uuid.UUID `gorm:"type:uuid;column:customerId" json:"customerId"`
    EmploymentStatus               *string    `gorm:"type:varchar(255);column:employmentStatus" json:"employmentStatus"`
    NatureOfBusiness               *string    `gorm:"type:varchar(255);column:natureOfBusiness" json:"natureOfBusiness"`
    PassportPhotograph             *string    `gorm:"type:text;column:passportPhotograph" json:"passportPhotograph"`
    ProofOfIdentity                *string    `gorm:"type:text;column:proofOfIdentity" json:"proofOfIdentity"`
    ProofOfAddress                 *string    `gorm:"type:text;column:proofOfAddress" json:"proofOfAddress"`
    CustomerSignature              *string    `gorm:"type:text;column:customerSignature" json:"customerSignature"`
    PrimarySignatory                bool       `gorm:"default:false;column:primarySignatory" json:"primarySignatory"`
    BVN                            *string    `gorm:"type:text;column:bvn" json:"bvn"`
    NIN                            *string    `gorm:"type:text;column:nin" json:"nin"`
    CreatedAt                      time.Time  `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt                      time.Time  `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Signatory) TableName() string {
    return "signatory"
}
```

#### 7. Other Account Model (Bank Accounts)

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type OtherAccount struct {
    OtherAccountID     uuid.UUID  `gorm:"type:uuid;primary_key;column:otherAccountId" json:"otherAccountId"`
    CustomerID         uuid.UUID  `gorm:"type:uuid;not null;column:customerId" json:"customerId"`
    CustomerProfileID  uuid.UUID  `gorm:"type:uuid;not null;column:customerProfileId" json:"customerProfileId"`
    BankName           *string    `gorm:"type:varchar(255);column:bankName" json:"bankName"`
    AccountNumber      *string    `gorm:"type:varchar(255);column:accountNumber" json:"accountNumber"`
    AccountType        *string    `gorm:"type:varchar(255);column:accountType" json:"accountType"`
    CreatedAt          time.Time  `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt          time.Time  `gorm:"column:updatedAt" json:"updatedAt"`
}

func (OtherAccount) TableName() string {
    return "other_account"
}
```

#### 8. Reference Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Reference struct {
    ReferenceID    uuid.UUID  `gorm:"type:uuid;primary_key;column:referenceId" json:"referenceId"`
    CustomerID     uuid.UUID  `gorm:"type:uuid;not null;column:customerId" json:"customerId"`
    Name           *string    `gorm:"type:varchar(255)" json:"name"`
    PhoneNumber    *string    `gorm:"type:varchar(255);column:phoneNumber" json:"phoneNumber"`
    Email          *string    `gorm:"type:varchar(255)" json:"email"`
    Address        *string    `gorm:"type:text" json:"address"`
    Relationship   *string    `gorm:"type:varchar(255)" json:"relationship"`
    CreatedAt      time.Time  `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt      time.Time  `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Reference) TableName() string {
    return "reference"
}
```

#### 9. Interim Approval Config Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type ConfigStatus string
const (
    ConfigStatusApproved ConfigStatus = "Approved"
    ConfigStatusPending  ConfigStatus = "Pending"
    ConfigStatusRejected ConfigStatus = "Rejected"
)

type DurationUnit string
const (
    DurationUnitDays   DurationUnit = "Days"
    DurationUnitHours  DurationUnit = "Hours"
    DurationUnitMonths DurationUnit = "Months"
    DurationUnitWeeks  DurationUnit = "Weeks"
)

type GracePeriodExpiryAction string
const (
    GracePeriodExpiryActionDeactivateCustomer GracePeriodExpiryAction = "Deactivate Customer"
    GracePeriodExpiryActionPostNoDebit        GracePeriodExpiryAction = "Post No Debit"
)

type InterimApprovalConfig struct {
    InterimApprovalConfigID                    uuid.UUID              `gorm:"type:uuid;primary_key;column:interimApprovalConfigId" json:"interimApprovalConfigId"`
    Status                                      ConfigStatus           `gorm:"type:varchar(50);default:'Pending'" json:"status"`
    CustomerType                                CustomerType           `gorm:"type:varchar(50);not null;default:'Individual';column:customerType" json:"customerType"`
    GracePeriod                                 bool                   `gorm:"not null;column:gracePeriod" json:"gracePeriod"`
    GracePeriodBeforeAction                     *int                   `gorm:"column:gracePeriodBeforeAction" json:"gracePeriodBeforeAction"`
    GraceDurationBeforeAction                   *DurationUnit         `gorm:"type:varchar(50);default:'Days';column:graceDurationBeforeAction" json:"graceDurationBeforeAction"`
    GracePeriodExpiryAction                     *GracePeriodExpiryAction `gorm:"type:varchar(50);default:'Deactivate Customer';column:gracePeriodExpiryAction" json:"gracePeriodExpiryAction"`
    NotifyCustomer                              bool                   `gorm:"not null;column:notifyCustomer" json:"notifyCustomer"`
    NotifyCustomerPeriodBeforeAction            *int                   `gorm:"column:notifyCustomerPeriodBeforeAction" json:"notifyCustomerPeriodBeforeAction"`
    NotifyCustomerDurationBeforeAction          *DurationUnit          `gorm:"type:varchar(50);default:'Days';column:notifyCustomerDurationBeforeAction" json:"notifyCustomerDurationBeforeAction"`
    NotifyCustomerChannels                      *string                `gorm:"type:varchar(255);column:notifyCustomerChannels" json:"notifyCustomerChannels"`
    NotifyCustomerMessageTemplate               *string                `gorm:"type:text;column:notifyCustomerMessageTemplate" json:"notifyCustomerMessageTemplate"`
    NotifyRelationshipTeam                      bool                   `gorm:"not null;column:notifyRelationshipTeam" json:"notifyRelationshipTeam"`
    NotifyRelationshipTeamPeriodBeforeAction     *int                   `gorm:"column:notifyRelationshipTeamPeriodBeforeAction" json:"notifyRelationshipTeamPeriodBeforeAction"`
    NotifyRelationshipTeamDurationBeforeAction  *string                `gorm:"type:varchar(255);column:notifyRelationshipTeamDurationBeforeAction" json:"notifyRelationshipTeamDurationBeforeAction"`
    NotifyRelationshipTeamChannels              *string                `gorm:"type:varchar(255);column:notifyRelationshipTeamChannels" json:"notifyRelationshipTeamChannels"`
    NotifyRelationshipTeamMessageTemplate       *string                `gorm:"type:text;column:notifyRelationshipTeamMessageTemplate" json:"notifyRelationshipTeamMessageTemplate"`
    DepositProducts                            datatypes.JSON         `gorm:"type:jsonb;column:depositProducts" json:"depositProducts"`
    PaymentProducts                            datatypes.JSON         `gorm:"type:jsonb;column:paymentProducts" json:"paymentProducts"`
    Initiator                                   string                 `gorm:"type:varchar(255);not null" json:"initiator"`
    InitiatorID                                 uuid.UUID              `gorm:"type:uuid;not null;column:initiatorId" json:"initiatorId"`
    Approver                                    *string                `gorm:"type:varchar(255)" json:"approver"`
    ApproverID                                  *uuid.UUID             `gorm:"type:uuid;column:approverId" json:"approverId"`
    RelationshipTeam                            datatypes.JSON         `gorm:"type:jsonb;column:relationshipTeam" json:"relationshipTeam"`
    CreatedAt                                   time.Time              `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt                                   time.Time              `gorm:"column:updatedAt" json:"updatedAt"`
}

func (InterimApprovalConfig) TableName() string {
    return "interim_approval_config"
}
```

#### 10. Customer Grace Period Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type CustomerGracePeriod struct {
    CustomerGracePeriodID      uuid.UUID      `gorm:"type:uuid;primary_key;column:customerGracePeriodId" json:"customerGracePeriodId"`
    InterimApprovalConfigID     *uuid.UUID     `gorm:"type:uuid;column:interimApprovalConfigId" json:"interimApprovalConfigId"`
    CustomerID                  *uuid.UUID     `gorm:"type:uuid;column:customerId" json:"customerId"`
    GracePeriod                 bool           `gorm:"not null;column:gracePeriod" json:"gracePeriod"`
    GracePeriodDate             *string        `gorm:"type:varchar(255);column:gracePeriodDate" json:"gracePeriodDate"`
    NotifyCustomer              bool           `gorm:"not null;column:notifyCustomer" json:"notifyCustomer"`
    NotifyCustomerDate          *string        `gorm:"type:varchar(255);column:notifyCustomerDate" json:"notifyCustomerDate"`
    NotifyCustomerChannels      *string        `gorm:"type:varchar(255);column:notifyCustomerChannels" json:"notifyCustomerChannels"`
    NotifyRelationshipTeam      bool           `gorm:"not null;column:notifyRelationshipTeam" json:"notifyRelationshipTeam"`
    NotifyRelationshipTeamDate  *string        `gorm:"type:varchar(255);column:notifyRelationshipTeamDate" json:"notifyRelationshipTeamDate"`
    RelationshipTeam            datatypes.JSON `gorm:"type:jsonb;column:relationshipTeam" json:"relationshipTeam"`
    InterimApprovalConfigData   datatypes.JSON `gorm:"type:jsonb;column:interimApprovalConfigData" json:"interimApprovalConfigData"`
    CreatedAt                   time.Time      `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt                   time.Time      `gorm:"column:updatedAt" json:"updatedAt"`
}

func (CustomerGracePeriod) TableName() string {
    return "customer_grace_period"
}
```

#### 11. Waiver Request Model

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type WaiverRequest struct {
    WaiverRequestID uuid.UUID  `gorm:"type:uuid;primary_key;column:waiverRequestId" json:"waiverRequestId"`
    RequestID       *uuid.UUID `gorm:"type:uuid;column:requestId" json:"requestId"`
    CustomerID      *uuid.UUID `gorm:"type:uuid;column:customerId" json:"customerId"`
    Status          *string    `gorm:"type:varchar(50)" json:"status"`
    Approver        *string    `gorm:"type:varchar(255)" json:"approver"`
    ApproverID      *uuid.UUID `gorm:"type:uuid;column:approverId" json:"approverId"`
    Justification   *string    `gorm:"type:text" json:"justification"`
    CreatedAt       time.Time  `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt       time.Time  `gorm:"column:updatedAt" json:"updatedAt"`
}

func (WaiverRequest) TableName() string {
    return "waiver_request"
}
```

#### 12. Industry Model

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Industry struct {
    IndustryID   string    `gorm:"type:varchar(255);primary_key;column:industryId" json:"industryId"`
    Description  string    `gorm:"type:text;not null" json:"description"`
    CreatedAt    time.Time `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt    time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Industry) TableName() string {
    return "industry"
}
```

#### 13. Sector Model

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Sector struct {
    SectorID    string    `gorm:"type:varchar(255);primary_key;column:sectorId" json:"sectorId"`
    Description string    `gorm:"type:text;not null" json:"description"`
    CreatedAt   time.Time `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt   time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Sector) TableName() string {
    return "sector"
}
```

#### 14. Target Model

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Target struct {
    TargetID    string    `gorm:"type:varchar(255);primary_key;column:targetId" json:"targetId"`
    Description string    `gorm:"type:text;not null" json:"description"`
    ShortName   *string   `gorm:"type:varchar(255);column:shortName" json:"shortName"`
    CreatedAt   time.Time `gorm:"column:createdAt" json:"createdAt"`
    UpdatedAt   time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Target) TableName() string {
    return "target"
}
```

---

## External Services

### 1. ID Generation Service

**Purpose**: Generate unique 10-digit customer numbers and entity IDs

**Implementation**: This is an internal service that:
- Generates random 10-digit numbers
- Checks uniqueness against `customer_profile.customerNumber`
- Retries if duplicate found

**Go Implementation**:
```go
package services

import (
    "crypto/rand"
    "fmt"
    "gorm.io/gorm"
)

type IDGeneratorService struct {
    db *gorm.DB
}

func NewIDGeneratorService(db *gorm.DB) *IDGeneratorService {
    return &IDGeneratorService{db: db}
}

func (s *IDGeneratorService) Generate() string {
    bytes := make([]byte, 10)
    rand.Read(bytes)
    
    var number string
    for _, b := range bytes {
        number += fmt.Sprintf("%d", b%10)
    }
    
    return number[:10]
}

func (s *IDGeneratorService) CheckAndGenerateID(id *string) (string, error) {
    var generatedID string
    if id != nil {
        generatedID = *id
    } else {
        generatedID = s.Generate()
    }
    
    // Check if exists
    var exists bool
    err := s.db.Model(&models.CustomerProfile{}).
        Select("COUNT(*) > 0").
        Where("customerNumber = ?", generatedID).
        Scan(&exists).Error
    
    if err != nil {
        return "", err
    }
    
    if !exists {
        return generatedID, nil
    }
    
    // Retry with new ID
    return s.CheckAndGenerateID(nil)
}
```

### 2. Industry/Sector/Target Lookup Service

**Purpose**: Process and lookup industry, sector, and target codes/descriptions

**Implementation**: 
- Lookup existing records by description or code
- Create new records if not found
- Update descriptions if changed

**Go Implementation**:
```go
package services

import (
    "gorm.io/gorm"
    "strings"
)

type IndustrySectorTargetService struct {
    db *gorm.DB
}

func NewIndustrySectorTargetService(db *gorm.DB) *IndustrySectorTargetService {
    return &IndustrySectorTargetService{db: db}
}

func (s *IndustrySectorTargetService) ProcessIndustrySectorTarget(
    profileData *models.CustomerProfile,
    rawSectorValue *string,
    rawIndustryValue *string,
    rawTargetValue *string,
    targetShortName *string,
) error {
    // Process Industry
    if rawIndustryValue != nil && *rawIndustryValue != "" {
        industryID, industryDesc, err := s.processIndustry(*rawIndustryValue)
        if err != nil {
            return err
        }
        if industryID != "" {
            profileData.IndustryID = &industryID
            // Update customerProfileData JSON
        }
    }
    
    // Process Sector
    if rawSectorValue != nil && *rawSectorValue != "" {
        sectorID, sectorDesc, err := s.processSector(*rawSectorValue)
        if err != nil {
            return err
        }
        if sectorID != "" {
            profileData.SectorID = &sectorID
        }
    }
    
    // Process Target
    if rawTargetValue != nil && *rawTargetValue != "" {
        targetID, targetDesc, err := s.processTarget(*rawTargetValue, targetShortName)
        if err != nil {
            return err
        }
        if targetID != "" {
            profileData.TargetID = &targetID
        }
    }
    
    return nil
}

func (s *IndustrySectorTargetService) processIndustry(value string) (string, string, error) {
    // Check if it's a code (starts with number) or description
    value = strings.TrimSpace(value)
    
    // If starts with number, treat as code
    if len(value) > 0 && value[0] >= '0' && value[0] <= '9' {
        code := value
        description := value // Default description
        // Lookup or create
        var industry models.Industry
        err := s.db.Where("industryId = ?", code).First(&industry).Error
        if err == gorm.ErrRecordNotFound {
            industry = models.Industry{
                IndustryID:  code,
                Description: description,
            }
            err = s.db.Create(&industry).Error
        }
        return industry.IndustryID, industry.Description, err
    }
    
    // Otherwise, lookup by description
    var industry models.Industry
    err := s.db.Where("description = ?", value).First(&industry).Error
    if err == gorm.ErrRecordNotFound {
        // Generate code or use description as code
        code := value // Or generate unique code
        industry = models.Industry{
            IndustryID:  code,
            Description: value,
        }
        err = s.db.Create(&industry).Error
    }
    return industry.IndustryID, industry.Description, err
}

// Similar implementations for processSector and processTarget
```

---

## Step-by-Step Implementation

### Phase 1: Draft Creation

#### Step 1.1: Extract User Details from Token
```go
type UserDetails struct {
    Name   string
    ID     uuid.UUID
    Branch string
}

func extractUserDetails(token string) (*UserDetails, error) {
    // Parse JWT token and extract claims
    // Return user details
}
```

#### Step 1.2: Validate Request
```go
func validateDraftRequest(req *DraftRequestPayload) error {
    if req.CustomerType == "" {
        return errors.New("customer type not specified, Individual or SME")
    }
    if req.Data.RequestData.RequestType == "" {
        return errors.New("request type not specified")
    }
    return nil
}
```

#### Step 1.3: Create Draft Request
```go
func createDraftRequest(db *gorm.DB, payload *DraftRequestPayload, userDetails *UserDetails) (*models.Request, error) {
    requestID := uuid.New()
    
    // Determine request type
    requestType := mapRequestType(payload.Data.RequestData.RequestType)
    
    // Determine creation mode
    creationMode := mapCreationMode(payload.Data.FormInformation.FormType)
    
    // Build request title
    requestTitle := buildRequestTitle(payload, requestType)
    
    // Create request
    request := &models.Request{
        RequestID:      requestID,
        RequestTitle:   requestTitle,
        RequestType:    requestType,
        CustomerType:   mapCustomerType(payload.CustomerType),
        CreationMode:   &creationMode,
        Status:         models.RequestStatusDraft,
        ApprovalStatus: models.ApprovalStatusPending,
        Initiator:      userDetails.Name,
        InitiatorID:    userDetails.ID,
        Branch:         &userDetails.Branch,
        Data:           payload.Data,
    }
    
    err := db.Create(request).Error
    if err != nil {
        return nil, err
    }
    
    // Create activity log
    activityLog := &models.ActivityLog{
        ActivityLogID: uuid.New(),
        RequestID:     &requestID,
        Description:    fmt.Sprintf("Customer %s request saved to draft by %s", requestType, userDetails.Name),
    }
    db.Create(activityLog)
    
    return request, nil
}
```

---

### Phase 2: Approval and Customer Creation

#### Step 2.1: Fetch Request
```go
func fetchRequest(db *gorm.DB, requestID uuid.UUID) (*models.Request, error) {
    var request models.Request
    err := db.Where("requestId = ?", requestID).First(&request).Error
    if err != nil {
        return nil, err
    }
    return &request, nil
}
```

#### Step 2.2: Process Individual Customer Approval
```go
func processIndividualCustomerOnApproval(
    db *gorm.DB,
    requestID uuid.UUID,
    approverPayload *ApproverPayload,
    creationData *CreationData,
    requestStatus string,
) (*CustomerCreationResponse, error) {
    
    // Generate UUIDs
    customerID := uuid.New()
    customerProfileID := uuid.New()
    
    // Check interim config
    var interimConfig *models.InterimApprovalConfig
    err := db.Where("customerType = ? AND gracePeriod = ? AND status = ?", 
        models.CustomerTypeIndividual, true, models.ConfigStatusApproved).
        First(&interimConfig).Error
    
    hasInterimConfig := err == nil
    
    // Determine status
    status := models.CustomerStatusInactive
    approvalStatus := "Interim Approval"
    if requestStatus == "Approved" {
        approvalStatus = "Approved"
    }
    
    // Get request info
    var request models.Request
    db.Where("requestId = ?", requestID).First(&request)
    
    // Create customer
    customer, err := approveCustomerRequest(db, &ApproveCustomerPayload{
        CustomerID:      customerID,
        CustomerProfileID: customerProfileID,
        CustomerType:    models.CustomerTypeIndividual,
        CustomerData:    creationData.CustomerData,
        StatusData: StatusData{
            Status:         string(status),
            ApprovalStatus: approvalStatus,
        },
        TenantID:        creationData.TenantID,
        Initiator:       request.Initiator,
        InitiatorID:     request.InitiatorID,
        Approver:        approverPayload.Approver,
        ApproverID:      approverPayload.ApproverID,
        Branch:          request.Branch,
        Department:      request.Department,
        CreationMode:    request.CreationMode,
    })
    if err != nil {
        return nil, err
    }
    
    // Update request and activity logs
    err = updateRequestAndActivityLogs(db, requestID, customerID)
    if err != nil {
        return nil, err
    }
    
    // Create risk assessments
    if len(creationData.RiskAssessmentData) > 0 {
        err = createRiskAssessments(db, customerID, creationData.RiskAssessmentData)
        if err != nil {
            return nil, err
        }
    }
    
    // Create bank accounts
    if len(creationData.AccountData) > 0 && len(creationData.AccountData[0].Data) > 0 {
        err = createAccounts(db, customerID, customerProfileID, creationData.AccountData)
        if err != nil {
            return nil, err
        }
    }
    
    // Create signatories
    if len(creationData.SignatoryData) > 0 && len(creationData.SignatoryData[0].Data) > 0 {
        err = createSignatories(db, customerID, requestID, creationData.SignatoryData)
        if err != nil {
            return nil, err
        }
    }
    
    // Create references
    if len(creationData.ReferenceData) > 0 && len(creationData.ReferenceData[0].Data) > 0 {
        err = createReferences(db, customerID, creationData.ReferenceData)
        if err != nil {
            return nil, err
        }
    }
    
    // Create activity log
    activityLog := &models.ActivityLog{
        ActivityLogID: uuid.New(),
        CustomerID:    &customerID,
        RequestID:     &requestID,
        Description:   fmt.Sprintf("Customer creation request approved by %s", approverPayload.Approver),
    }
    db.Create(activityLog)
    
    // Create grace period data if Interim Approval
    if approvalStatus == "Interim Approval" && hasInterimConfig {
        err = createGracePeriodData(db, customerID, interimConfig)
        if err != nil {
            return nil, err
        }
    }
    
    return &CustomerCreationResponse{
        CustomerID: customerID,
        RequestID:  requestID,
        Status:     "success",
        Message:    "Individual customer created successfully",
    }, nil
}
```

#### Step 2.3: Approve Customer Request (Create Customer and Profile)
```go
func approveCustomerRequest(db *gorm.DB, payload *ApproveCustomerPayload) (*models.Customer, error) {
    
    // Parse customer data
    customerProfileObj := parseCustomerData(payload.CustomerData)
    
    // Determine approval status
    approvalStatus := payload.StatusData.ApprovalStatus
    if approvalStatus == "" {
        if payload.CreationMode != nil && 
           (*payload.CreationMode == models.CreationModeAccelerated || 
            *payload.CreationMode == models.CreationModeBulk) {
            approvalStatus = "Interim Approval"
        } else {
            approvalStatus = "Approved"
        }
    }
    
    // Determine status
    status := models.CustomerStatusInactive
    if payload.CreationMode != nil && *payload.CreationMode == models.CreationModeLegacy {
        status = models.CustomerStatusActive
    }
    
    // Create customer record
    customer := &models.Customer{
        CustomerID:     payload.CustomerID,
        CustomerType:   payload.CustomerType,
        Status:         status,
        ApprovalStatus: &approvalStatus,
        TenantID:       payload.TenantID,
        Initiator:      payload.Initiator,
        InitiatorID:    payload.InitiatorID,
        Approver:       payload.Approver,
        ApproverID:     payload.ApproverID,
        Branch:         payload.Branch,
        ApproverBranch: payload.ApproverBranch,
        Department:     payload.Department,
        CategoryOfCustomer: stringPtr("Regular"),
    }
    
    err := db.Create(customer).Error
    if err != nil {
        return nil, err
    }
    
    // Generate customer number and entity ID
    idGenService := services.NewIDGeneratorService(db)
    customerNumber, err := idGenService.CheckAndGenerateID(nil)
    if err != nil {
        return nil, err
    }
    
    customerEntityID, err := idGenService.CheckAndGenerateID(nil)
    if err != nil {
        return nil, err
    }
    
    // Parse customer info
    parsedInfo := parseCustomerInfo(customerProfileObj)
    
    // Process industry/sector/target
    industrySectorTargetService := services.NewIndustrySectorTargetService(db)
    rawIndustry := getField(customerProfileObj, "industry", "industryCode")
    rawSector := getField(customerProfileObj, "sector", "sectorCode")
    rawTarget := getField(customerProfileObj, "target", "targetCode")
    targetShortName := getField(customerProfileObj, "targetShortName", "")
    
    customerProfile := &models.CustomerProfile{
        CustomerProfileID: payload.CustomerProfileID,
        CustomerID:        payload.CustomerID,
        CustomerNumber:    &customerNumber,
        FirstName:         parsedInfo.FirstName,
        Surname:           parsedInfo.Surname,
        OtherNames:        parsedInfo.OtherNames,
        Title:             parsedInfo.Title,
        DateOfBirth:       parsedInfo.DateOfBirth,
        MobileNumber:      parsedInfo.MobileNumber,
        EmailAddress:      parsedInfo.EmailAddress,
        BVN:               parsedInfo.BVN,
        NIN:               parsedInfo.NIN,
        Nationality:       parsedInfo.Nationality,
    }
    
    // Process industry/sector/target
    err = industrySectorTargetService.ProcessIndustrySectorTarget(
        customerProfile,
        rawSector,
        rawIndustry,
        rawTarget,
        targetShortName,
    )
    if err != nil {
        return nil, err
    }
    
    // Build customerProfileData JSON
    customerProfileData := buildCustomerProfileDataJSON(customerProfileObj, customerEntityID, customerNumber)
    customerProfile.CustomerProfileData = customerProfileData
    
    // Create customer profile
    err = db.Create(customerProfile).Error
    if err != nil {
        return nil, err
    }
    
    return customer, nil
}
```

#### Step 2.4: Update Request and Activity Logs
```go
func updateRequestAndActivityLogs(db *gorm.DB, requestID uuid.UUID, customerID uuid.UUID) error {
    // Update request
    err := db.Model(&models.Request{}).
        Where("requestId = ?", requestID).
        Update("customerId", customerID).Error
    if err != nil {
        return err
    }
    
    // Update waiver request if exists
    db.Model(&models.WaiverRequest{}).
        Where("requestId = ?", requestID).
        Update("customerId", customerID)
    
    // Update activity logs
    db.Model(&models.ActivityLog{}).
        Where("requestId = ?", requestID).
        Update("customerId", customerID)
    
    return nil
}
```

#### Step 2.5: Create Risk Assessments
```go
func createRiskAssessments(db *gorm.DB, customerID uuid.UUID, riskAssessmentData []RiskAssessmentSection) error {
    var riskAssessments []models.RiskAssessment
    
    for _, section := range riskAssessmentData {
        for _, item := range section.Data {
            riskAssessment := models.RiskAssessment{
                RiskAssessmentID:       uuid.New(),
                CustomerID:             &customerID,
                SectionName:            &section.SectionName,
                Parameter:              item.Parameter,
                ImpliedWeight:          item.ImpliedWeight,
                ParameterOption:        item.ParameterOption,
                AssessmentType:         item.AssessmentType,
                EscalationFactor:       item.EscalationFactor,
                OptionsWeightAllocation: item.OptionsWeightAllocation,
                Score:                  item.Score,
                TypeOfPoliticalExposure: item.TypeOfPoliticalExposure,
                TypeOfPoliticalExposureDesc: item.TypeOfPoliticalExposureDesc,
                Data:                   item, // Store full item as JSON
            }
            riskAssessments = append(riskAssessments, riskAssessment)
        }
    }
    
    if len(riskAssessments) > 0 {
        return db.Create(&riskAssessments).Error
    }
    
    return nil
}
```

#### Step 2.6: Create Accounts
```go
func createAccounts(db *gorm.DB, customerID uuid.UUID, customerProfileID uuid.UUID, accountData []AccountSection) error {
    if len(accountData) == 0 || len(accountData[0].Data) == 0 {
        return nil
    }
    
    var accounts []models.OtherAccount
    
    for _, item := range accountData[0].Data {
        account := models.OtherAccount{
            OtherAccountID:    uuid.New(),
            CustomerID:        customerID,
            CustomerProfileID: customerProfileID,
            BankName:          item.BankName,
            AccountNumber:     item.AccountNumber,
            AccountType:       item.AccountType,
        }
        accounts = append(accounts, account)
    }
    
    if len(accounts) > 0 {
        err := db.Create(&accounts).Error
        if err != nil {
            return err
        }
        
        // Create activity log
        activityLog := &models.ActivityLog{
            ActivityLogID: uuid.New(),
            CustomerID:    &customerID,
            Description:   "Other accounts added for customer",
        }
        db.Create(activityLog)
    }
    
    return nil
}
```

#### Step 2.7: Create Signatories
```go
func createSignatories(db *gorm.DB, customerID uuid.UUID, requestID uuid.UUID, signatoryData []SignatorySection) error {
    if len(signatoryData) == 0 || len(signatoryData[0].Data) == 0 {
        return nil
    }
    
    var signatories []models.Signatory
    
    for _, item := range signatoryData[0].Data {
        signatory := models.Signatory{
            SignatoryID:     uuid.New(),
            CustomerID:      &customerID,
            Title:           item.Title,
            FirstName:       item.FirstName,
            Surname:         item.Surname,
            OtherNames:      item.OtherNames,
            DateOfBirth:     item.DateOfBirth,
            MobileNumber:    item.MobileNumber,
            EmailAddress:    item.EmailAddress,
            PrimarySignatory: item.PrimarySignatory != nil && *item.PrimarySignatory,
            BVN:             item.BVN,
            NIN:             item.NIN,
            // ... map all other fields
        }
        signatories = append(signatories, signatory)
    }
    
    if len(signatories) > 0 {
        err := db.Create(&signatories).Error
        if err != nil {
            return err
        }
        
        // Update request data with signatory data
        // ... update request.data.signatoryData
        
        // Create activity log
        activityLog := &models.ActivityLog{
            ActivityLogID: uuid.New(),
            CustomerID:    &customerID,
            Description:   "Signatories added for customer",
        }
        db.Create(activityLog)
    }
    
    return nil
}
```

#### Step 2.8: Create References
```go
func createReferences(db *gorm.DB, customerID uuid.UUID, referenceData []ReferenceSection) error {
    if len(referenceData) == 0 || len(referenceData[0].Data) == 0 {
        return nil
    }
    
    var references []models.Reference
    
    for _, item := range referenceData[0].Data {
        reference := models.Reference{
            ReferenceID: uuid.New(),
            CustomerID:  customerID,
            Name:        item.Name,
            PhoneNumber: item.PhoneNumber,
            Email:       item.Email,
            Address:     item.Address,
            Relationship: item.Relationship,
        }
        references = append(references, reference)
    }
    
    if len(references) > 0 {
        err := db.Create(&references).Error
        if err != nil {
            return err
        }
        
        // Create activity log
        activityLog := &models.ActivityLog{
            ActivityLogID: uuid.New(),
            CustomerID:    &customerID,
            Description:   "Referees added for customer",
        }
        db.Create(activityLog)
    }
    
    return nil
}
```

#### Step 2.9: Create Grace Period Data
```go
func createGracePeriodData(db *gorm.DB, customerID uuid.UUID, interimConfig *models.InterimApprovalConfig) error {
    if interimConfig == nil {
        return nil
    }
    
    customerGracePeriodID := uuid.New()
    requestDate := time.Now()
    
    // Calculate grace period date
    gracePeriodDate := calculateDate(
        requestDate,
        interimConfig.GracePeriodBeforeAction,
        interimConfig.GraceDurationBeforeAction,
    )
    
    // Calculate customer notification date
    var notifyCustomerDate *string
    if interimConfig.NotifyCustomer {
        date := calculateDate(
            requestDate,
            interimConfig.NotifyCustomerPeriodBeforeAction,
            interimConfig.NotifyCustomerDurationBeforeAction,
        )
        notifyCustomerDate = &date
    }
    
    // Calculate relationship team notification date
    var notifyRelationshipTeamDate *string
    if interimConfig.NotifyRelationshipTeam {
        date := calculateDate(
            requestDate,
            interimConfig.NotifyRelationshipTeamPeriodBeforeAction,
            interimConfig.NotifyRelationshipTeamDurationBeforeAction,
        )
        notifyRelationshipTeamDate = &date
    }
    
    gracePeriod := &models.CustomerGracePeriod{
        CustomerGracePeriodID:      customerGracePeriodID,
        InterimApprovalConfigID:     &interimConfig.InterimApprovalConfigID,
        CustomerID:                  &customerID,
        GracePeriod:                 interimConfig.GracePeriod,
        GracePeriodDate:             &gracePeriodDate,
        NotifyCustomer:              interimConfig.NotifyCustomer,
        NotifyCustomerDate:          notifyCustomerDate,
        NotifyCustomerChannels:      interimConfig.NotifyCustomerChannels,
        NotifyRelationshipTeam:      interimConfig.NotifyRelationshipTeam,
        NotifyRelationshipTeamDate:   notifyRelationshipTeamDate,
        RelationshipTeam:            interimConfig.RelationshipTeam,
        InterimApprovalConfigData:   interimConfig, // Store full config as JSON
    }
    
    err := db.Create(gracePeriod).Error
    if err != nil {
        return err
    }
    
    // Schedule reminders (if you have a scheduler service)
    // scheduler.ScheduleExpiryReminder(gracePeriodDate, customerGracePeriodID)
    // if notifyCustomerDate != nil {
    //     scheduler.ScheduleCustomerReminder(*notifyCustomerDate, customerGracePeriodID)
    // }
    // if notifyRelationshipTeamDate != nil {
    //     scheduler.ScheduleRelationshipTeamReminder(*notifyRelationshipTeamDate, customerGracePeriodID)
    // }
    
    return nil
}

func calculateDate(startDate time.Time, period *int, duration *string) string {
    if period == nil || duration == nil {
        return ""
    }
    
    var result time.Time
    switch *duration {
    case "Days":
        result = startDate.AddDate(0, 0, *period)
    case "Weeks":
        result = startDate.AddDate(0, 0, *period*7)
    case "Months":
        result = startDate.AddDate(0, *period, 0)
    case "Hours":
        result = startDate.Add(time.Duration(*period) * time.Hour)
    default:
        result = startDate
    }
    
    return result.Format("2006-01-02 15:04:05")
}
```

---

## Helper Functions

### Parse Customer Data
```go
func parseCustomerData(customerData []CustomerDataSection) map[string]interface{} {
    result := make(map[string]interface{})
    
    for _, section := range customerData {
        for key, value := range section.Data {
            // Handle different field name variations
            if key == "bVN" {
                result["bvn"] = value
            } else if key == "chooseAnID" {
                result["chooseAnId"] = value
            } else if key == "iDNumber" || key == "ID Number" {
                result["idNumber"] = value
            } else if key == "emailAddress" {
                result["email"] = value
            } else {
                result[key] = value
            }
        }
    }
    
    return result
}
```

### Parse Customer Info
```go
type ParsedCustomerInfo struct {
    FirstName     *string
    Surname       *string
    OtherNames    *string
    Title         *string
    DateOfBirth   *string
    MobileNumber  *string
    EmailAddress  *string
    BVN           *string
    NIN           *string
    Nationality   *string
    // ... other fields
}

func parseCustomerInfo(customerProfileObj map[string]interface{}) *ParsedCustomerInfo {
    return &ParsedCustomerInfo{
        FirstName:    getString(customerProfileObj, "firstName"),
        Surname:      getString(customerProfileObj, "surname"),
        OtherNames:   getString(customerProfileObj, "otherNames"),
        Title:        getString(customerProfileObj, "title"),
        DateOfBirth:  getString(customerProfileObj, "dateOfBirth"),
        MobileNumber: getString(customerProfileObj, "mobileNumber", "phoneNumber"),
        EmailAddress: getString(customerProfileObj, "emailAddress", "email"),
        BVN:          getString(customerProfileObj, "bvn", "bVN"),
        NIN:          getString(customerProfileObj, "nin"),
        Nationality:  getString(customerProfileObj, "nationality"),
    }
}
```

### Build Customer Profile Data JSON
```go
func buildCustomerProfileDataJSON(
    customerProfileObj map[string]interface{},
    customerEntityID string,
    customerNumber string,
) datatypes.JSON {
    data := make(map[string]interface{})
    
    // Copy all fields from customerProfileObj
    for k, v := range customerProfileObj {
        data[k] = v
    }
    
    // Add computed fields
    data["customerEntityId"] = customerEntityID
    data["customerNumber"] = customerNumber
    data["fullName"] = fmt.Sprintf("%s %s", 
        getString(customerProfileObj, "firstName", ""),
        getString(customerProfileObj, "surname", ""))
    
    // Normalize field names
    if bvn := getString(customerProfileObj, "bvn", "bVN"); bvn != "" {
        data["bvn"] = bvn
    }
    if idType := getString(customerProfileObj, "chooseAnId", "chooseAnID"); idType != "" {
        data["idType"] = idType
        data["chooseAnId"] = idType
    }
    if idNumber := getString(customerProfileObj, "idNumber", "iDNumber", "ID Number"); idNumber != "" {
        data["idNumber"] = idNumber
    }
    if email := getString(customerProfileObj, "email", "emailAddress"); email != "" {
        data["email"] = email
        data["emailAddress"] = email
    }
    
    jsonBytes, _ := json.Marshal(data)
    return datatypes.JSON(jsonBytes)
}
```

### Map Request Type
```go
func mapRequestType(requestType string) models.RequestType {
    switch strings.ToLower(strings.TrimSpace(requestType)) {
    case "creation":
        return models.RequestTypeCreation
    case "deactivation":
        return models.RequestTypeDeactivation
    case "reactivation":
        return models.RequestTypeReactivation
    case "modification":
        return models.RequestTypeModification
    default:
        return models.RequestTypeCreation
    }
}
```

### Map Creation Mode
```go
func mapCreationMode(formType string) models.CreationMode {
    switch strings.ToLower(formType) {
    case "accelerated":
        return models.CreationModeAccelerated
    case "bulk":
        return models.CreationModeBulk
    case "legacy":
        return models.CreationModeLegacy
    default:
        return models.CreationModeLegacy
    }
}
```

### Map Customer Type
```go
func mapCustomerType(customerType string) models.CustomerType {
    if strings.ToLower(customerType) == "sme" {
        return models.CustomerTypeSME
    }
    return models.CustomerTypeIndividual
}
```

### Build Request Title
```go
func buildRequestTitle(payload *DraftRequestPayload, requestType models.RequestType) string {
    // Extract customer name from customerData
    var firstName, surname string
    
    for _, section := range payload.Data.CustomerData {
        if section.SectionName == "bioData" {
            if val, ok := section.Data["firstName"].(string); ok {
                firstName = val
            }
            if val, ok := section.Data["surname"].(string); ok {
                surname = val
            }
        }
    }
    
    if firstName != "" || surname != "" {
        name := strings.TrimSpace(fmt.Sprintf("%s %s", firstName, surname))
        return fmt.Sprintf("%s of %s", requestType, name)
    }
    
    return fmt.Sprintf("%s request", requestType)
}
```

### Utility Functions
```go
func getString(m map[string]interface{}, keys ...string) *string {
    for _, key := range keys {
        if val, ok := m[key]; ok {
            if str, ok := val.(string); ok && str != "" {
                return &str
            }
        }
    }
    return nil
}

func stringPtr(s string) *string {
    return &s
}

func getField(m map[string]interface{}, keys ...string) *string {
    return getString(m, keys...)
}
```

---

## Error Handling

### Common Errors

1. **Customer Type Not Specified**
   - Status Code: 400
   - Message: "Customer type not specified, Individual or SME"

2. **Request Not Found**
   - Status Code: 404
   - Message: "Request not found"

3. **Phone Number Already Exists** (for Individual customers)
   - Status Code: 400
   - Message: "Customer with Phone number already exists"

4. **Database Errors**
   - Status Code: 500
   - Message: "Internal server error"
   - Log full error details

### Error Response Format
```go
type ErrorResponse struct {
    Status      string `json:"status"`
    Message     string `json:"message"`
    ResponseCode int   `json:"responseCode"`
}
```

---

## Database Transactions

**Important**: All operations in the approval phase should be wrapped in a database transaction to ensure atomicity:

```go
func approveRequest(db *gorm.DB, requestID uuid.UUID, status string) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // All database operations here use tx instead of db
        // If any operation fails, entire transaction is rolled back
        
        // Fetch request
        var request models.Request
        if err := tx.Where("requestId = ?", requestID).First(&request).Error; err != nil {
            return err
        }
        
        // Process customer creation
        // ... all operations
        
        return nil
    })
}
```

---

## Testing Checklist

- [ ] Draft creation with valid payload
- [ ] Draft creation with missing customerType
- [ ] Draft creation with missing requestType
- [ ] Approval with Interim Approval status
- [ ] Approval with Approved status
- [ ] Approval with Rejected status
- [ ] Customer creation with all related data (risk assessment, accounts, signatories, references)
- [ ] Customer creation with Interim Approval and grace period
- [ ] ID generation uniqueness
- [ ] Industry/sector/target lookup and creation
- [ ] Transaction rollback on error
- [ ] Phone number duplicate check (Individual customers)

---

## Notes

1. **UUID Generation**: Use `github.com/google/uuid` for UUID generation
2. **JSON Handling**: Use `gorm.io/datatypes` for JSONB fields
3. **Database**: Use GORM for database operations
4. **Validation**: Validate all inputs before processing
5. **Logging**: Log all major operations for debugging
6. **Performance**: Consider batch operations for bulk inserts
7. **Grace Period Scheduling**: Implement reminder scheduling service separately
8. **Phone Number Check**: For Individual customers, check if phone number already exists before creation

---

## Additional Considerations

1. **Product Assignment**: If `productData` is present in the request, products need to be assigned to the customer (this may involve calling an external product service)
2. **Account Transfer**: If `accountTransferData` is present, handle account transfers
3. **Waiver Requests**: Update waiver request status when approving
4. **Audit Logging**: Consider implementing audit log records for compliance
5. **Caching**: Consider caching interim approval configs for performance

---

## Conclusion

This document provides a comprehensive guide for implementing the customer creation flow in Go. Follow the step-by-step implementation guide, use the provided model definitions, and ensure proper error handling and transaction management.

For questions or clarifications, refer to the original Node.js implementation or contact the development team.

