# Golang Customer Creation Microservice Documentation

## Overview
This document provides detailed specifications for implementing a production-ready Golang microservice from scratch using Gin framework and PostgreSQL. The service will handle customer creation operations with a focus on scalability, maintainability, and best practices.

## Table of Contents
1. [Project Structure](#project-structure)
2. [Application Setup Requirements](#application-setup-requirements)
3. [Endpoint Specification](#endpoint-specification)
4. [Implementation Flow](#implementation-flow)
5. [Database Operations](#database-operations)
6. [Error Handling](#error-handling)
7. [Testing Considerations](#testing-considerations)

## Project Structure

```
customer-service/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── database/
│   │   ├── connection.go         # Database connection setup
│   │   ├── migrations/           # SQL migration files
│   │   └── sequences.go          # Sequence initialization
│   ├── models/
│   │   ├── customer.go           # Customer model structs
│   │   ├── customer_profile.go   # Customer profile model structs
│   │   └── activity_log.go       # Activity log model structs
│   ├── repository/
│   │   ├── customer.go           # Customer repository (database operations)
│   │   ├── customer_profile.go   # Customer profile repository
│   │   ├── sector.go             # Sector repository
│   │   ├── industry.go           # Industry repository
│   │   ├── target.go             # Target repository
│   │   └── activity_log.go       # Activity log repository
│   ├── service/
│   │   ├── customer.go           # Customer business logic
│   │   ├── validation.go         # Validation service
│   │   └── id_generator.go       # ID generation service
│   ├── handler/
│   │   ├── channel/
│   │   │   └── customer.go       # Channel endpoint handlers
│   │   └── middleware/
│   │       ├── auth.go           # API key authentication middleware
│   │       ├── logger.go         # Request logging middleware
│   │       └── recovery.go       # Panic recovery middleware
│   ├── dto/
│   │   ├── request.go            # Request DTOs
│   │   └── response.go           # Response DTOs
│   ├── utils/
│   │   ├── errors.go             # Custom error types
│   │   ├── response.go           # Response formatting utilities
│   │   ├── validation.go         # Validation utilities
│   │   └── date.go               # Date utilities
│   └── logger/
│       └── logger.go              # Structured logging setup
├── pkg/
│   └── postgres/
│       └── transaction.go        # Transaction management utilities
├── api/
│   └── docs/
│       └── swagger.yaml           # API documentation (OpenAPI/Swagger)
├── scripts/
│   ├── migrate.sh                # Migration script
│   └── seed.sh                   # Database seeding script
├── tests/
│   ├── integration/              # Integration tests
│   └── unit/                     # Unit tests
├── deployments/
│   ├── Dockerfile                # Docker image definition
│   └── docker-compose.yml       # Local development setup
├── .env.example                  # Environment variables template
├── .gitignore
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── Makefile                      # Build and deployment commands
└── README.md                     # Project documentation
```

## Application Setup Requirements

### 1. Go Module Initialization
- Initialize Go module: `go mod init github.com/yourorg/customer-service`
- Minimum Go version: 1.21+
- Use Go modules for dependency management

### 2. Core Dependencies
```go
// Required packages
- github.com/gin-gonic/gin          // Web framework
- github.com/lib/pq                 // PostgreSQL driver
- github.com/google/uuid            // UUID generation
- github.com/jmoiron/sqlx           // Extended database driver (recommended)
- github.com/joho/godotenv         // Environment variable loading
- github.com/swaggo/gin-swagger     // Swagger documentation (optional)
- github.com/swaggo/files           // Swagger files (optional)
```

### 3. Configuration Management
**File**: `internal/config/config.go`

**Environment Variables**:
- `PORT`: Server port (default: 8080)
- `ENV`: Environment (development, staging, production)
- `DB_HOST`: Database host
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `DB_SSL_MODE`: SSL mode (disable, require, verify-full)
- `DB_MAX_OPEN_CONNS`: Maximum open connections (default: 25)
- `DB_MAX_IDLE_CONNS`: Maximum idle connections (default: 5)
- `DB_CONN_MAX_LIFETIME`: Connection max lifetime in minutes (default: 5)
- `API_KEY`: API key for channel endpoints
- `LOG_LEVEL`: Log level (debug, info, warn, error)
- `LOG_FORMAT`: Log format (json, text)

**Configuration Structure**:
```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    API      APIConfig
    Logger   LoggerConfig
}

type ServerConfig struct {
    Port string
    Env  string
}

type DatabaseConfig struct {
    Host            string
    Port            string
    User            string
    Password        string
    Name            string
    SSLMode         string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

type APIConfig struct {
    Key string
}

type LoggerConfig struct {
    Level  string
    Format string
}
```

### 4. Database Connection Setup
**File**: `internal/database/connection.go`

**Requirements**:
- Connection pooling with configurable limits
- Health check function
- Graceful connection closing
- Retry logic for initial connection
- Connection string builder with SSL support
- Use `sqlx` or `database/sql` with `lib/pq`

**Connection Pool Settings**:
- SetMaxOpenConns: From config (default 25)
- SetMaxIdleConns: From config (default 5)
- SetConnMaxLifetime: From config (default 5 minutes)
- SetConnMaxIdleTime: 10 minutes

### 5. Database Migrations
**Directory**: `internal/database/migrations/`

**Migration Files**:
- `001_create_sequences.sql`: Create PostgreSQL sequences
- `002_create_indexes.sql`: Create necessary indexes (if not exists in main DB)
- `003_create_customer_profile_lookup_indexes.sql`: Lookup table indexes

**Migration Tool**: Use `golang-migrate/migrate` or custom migration runner

**Sequence Creation**:
```sql
-- 001_create_sequences.sql
CREATE SEQUENCE IF NOT EXISTS customer_number_seq START 1;
CREATE SEQUENCE IF NOT EXISTS customer_entity_id_seq START 1;
```

### 6. Logging Setup
**File**: `internal/logger/logger.go`

**Requirements**:
- Structured logging (JSON format for production, text for development)
- Log levels: DEBUG, INFO, WARN, ERROR
- Request ID tracking
- Correlation IDs for distributed tracing
- Use `logrus`, `zap`, or standard `log` package with structured output

**Log Fields**:
- Timestamp
- Level
- Message
- Request ID (if available)
- Error details (if error)
- Context fields

### 7. Error Handling
**File**: `internal/utils/errors.go`

**Custom Error Types**:
```go
type AppError struct {
    Code    int
    Message string
    Err     error
}

// Error types
- ValidationError
- NotFoundError
- DuplicateError
- DatabaseError
- InternalError
```

**Error Response Format**:
```go
type ErrorResponse struct {
    Status      string `json:"status"`
    Message     string `json:"message"`
    ResponseCode int   `json:"responseCode"`
}
```

### 8. Response Formatting
**File**: `internal/utils/response.go`

**Response Structure**:
```go
type SuccessResponse struct {
    Status      string      `json:"status"`
    Message     string      `json:"message"`
    Data        interface{} `json:"data"`
    ResponseCode int       `json:"responseCode"`
}
```

### 9. Middleware Setup
**Directory**: `internal/handler/middleware/`

**Required Middleware**:
1. **Authentication Middleware** (`auth.go`):
   - Validate API key from `X-API-Key` header
   - Return 401 if invalid/missing

2. **Logger Middleware** (`logger.go`):
   - Log incoming requests
   - Log response status and duration
   - Include request ID

3. **Recovery Middleware** (`recovery.go`):
   - Catch panics
   - Log panic details
   - Return 500 error response

4. **Request ID Middleware** (`request_id.go`):
   - Generate/use request ID from header
   - Add to context

5. **CORS Middleware** (if needed):
   - Configure allowed origins
   - Handle preflight requests

### 10. Router Setup
**File**: `cmd/server/main.go` or `internal/router/router.go`

**Route Groups**:
- `/channel/*`: Channel endpoints (API key auth required)
- `/customer/*`: Customer endpoints (for future dashboard)
- `/health`: Health check endpoint
- `/ready`: Readiness check endpoint

**Middleware Order**:
1. Recovery
2. Logger
3. Request ID
4. CORS (if needed)
5. Authentication (for protected routes)

### 11. Health Check Endpoints
**Endpoints**:
- `GET /health`: Basic health check (always returns 200)
- `GET /ready`: Readiness check (checks database connection)

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "database": "connected"
}
```

### 12. Graceful Shutdown
**Requirements**:
- Handle SIGTERM and SIGINT signals
- Wait for in-flight requests to complete (with timeout)
- Close database connections gracefully
- Log shutdown process

**Implementation**:
- Use `context.Context` for cancellation
- Set shutdown timeout (e.g., 30 seconds)
- Use `http.Server.Shutdown()`

### 13. Transaction Management
**File**: `pkg/postgres/transaction.go`

**Requirements**:
- Transaction wrapper function
- Automatic rollback on error
- Context support for cancellation
- Nested transaction support (if needed)

**Usage Pattern**:
```go
func (r *Repository) CreateCustomer(ctx context.Context, data CustomerData) error {
    return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
        // Database operations
        return nil
    })
}
```

### 14. Testing Structure
**Directory**: `tests/`

**Test Types**:
- **Unit Tests**: Test individual functions/methods
- **Integration Tests**: Test database operations with test DB
- **Handler Tests**: Test HTTP endpoints with test server

**Test Utilities**:
- Test database setup/teardown
- Test data fixtures
- Mock generators

### 15. Docker Setup
**File**: `deployments/Dockerfile`

**Multi-stage Build**:
- Build stage: Compile Go application
- Runtime stage: Minimal image with compiled binary

**Base Image**: `golang:1.21-alpine` for build, `alpine:latest` for runtime

**File**: `deployments/docker-compose.yml`
- Application service
- PostgreSQL service (for local development)
- Environment variables

### 16. Makefile
**File**: `Makefile`

**Commands**:
- `make build`: Build application
- `make run`: Run application locally
- `make test`: Run tests
- `make test-coverage`: Generate test coverage
- `make migrate-up`: Run database migrations
- `make migrate-down`: Rollback migrations
- `make docker-build`: Build Docker image
- `make docker-run`: Run Docker container
- `make lint`: Run linter
- `make fmt`: Format code

### 17. Code Organization Best Practices

**Repository Pattern**:
- Separate data access layer (repository) from business logic (service)
- Repository handles all database operations
- Service contains business logic and orchestrates repository calls

**Dependency Injection**:
- Use interfaces for repositories and services
- Inject dependencies through constructors
- Enable easy testing with mocks

**Context Propagation**:
- Pass `context.Context` through all layers
- Use context for cancellation and timeouts
- Include request ID in context

**Error Handling**:
- Return errors, don't panic (except in unrecoverable situations)
- Wrap errors with context
- Use custom error types for different scenarios

**Validation**:
- Validate input at handler level
- Use validation tags or custom validators
- Return clear validation error messages

### 18. API Documentation
**File**: `api/docs/swagger.yaml`

**Requirements**:
- OpenAPI 3.0 specification
- Endpoint documentation
- Request/response schemas
- Authentication requirements
- Error responses

**Optional**: Use `swaggo/swag` for automatic generation from code comments

### 19. Environment-Specific Configuration
**Files**: `.env.example`, `.env.development`, `.env.production`

**Best Practices**:
- Never commit `.env` files
- Provide `.env.example` with all required variables
- Use different configs for different environments
- Validate required environment variables on startup

### 20. Database Best Practices

**Connection Management**:
- Use connection pooling
- Set appropriate pool sizes based on load
- Monitor connection pool metrics

**Query Optimization**:
- Use prepared statements for repeated queries
- Use indexes appropriately
- Avoid N+1 query problems
- Use transactions for related operations

**Migration Management**:
- Version all migrations
- Test migrations on staging before production
- Have rollback strategy for each migration

### 21. Monitoring and Observability

**Metrics** (Future Enhancement):
- Request duration
- Error rates
- Database query duration
- Connection pool usage

**Logging**:
- Structured logging for easy parsing
- Include correlation IDs
- Log at appropriate levels

**Tracing** (Future Enhancement):
- Distributed tracing support
- Request flow tracking

### 22. Security Best Practices

**API Security**:
- Validate and sanitize all inputs
- Use parameterized queries (prevent SQL injection)
- Rate limiting (future enhancement)
- Input size limits

**Database Security**:
- Use SSL/TLS for database connections
- Store credentials securely (environment variables, secrets manager)
- Use least privilege principle for database user

**Application Security**:
- Keep dependencies updated
- Use `go mod verify` to check dependencies
- Regular security audits

### 23. Build and Deployment

**Build Process**:
- Use Go build flags for version info
- Build statically linked binaries
- Optimize binary size

**Deployment**:
- Use containerization (Docker)
- Health check endpoints for orchestration
- Zero-downtime deployment strategy
- Rollback capability

### 24. Development Workflow

**Local Development**:
- Use docker-compose for local database
- Hot reload capability (air, reflex, or similar)
- Local environment configuration

**Code Quality**:
- Use `golangci-lint` or similar
- Format code with `gofmt` or `goimports`
- Pre-commit hooks for linting and formatting

**Git Workflow**:
- Feature branches
- Pull request reviews
- CI/CD pipeline for testing and deployment

## Endpoint Specification

### Route
- **Path**: `/channel/directCreateCustomer`
- **Method**: `POST`
- **Authentication**: API Key via `X-API-Key` or `x-api-key` header

### Request Payload Structure
```json
{
  "data": {
    "branch": "002",                    // Required - string
    "customerType": "Individual",       // Required - "Individual" or "SME"
    "customerData": [                   // Required - array with at least one element
      {
        "sectionName": "bioData",        // Required - string
        "data": {                        // Required - object
          "branch": "002",               // Optional - string
          "title": "Mr",                 // Optional - string
          "firstName": "John",           // Required - string
          "surname": "Nwosu",            // Required - string
          "otherNames": "Chris",         // Optional - string
          "street": "Kemi STREET",       // Optional - string
          "target": "1",                 // Required - string/number
          "residence": "Ilupeju",        // Optional - string
          "nationality": "NG",           // Optional - string
          "bvn": "1234567849",          // Required for Individual - string (numeric only)
          "nin": null,                  // Optional - string (numeric only)
          "dateOfBirth": "1941-07-15",  // Required - ISO 8601 date string
          "gender": "MALE",             // Optional - string
          "emailAddress": "email@example.com", // Required - string
          "mobileNumber": "12346118939", // Required - string
          "state": "Imo",               // Optional - string
          "sectorCode": "4200",         // Required - string/number
          "sectorDescription": "ADMIN AND SUPPORT SERVICE ACT.", // Optional
          "industryCode": "4202",       // Required - string/number
          "industryDescription": "Employment activities", // Optional
          "targetCode": "1",            // Required - string/number
          "targetDescription": "Youth Banking", // Optional
          "introducer": "5b2c30af-8da0-4e02-98b7-4f9fc8c11dd3" // Optional - UUID string
        }
      }
    ],
    "formInformation": {
      "formType": "accelerated"         // Required - string
    },
    "requestData": {
      "requestType": "creation"         // Required - string
    },
    "tenantId": "uuid",                 // Optional - UUID string
    "department": "SYSTEM_DEPARTMENT",  // Optional - string
    "relatedEntity": "uuid",            // Optional - UUID string
    "flagCustomer": {                   // Optional - object
      "status": false                   // Optional - boolean
    }
  }
}
```

### Response Structure
**Success Response (201 Created)**:
```json
{
  "status": "success",
  "message": "Customer created successfully",
  "data": {
    "customerId": "32afb0ec-a7c8-450a-b55e-d38506d37a8b",
    "requestId": "fe0e2106-d338-4132-b36c-864111ec9020"
  },
  "responseCode": 200
}
```

**Error Response (400 Bad Request)**:
```json
{
  "status": "error",
  "message": "Error message here",
  "responseCode": 400
}
```

## Implementation Flow

### Step 1: API Key Validation
1. Extract API key from `X-API-Key` or `x-api-key` header (case-insensitive)
2. Compare with configured API key (from environment/config)
3. Return 401 Unauthorized if missing or invalid

### Step 2: Payload Validation
1. Validate `data` object exists
2. Validate `customerData` is non-empty array
3. Validate `customerType` exists and is "Individual" or "SME" (case-insensitive, normalize to "Individual" or "SME")
4. Validate `branch` exists in payload
5. Extract `customerData[0].data` as `bioData`
6. Validate required fields based on customer type:
   - **Individual**: firstName, surname, dateOfBirth, bvn, emailAddress, mobileNumber, sectorCode, industryCode, targetCode
   - **SME**: companyNameBusiness, taxIdentificationNumber, dateOfRegistration, sectorCode, industryCode, targetCode

### Step 3: Data Normalization
1. Normalize `customerType`: lowercase input → "Individual" or "SME"
2. Handle field name variations (case-insensitive matching):
   - `bvn` or `bVN` → use `bvn`
   - `nin` → use `nin`
   - `email` or `emailAddress` → use `emailAddress`
   - `mobileNumber`, `mobile`, `phoneNumber`, `contactNumber` → use `mobileNumber`
   - `chooseAnId` or `chooseAnID` → use `chooseAnId`
   - `idNumber` or `iDNumber` or `ID Number` → use `idNumber`
   - `certificateOfIncorporation` or `certificateofincorporation` or `rcNumber` → use `certificateOfIncorporation`
3. Remove null/empty values from bioData
4. Extract branch priority: `branch` from payload → `bioData.branch` → fallback to system default

### Step 4: validateCustomerBioData
**Location**: Called before database operations

**Validation Rules**:
1. **BVN Validation** (if present):
   - Must be string type
   - Must contain only digits (`/^\d+$/`)
   - Must not contain asterisks (`*`) (redacted check)
   - Throw error: `"BVN must be a string."` or `"BVN must contain only numbers."` or `"BVN must not be redacted (should not contain '*')."`

2. **NIN Validation** (if present):
   - Same rules as BVN
   - Error messages: `"NIN must be a string."` etc.

3. **Tax Identification Number Validation** (if present):
   - Must be string type
   - Must not contain asterisks
   - Error messages: `"Tax Identification Number must be a string."` etc.

4. **Certificate of Incorporation/RC Number Validation** (if present):
   - Must be string type
   - Must not contain asterisks
   - Error messages: `"Certificate of Incorporation must be a string."` etc.

5. **Date of Birth Validation** (if present):
   - Must be valid ISO 8601 date format
   - Error message: `"Date must be a valid date string in ISO 8601 or RFC 2822 format."`

6. **Phone Number Duplicate Check** (if phone exists and age > 18):
   - Extract phone from: `phoneNumber`, `mobile`, `mobileNumber`, `contactNumber`
   - Calculate age from `dateOfBirth`
   - If age > 18, check if phone exists in database (see Step 5)
   - For channel requests, skip throwing error (just return)

**Note**: This validation should NOT throw errors for channel requests if phone exists (per Node.js implementation with `isChannelRequest = true`)

### Step 5: checkPhoneExists
**Location**: Called after validateCustomerBioData

**Logic**:
1. Extract phone number from bioData (check keys: `phoneNumber`, `mobile`, `mobileNumber`, `contactNumber`)
2. If no phone found, return `false`
3. If `dateOfBirth` not present, log warning and return `false`
4. Calculate age from `dateOfBirth`
5. If age <= 18, return `false` (skip check for minors)
6. If age > 18:
   - Query database:
     ```sql
     SELECT EXISTS (
       SELECT 1 
       FROM customer_profile cp
       INNER JOIN customer c ON cp."customerId" = c."customerId"
       WHERE c."customerType" = 'Individual'
       AND (
         cp."mobileNumber" = $1
         OR cp."customerProfileData"->>'phoneNumber' = $1
         OR cp."customerProfileData"->>'phone' = $1
         OR cp."customerProfileData"->>'mobile' = $1
         OR cp."customerProfileData"->>'mobileNumber' = $1
         OR cp."customerProfileData"->>'contactNumber' = $1
       )
       LIMIT 1
     ) as exists;
     ```
   - Return `true` if exists, `false` otherwise
7. If phone exists, throw error: `"Customer with Phone number already exists"`

### Step 6: lookupCustomer
**Location**: Called after checkPhoneExists, before transaction starts

**Purpose**: Check for duplicate customers and insert into `customer_profile_lookup` table

**For Individual Customers**:
1. Parse customer data to extract: `firstName`, `surname`, `dateOfBirth`, `bvn`, `nin`
2. Build query conditions:
   - Base conditions: `firstName`, `surname`, `dateOfBirth` (all must match)
   - OR conditions: `bvn` OR `nin` (if present)
3. Query `customer_profile` table:
   ```sql
   SELECT cp."customerId", cp."firstName", cp."surname"
   FROM customer_profile cp
   WHERE cp."firstName" = $1
     AND cp."surname" = $2
     AND cp."dateOfBirth" = $3
     AND (
       cp."bvn" = $4 OR cp."nin" = $5
     )
   LIMIT 1;
   ```
4. If duplicate found, throw error: `"Customer with BVN and NIN already exists"` or `"Customer with details passed already exists"`
5. If no duplicate, prepare insert data for `customer_profile_lookup`:
   - `firstName`, `surname`, `dateOfBirth`
   - `bvn` (if present)
   - `nin` (if present)
6. Insert into `customer_profile_lookup` table
7. If unique constraint violation occurs (race condition):
   - Re-query `customer_profile` table
   - If duplicate found, throw same error as above

**For SME Customers**:
1. Parse customer data to extract: `companyNameBusiness`, `taxIdentificationNumber`, `dateOfRegistration`
2. Check which fields are present:
   - **All fields present**: Query with all three fields (case-insensitive for companyNameBusiness and taxIdentificationNumber)
   - **Only companyNameBusiness**: Query with company name only
   - **Missing required**: Throw error with missing fields
3. Query `customer_profile` table:
   ```sql
   -- If all fields present:
   SELECT cp."customerId", cp."companyNameBusiness"
   FROM customer_profile cp
   WHERE LOWER(cp."companyNameBusiness") = LOWER($1)
     AND LOWER(cp."taxIdentificationNumber") = LOWER($2)
     AND cp."dateOfRegistration" = $3
   LIMIT 1;
   
   -- If only companyNameBusiness:
   SELECT cp."customerId", cp."companyNameBusiness"
   FROM customer_profile cp
   WHERE LOWER(cp."companyNameBusiness") = LOWER($1)
   LIMIT 1;
   ```
4. If duplicate found, throw `SequelizeUniqueConstraintError`-like error: `"Customer with companyNameBusiness and taxIdentificationNumber and dateOfRegistration already exists"`
5. If no duplicate, prepare insert data for `customer_profile_lookup`:
   - `companyNameBusiness` (always)
   - `certificateOfIncorporation` (if present, use `rcNumber` or `certificateOfIncorporation`)
   - `taxIdentificationNumber` (if present)
   - `dateOfRegistration` (if present)
6. Insert into `customer_profile_lookup` table
7. Handle race conditions same as Individual

### Step 7: Start Database Transaction
1. Begin PostgreSQL transaction
2. Ensure transaction is properly managed (defer rollback on error, commit on success)

### Step 8: Generate IDs
1. Generate `customerId`: UUID v4
2. Generate `customerProfileId`: UUID v4
3. Generate `customerNumber`: Use PostgreSQL sequence (create sequence if not exists: `customer_number_seq`)
   ```sql
   SELECT nextval('customer_number_seq')::text;
   ```
   - Ensure uniqueness: Check `customer_profile.customerNumber` doesn't exist
   - If exists, get next sequence value (sequences are unique by design)
4. Generate `customerEntityId`: Use PostgreSQL sequence (create sequence if not exists: `customer_entity_id_seq`)
   ```sql
   SELECT nextval('customer_entity_id_seq')::text;
   ```

### Step 9: Process Industry/Sector/Target Data
**Location**: Before creating customer_profile record

**Logic** (Option 4b - Only reference existing, fail if not found):
1. Extract `sectorCode`, `industryCode`, `targetCode` from bioData
2. Parse codes:
   - If code is numeric string (e.g., "4200"), use as-is
   - If code contains "-" (e.g., "1 - Youth Banking"), extract number part before "-"
   - If code is description only, query by description (case-insensitive)
3. **Validate Sector**:
   ```sql
   SELECT "sectorId", description
   FROM sector
   WHERE "sectorId" = $1 OR LOWER(description) = LOWER($1)
   LIMIT 1;
   ```
   - If not found, throw error: `"Sector not found: {sectorCode}"`
   - Store `sectorId` and `sectorDescription`
4. **Validate Industry**:
   ```sql
   SELECT "industryId", "sectorId", description
   FROM industry
   WHERE "industryId" = $1 OR LOWER(description) = LOWER($1)
   LIMIT 1;
   ```
   - If not found, throw error: `"Industry not found: {industryCode}"`
   - Store `industryId` and `industryDescription`
   - Validate industry's `sectorId` matches provided sector (if both provided)
5. **Validate Target**:
   ```sql
   SELECT "targetId", description, "shortName"
   FROM target
   WHERE "targetId" = $1 OR LOWER(description) = LOWER($1)
   LIMIT 1;
   ```
   - If not found, throw error: `"Target not found: {targetCode}"`
   - Store `targetId` and `targetDescription`
6. Store validated IDs for use in customer_profile record

### Step 10: Prepare Customer Record
**Table**: `customer`

**Fields to Set**:
- `customerId`: Generated UUID
- `customerType`: Normalized ("Individual" or "SME")
- `status`: "Active" (hardcoded for directCreateCustomer)
- `monitoring`: `flagCustomer?.status` or `false`
- `approvalStatus`: "Approved" (hardcoded)
- `tenantId`: From payload or generate UUID
- `initiator`: "SYSTEM_USER" (hardcoded)
- `initiatorId`: Generate UUID
- `approver`: "SYSTEM_USER" (same as initiator)
- `approverId`: Same as initiatorId
- `branch`: From payload (required)
- `approverBranch`: Same as branch
- `department`: From payload or "SYSTEM_DEPARTMENT"
- `categoryOfCustomer`: From bioData or "Regular" (default)
- `customerSubType`: Map from enum (see Customer Sub-Type Enum section)
- `relatedEntity`: From payload (optional)

**Customer Sub-Type Enum Mapping**:
```go
// Map customerSubType string to integer
var customerSubTypeMap = map[string]int{
    "Individual": 1,
    "PrivateBankInd": 10,
    "Youth Corper": 11,
    "Salary Customer": 12,
    "Minor": 13,
    "Student": 14,
    "Not for Profit": 15,
    "Joint Customers": 16,
    "PrivateBankCorp": 17,
    "MSME": 18,
    "Corporate": 2,
    "EMBASSY/HIGH CO": 22,
    "Govt./MDA (LG)": 23,
    "EX-Staff": 24,
    "Enterprise": 3,
    "Govt./MDA (FGN)": 4,
    "Internal": 5,
    "Ind. Staff": 6,
    "Govt./MDA (ST)": 7,
    "DSE": 8,
    "Trainee": 9,
}
```
- Extract `customerSubType` from bioData
- Look up in map, use integer value
- If not found or not provided, set to `NULL`

### Step 11: Prepare Customer Profile Record
**Table**: `customer_profile`

**Fields to Set**:
- `customerProfileId`: Generated UUID
- `customerId`: From customer record
- `customerNumber`: Generated from sequence
- `title`: From bioData (optional)
- `firstName`: From bioData (Individual only)
- `surname`: From bioData (Individual only)
- `otherNames`: From bioData (optional)
- `mobileNumber`: From bioData (normalized)
- `alternateMobileNumber`: From bioData (optional)
- `emailAddress`: From bioData (normalized)
- `dateOfBirth`: From bioData, format as "YYYY-MM-DD" (Individual only)
- `companyNameBusiness`: From bioData (SME only)
- `bvn`: From bioData (Individual, normalized)
- `nin`: From bioData (Individual, optional)
- `certificateOfIncorporation`: From bioData (SME, normalized from rcNumber/certificateOfIncorporation)
- `taxIdentificationNumber`: From bioData (SME, normalized)
- `nationality`: From bioData (optional)
- `categoryOfBusiness`: From bioData (optional)
- `dateOfRegistration`: From bioData (SME, format as "YYYY-MM-DD")
- `introducer`: From bioData (optional, UUID)
- `sectorId`: From Step 9 validation
- `industryId`: From Step 9 validation
- `targetId`: From Step 9 validation
- `customerProfileData`: JSONB object containing:
  ```json
  {
    "customerEntityId": "<generated_sequence_value>",
    ...all_bioData_fields,
    "bvn": "<bvn>",
    "idType": "<chooseAnId>",
    "idNumber": "<idNumber>",
    "chooseAnId": "<chooseAnId>",
    "email": "<email>",
    "emailAddress": "<emailAddress>",
    "target": "<targetCode>",
    "industry": "<industryCode>",
    "sector": "<sectorCode>"
  }
  ```

**Note**: Store all original bioData fields in `customerProfileData` JSONB for backward compatibility

### Step 12: Create Database Records
1. **Insert into `customer` table** (within transaction)
2. **Insert into `customer_profile` table** (within transaction)
3. **Insert into `activity_log` table**:
   - `activityLogId`: Generate UUID
   - `description`: `"Customer created via channel direct flow by SYSTEM_USER"`
   - `customerId`: From customer record
   - `createdAt`: Current timestamp
   - `updatedAt`: Current timestamp

### Step 13: Commit Transaction
1. If all operations successful, commit transaction
2. If any error occurs, rollback transaction

### Step 14: Return Response
1. Generate `requestId`: UUID v4 (for response only, not stored)
2. Return success response with `customerId` and `requestId`

## Error Handling

### Error Response Format
```json
{
  "status": "error",
  "message": "<error_message>",
  "responseCode": <http_status_code>
}
```

### Error Scenarios
1. **401 Unauthorized**: Invalid or missing API key
2. **400 Bad Request**: 
   - Missing required fields
   - Invalid customerType
   - Validation errors (BVN format, date format, etc.)
   - Phone number already exists
   - Customer duplicate found
3. **404 Not Found**: Sector/Industry/Target not found
4. **500 Internal Server Error**: Database errors, transaction failures

### Error Messages (for backward compatibility)
- `"Customer data is required"`
- `"Customer type is required"`
- `"The payload is missing the 'branch' field"`
- `"BVN must be a string."`
- `"BVN must contain only numbers."`
- `"BVN must not be redacted (should not contain '*')."`
- `"Date must be a valid date string in ISO 8601 or RFC 2822 format."`
- `"Customer with Phone number already exists"`
- `"Customer with BVN and NIN already exists"`
- `"Customer with details passed already exists"`
- `"Customer with companyNameBusiness and taxIdentificationNumber and dateOfRegistration already exists"`
- `"Sector not found: {code}"`
- `"Industry not found: {code}"`
- `"Target not found: {code}"`

## Database Schema Reference

### customer table
- Primary Key: `customerId` (UUID)
- Required Fields: `customerId`, `customerType`, `status`, `initiator`, `initiatorId`, `branch`
- Foreign Keys: `relatedEntity` → other tables

### customer_profile table
- Primary Key: `customerProfileId` (UUID)
- Foreign Key: `customerId` → `customer.customerId`
- Foreign Keys: `sectorId` → `sector.sectorId`, `industryId` → `industry.industryId`, `targetId` → `target.targetId`
- JSONB Field: `customerProfileData`

### customer_profile_lookup table
- Primary Key: `customerProfileLookupId` (UUID, auto-generated)
- Unique Constraints:
  - `(firstName, surname, bvn, dateOfBirth)` for Individual
  - `(firstName, surname, nin, dateOfBirth)` for Individual
  - `(companyNameBusiness, certificateOfIncorporation, taxIdentificationNumber, dateOfRegistration)` for SME
  - `(companyNameBusiness)` for SME (partial index)

### activity_log table
- Primary Key: `activityLogId` (UUID)
- Foreign Key: `customerId` → `customer.customerId`

## PostgreSQL Sequences

### Create Sequences (if not exists)
```sql
-- For customerNumber
CREATE SEQUENCE IF NOT EXISTS customer_number_seq START 1;

-- For customerEntityId  
CREATE SEQUENCE IF NOT EXISTS customer_entity_id_seq START 1;
```

### Usage
```sql
-- Get next customer number
SELECT nextval('customer_number_seq')::text;

-- Get next entity ID
SELECT nextval('customer_entity_id_seq')::text;
```

## Implementation Notes

1. **Transaction Management**: Use PostgreSQL transactions to ensure atomicity of customer + customer_profile + activity_log creation
2. **Field Name Variations**: Support all case variations and field name aliases as specified
3. **Date Formatting**: Convert dates to "YYYY-MM-DD" format for storage
4. **UUID Generation**: Use standard UUID v4 for all ID fields except customerNumber and customerEntityId
5. **Null Handling**: Remove null/empty values from bioData before processing
6. **Case Sensitivity**: Use case-insensitive matching for company names, tax IDs, and sector/industry/target lookups
7. **Age Calculation**: Calculate age from dateOfBirth for phone validation (only check if age > 18)
8. **Race Condition Handling**: Handle concurrent requests that might create duplicates between lookup and insert

## Testing Considerations

1. Test with all field name variations
2. Test duplicate detection (BVN, NIN, phone, company name)
3. Test transaction rollback on errors
4. Test sector/industry/target validation failures
5. Test API key authentication
6. Test required field validation
7. Test date format validation
8. Test phone number validation (age > 18 check)
9. Test race conditions (concurrent duplicate requests)

