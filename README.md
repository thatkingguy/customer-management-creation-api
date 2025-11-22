# Customer Management Creation API

A production-ready Golang microservice for customer creation operations using Gin framework and PostgreSQL.

## Features

- RESTful API for customer creation
- PostgreSQL database with connection pooling
- Transaction management for data consistency
- Comprehensive validation and error handling
- Structured logging with request ID tracking
- API key authentication
- Health check endpoints
- Graceful shutdown
- Docker support
- Database migrations

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Docker and Docker Compose (for containerized deployment)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd customer-management-creation-api
```

2. Install dependencies:
```bash
make deps
```

3. Copy environment file:
```bash
cp .env.example .env
```

4. Update `.env` with your configuration:
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=your_database

# API Key
API_KEY=your-secure-api-key
```

## Running Locally

### Using Docker Compose (Recommended)

```bash
make docker-run
```

This will start both PostgreSQL and the application in containers.

### Manual Setup

1. Start PostgreSQL database

2. Run migrations:
```bash
make migrate-up
```

3. Run the application:
```bash
make run
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check
- `GET /health` - Basic health check
- `GET /ready` - Readiness check (includes database connection)

### Customer Creation
- `POST /channel/directCreateCustomer` - Create a new customer
  - Requires: `X-API-Key` header
  - See API documentation for request/response format

## Building

```bash
make build
```

## Testing

```bash
make test
make test-coverage
```

## Code Quality

```bash
make fmt    # Format code
make lint   # Run linter
```

## Docker

Build Docker image:
```bash
make docker-build
```

Run with Docker Compose:
```bash
make docker-run
```

Stop containers:
```bash
make docker-stop
```

## Project Structure

```
customer-service/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and migrations
│   ├── models/          # Data models
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic
│   ├── handler/         # HTTP handlers and middleware
│   ├── dto/             # Data transfer objects
│   ├── utils/           # Utility functions
│   └── logger/          # Logging setup
├── pkg/postgres/        # PostgreSQL utilities
├── deployments/         # Docker and deployment configs
└── scripts/             # Utility scripts
```

## Environment Variables

See `.env.example` for all available configuration options.

### Logging Configuration

The logging level can be controlled via the `LOG_LEVEL` environment variable:

- `debug` - Verbose logging (all logs including DEBUG level)
- `info` - Standard logging (INFO, WARN, ERROR levels only)
- `warn` - Warning and error logs only
- `error` - Error logs only

**Default behavior:**
- If `LOG_LEVEL` is set, it takes precedence
- If not set, defaults based on `ENV`/`NODE_ENV`:
  - `development`/`dev` → `debug`
  - `staging`/`qa`/`test` → `info`
  - `production`/`prod` → `info`

**Example:**
```bash
# Set explicit log level
LOG_LEVEL=info

# Or rely on environment-based defaults
ENV=production  # Will use 'info' level
```

## License

[Your License Here]



