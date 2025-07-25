# FeatureFlags Service

A robust backend service for managing feature flags with dependency support, audit logging, and circular dependency detection.

## Features

- ✅ **Feature Flag Management**: Create, enable, disable, and list feature flags
- ✅ **Dependency Support**: Flags can depend on other flags; dependent flags can only be enabled if all dependencies are active
- ✅ **Circular Dependency Detection**: Prevents creation of circular dependencies
- ✅ **Cascading Disables**: When a flag is disabled, all dependent flags are automatically disabled
- ✅ **Comprehensive Audit Logging**: Track all operations with timestamps, actors, and reasons
- ✅ **Graceful Shutdown**: Clean shutdown with configurable timeout
- ✅ **Structured Logging**: JSON-structured logs with configurable levels
- ✅ **Input Validation**: Comprehensive request validation with detailed error messages
- ✅ **Database Migrations**: Automated database schema management
- ✅ **Docker Support**: Fully containerized with Docker Compose
- ✅ **Comprehensive Tests**: Unit and integration tests with coverage reporting
- ✅ **Swagger Documentation**: Interactive API documentation with configurable enable/disable

## Architecture

The project follows a clean architecture pattern with clear separation of concerns:

```
featureflags/
├── cmd/                    # Application entry point
├── config/                 # Configuration management
├── entity/                 # Domain models
├── repository/             # Data access layer
├── service/                # Business logic layer
├── controller/             # HTTP handlers
├── handler/                # Route registration
├── validator/              # Input validation
├── test/                   # Test helpers
├── migrations/             # Database migrations
├── pkg/
│   └── logger/            # Structured logging
├── scripts/               # Utility scripts
└── docker-compose.yml     # Container orchestration
```

## API Endpoints

### Health Check
- `GET /health` - Service health status

### Documentation
- `GET /swagger/index.html` - Interactive Swagger API documentation (if enabled)

### Flag Management
- `POST /api/v1/flags` - Create a new flag
- `GET /api/v1/flags` - List all flags
- `GET /api/v1/flags/:id` - Get a specific flag
- `POST /api/v1/flags/:id/toggle` - Enable/disable a flag
- `GET /api/v1/flags/:id/audit` - Get audit logs for a flag

## Example API Usage

### Create a Flag
```bash
curl -X POST http://localhost:8080/api/v1/flags \
  -H "Content-Type: application/json" \
  -H "X-Actor: user123" \
  -d '{
    "name": "checkout_v2",
    "dependencies": [1, 2]
  }'
```

### Enable a Flag
```bash
curl -X POST http://localhost:8080/api/v1/flags/1/toggle \
  -H "Content-Type: application/json" \
  -H "X-Actor: user123" \
  -d '{
    "enable": true,
    "reason": "Ready for production"
  }'
```

### Error Response for Missing Dependencies
```json
{
  "error": "Missing active dependencies",
  "missing_dependencies": ["auth_v2", "user_profile_v2"]
}
```

## Configuration

The service supports configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_SERVER_PORT` | `8080` | HTTP server port |
| `DATABASE_HOST` | `db` | PostgreSQL host |
| `DATABASE_PORT` | `5432` | PostgreSQL port |
| `DATABASE_USER` | `featureflags` | Database user |
| `DATABASE_PASSWORD` | `featureflags` | Database password |
| `DATABASE_NAME` | `featureflags` | Database name |
| `LOGGER_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOGGER_MODE` | `production` | Log mode (development, production) |
| `APPLICATION_GRACEFUL_SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |
| `SWAGGER_ENABLED` | `true` | Enable/disable Swagger documentation |

## Running the Service

### Using Docker Compose (Recommended)

1. **Start the service:**
   ```bash
   docker-compose up --build
   ```

2. **Run tests:**
   ```bash
   docker-compose run --rm test
   ```

3. **Stop the service:**
   ```bash
   docker-compose down
   ```

### Development with Hot Reload

For the best development experience, use the development environment with hot reload:

```bash
# Start development environment (recommended)
make dev
# or
./scripts/dev.sh

# The service will be available at:
# - API: http://localhost:8080
# - Swagger: http://localhost:8080/swagger/index.html
```

**Development Features:**
- 🔥 **Hot Reload**: Automatic restart on code changes using [Air](https://github.com/cosmtrek/air)
- 🐛 **Debug Logging**: Enhanced logging for development
- 📚 **Swagger Enabled**: Interactive API documentation
- 🗄️ **PostgreSQL Included**: Database automatically configured
- 📁 **Live Code Mounting**: Edit code directly, see changes instantly

**Development Commands:**
```bash
make dev          # Start development environment
make dev-bg       # Start in background
make dev-logs     # View logs
make dev-stop     # Stop development environment
make dev-test     # Run tests in dev environment
```

### Local Development (Alternative)

If you prefer to run without Docker:

1. **Prerequisites:**
   - Go 1.23+
   - PostgreSQL 15+
   - Air (for hot reload): `make install-air`

2. **Setup database:**
   ```bash
   createdb featureflags
   ```

3. **Run with hot reload:**
   ```bash
   make air-local
   ```

4. **Or run normally:**
   ```bash
   make run
   ```

## Testing

The project includes comprehensive automated tests covering all core functionality and example scenarios from the requirements:

### Running Tests

```bash
# Run all tests with Docker Compose
docker-compose run --rm test

# Run example scenario tests
docker-compose run --rm scenario-test

# Run tests locally (requires PostgreSQL)
./scripts/test.sh

# Run scenario tests locally
./scripts/run-scenario-tests.sh

# Run specific test suites
go test -v ./service/...                    # Service layer unit tests
go test -v ./test/...                       # Integration tests
go test -v -run "TestExampleScenario" ./test/  # Example scenarios only
```

### Example Scenarios Tested

The test suite validates all example scenarios from the requirements:

1. **Scenario 1: Dependency Validation**
   - `checkout_v2` depends on `auth_v2` and `user_profile_v2`
   - Can only be enabled after both dependencies are active

2. **Scenario 2: Missing Dependency Error Format**
   - Returns exact error format: `{"error": "Missing active dependencies", "missing_dependencies": ["auth_v2"]}`

3. **Scenario 3: Cascading Disable**
   - When `auth_v2` is disabled, automatically disables `checkout_v2` and dependent flags
   - Logs cascading changes with `system` actor and `cascade_disable` action

4. **Scenario 4: Circular Dependency Detection**
   - Prevents creation of flags with circular dependencies
   - Returns clear error messages for circular dependency attempts

5. **Complex Integration Scenario**
   - Multi-service dependency tree with 7+ flags
   - Tests realistic microservice dependency patterns
   - Validates cascading behavior across complex dependency chains

### Test Coverage

Tests cover the following functionality:
- ✅ Flag creation with and without dependencies
- ✅ Dependency validation and missing dependency errors
- ✅ Cascading disable functionality with audit logging
- ✅ Circular dependency detection and prevention
- ✅ Flag toggling and status management
- ✅ Audit log creation and retrieval
- ✅ Input validation and error handling
- ✅ Database operations and transactions
- ✅ HTTP API endpoints and error responses
- ✅ Integration testing with full application stack

### CI/CD Testing

The project includes automated testing in both GitLab CI and GitHub Actions:
- **Unit Tests**: Run on every commit and pull request
- **Integration Tests**: Full application testing with PostgreSQL
- **Scenario Tests**: Validate all requirement examples
- **Security Scanning**: Vulnerability detection with Trivy
- **Coverage Reports**: Generate and upload test coverage
- **Docker Image Testing**: Test built images before deployment

## Database Schema

The service uses PostgreSQL with the following tables:

- **flags**: Store flag information (id, name, status, timestamps)
- **flag_dependencies**: Store flag dependency relationships
- **audit_logs**: Store audit trail of all operations
- **schema_migrations**: Track applied database migrations

## Graceful Shutdown

The service implements graceful shutdown:
- Listens for SIGTERM/SIGINT signals
- Stops accepting new requests
- Completes existing requests within timeout
- Closes database connections cleanly

## Logging

Structured JSON logging with configurable levels:
- **Development mode**: Human-readable format with colors
- **Production mode**: JSON format optimized for log aggregation
- **Request logging**: Automatic HTTP request/response logging
- **Error tracking**: Detailed error context and stack traces

## Example Scenarios

### Scenario 1: Basic Flag Management
```bash
# Create a simple flag
curl -X POST localhost:8080/api/v1/flags \
  -H "Content-Type: application/json" \
  -d '{"name": "simple_feature"}'

# Enable the flag
curl -X POST localhost:8080/api/v1/flags/1/toggle \
  -H "Content-Type: application/json" \
  -d '{"enable": true, "reason": "Feature is ready"}'
```

### Scenario 2: Dependency Management
```bash
# Create base flags
curl -X POST localhost:8080/api/v1/flags \
  -d '{"name": "auth_v2"}'
curl -X POST localhost:8080/api/v1/flags \
  -d '{"name": "user_profile_v2"}'

# Create dependent flag
curl -X POST localhost:8080/api/v1/flags \
  -d '{"name": "checkout_v2", "dependencies": [1, 2]}'

# Try to enable checkout_v2 (will fail - dependencies not enabled)
curl -X POST localhost:8080/api/v1/flags/3/toggle \
  -d '{"enable": true, "reason": "Launch checkout v2"}'
# Response: {"error": "Missing active dependencies", "missing_dependencies": ["auth_v2", "user_profile_v2"]}

# Enable dependencies first
curl -X POST localhost:8080/api/v1/flags/1/toggle \
  -d '{"enable": true, "reason": "Auth v2 ready"}'
curl -X POST localhost:8080/api/v1/flags/2/toggle \
  -d '{"enable": true, "reason": "User profile ready"}'

# Now enable checkout_v2 (will succeed)
curl -X POST localhost:8080/api/v1/flags/3/toggle \
  -d '{"enable": true, "reason": "All dependencies ready"}'
```

### Scenario 3: Cascade Disable
```bash
# Disable auth_v2 (will cascade disable checkout_v2)
curl -X POST localhost:8080/api/v1/flags/1/toggle \
  -d '{"enable": false, "reason": "Auth issues detected"}'

# Check audit logs to see cascade actions
curl localhost:8080/api/v1/flags/3/audit
```

## Development

### Adding New Features

1. **Add entity models** in `entity/`
2. **Implement repository methods** in `repository/`
3. **Add business logic** in `service/`
4. **Create HTTP handlers** in `controller/`
5. **Add validation** in `validator/`
6. **Write tests** in `*_test.go` files
7. **Update migrations** in `migrations/`

### Code Quality

- Follow Go best practices and idioms
- Maintain test coverage above 80%
- Use structured logging for observability
- Implement proper error handling
- Document public APIs

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Swagger Documentation

The service includes interactive Swagger/OpenAPI documentation that can be enabled or disabled via configuration.

### Accessing Swagger UI

When enabled (default), you can access the interactive API documentation at:
- **Local Development**: http://localhost:8080/swagger/index.html
- **Docker**: http://localhost:8080/swagger/index.html

### Configuration

Control Swagger documentation via environment variables:

```bash
# Enable/disable Swagger (default: true)
SWAGGER_ENABLED=true
```

### Disabling Swagger in Production

For production environments, you may want to disable Swagger:

```bash
# In your .env or environment
SWAGGER_ENABLED=false
```

The Swagger documentation includes:
- **Interactive API Testing**: Test endpoints directly from the browser
- **Request/Response Examples**: See example payloads and responses
- **Schema Documentation**: Detailed information about data models
- **Error Response Examples**: Including dependency error formats

## CI/CD Pipeline

The project includes comprehensive CI/CD pipelines for both GitLab CI and GitHub Actions.

### Pipeline Features

- ✅ **Automated Testing**: Unit, integration, and scenario tests
- ✅ **Security Scanning**: Vulnerability detection with Trivy
- ✅ **Docker Image Building**: Multi-platform (amd64/arm64) builds
- ✅ **Container Registry**: Automatic image publishing with git commit hash
- ✅ **Coverage Reports**: Test coverage analysis and reporting
- ✅ **Integration Testing**: Full application testing with Docker Compose
- ✅ **Deployment Automation**: Staging and production deployment workflows

### GitLab CI Pipeline

The `.gitlab-ci.yml` includes:

```yaml
stages:
  - test          # Run tests with PostgreSQL
  - build         # Build Docker images
  - publish       # Push to GitLab Container Registry
```

**Image Tags:**
- `registry.gitlab.com/project/featureflags:latest` (main branch)
- `registry.gitlab.com/project/featureflags:${CI_COMMIT_SHA}` (commit hash)
- `registry.gitlab.com/project/featureflags:${CI_COMMIT_REF_SLUG}` (branch name)

### GitHub Actions Pipeline

The `.github/workflows/ci.yml` includes:

```yaml
jobs:
  - test              # Unit and integration tests
  - build             # Docker image building
  - publish           # Push to GitHub Container Registry
  - security-scan     # Trivy vulnerability scanning
  - integration-test  # Full application testing
  - deploy-staging    # Staging deployment
  - deploy-production # Production deployment
```

**Image Tags:**
- `ghcr.io/username/featureflags:latest` (main branch)
- `ghcr.io/username/featureflags:${GITHUB_SHA}` (commit hash)
- `ghcr.io/username/featureflags:main-${GITHUB_SHA}` (branch-commit)

### Database Configuration for Testing

Both pipelines support the `POSTGRES_DB_test` naming convention:

```bash
# Environment variables
POSTGRES_DB=featureflags          # Base database name
TEST_DB_NAME=featureflags_test    # Test database name (auto-generated)
```

### Running CI/CD Locally

```bash
# Test the full CI pipeline locally
docker-compose run --rm scenario-test

# Test image building
docker build -t featureflags:test .

# Test with specific commit hash
docker build -t featureflags:$(git rev-parse --short HEAD) .
```

## License

This project is licensed under the MIT License. 