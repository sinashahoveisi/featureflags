#!/bin/sh

set -e

echo "Running FeatureFlags Service Tests"
echo "=================================="

# Set test environment variables
export TEST_DB_HOST=${TEST_DB_HOST:-db}
export TEST_DB_PORT=${TEST_DB_PORT:-5432}
export TEST_DB_USER=${TEST_DB_USER:-featureflags}
export TEST_DB_PASSWORD=${TEST_DB_PASSWORD:-featureflags}
export POSTGRES_DB=${POSTGRES_DB:-featureflags}
export TEST_DB_NAME=${TEST_DB_NAME:-${POSTGRES_DB}_test}

# Wait for database to be ready
echo "Waiting for test database to be ready..."
until pg_isready -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER; do
  echo "Database is unavailable - sleeping"
  sleep 2
done

echo "Database is ready!"

# Create test database if it doesn't exist
echo "Creating test database if it doesn't exist..."
PGPASSWORD=$TEST_DB_PASSWORD psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$TEST_DB_NAME'" | grep -q 1 || PGPASSWORD=$TEST_DB_PASSWORD psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER -c "CREATE DATABASE $TEST_DB_NAME"

# Run tests
echo "Running unit and integration tests..."
go test -v -coverprofile=coverage.out ./service/...

# Generate coverage report
echo "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "Tests completed successfully!"
echo "Coverage report saved to coverage.html" 