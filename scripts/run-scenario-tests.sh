#!/bin/sh

set -e

echo "🧪 Running FeatureFlags Scenario Tests"
echo "======================================"

# Set test environment variables
export TEST_DB_HOST=${TEST_DB_HOST:-db}
export TEST_DB_PORT=${TEST_DB_PORT:-5432}
export TEST_DB_USER=${TEST_DB_USER:-featureflags}
export TEST_DB_PASSWORD=${TEST_DB_PASSWORD:-featureflags}
export POSTGRES_DB=${POSTGRES_DB:-featureflags}
export TEST_DB_NAME=${TEST_DB_NAME:-${POSTGRES_DB}_test}

# Wait for database to be ready
echo "⏳ Waiting for test database to be ready..."
until pg_isready -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER; do
  echo "Database is unavailable - sleeping"
  sleep 2
done

echo "✅ Database is ready!"

# Create test database if it doesn't exist
echo "🔧 Creating test database if it doesn't exist..."
PGPASSWORD=$TEST_DB_PASSWORD psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$TEST_DB_NAME'" | grep -q 1 || PGPASSWORD=$TEST_DB_PASSWORD psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER -c "CREATE DATABASE $TEST_DB_NAME"

echo ""
echo "🎯 Running Example Scenario Tests"
echo "================================="

# Run all scenario tests
echo "Running scenario integration tests..."
go test -v -run "TestExampleScenario" ./test/

echo ""
echo "🔄 Running Complex Integration Tests"
echo "==================================="

# Run complex integration test
go test -v -run "TestComplexScenarioIntegration" ./test/

echo ""
echo "⚡ Running Service Unit Tests"
echo "============================"

# Run service layer tests
go test -v -coverprofile=coverage.out ./service/...

echo ""
echo "📊 Generating Coverage Report"
echo "============================"

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out

echo ""
echo "✅ All Scenario Tests Completed Successfully!"
echo ""
echo "📋 Test Summary:"
echo "  ✅ Example Scenario 1: checkout_v2 dependencies"
echo "  ✅ Example Scenario 2: Missing dependency error format" 
echo "  ✅ Example Scenario 3: Cascading disable"
echo "  ✅ Example Scenario 4: Circular dependency detection"
echo "  ✅ Complex Integration: Multi-service dependency tree"
echo "  ✅ Service Unit Tests: Business logic validation"
echo ""
echo "📁 Generated Files:"
echo "  - coverage.out: Coverage data"
echo "  - coverage.html: HTML coverage report"
echo ""
echo "🎉 Ready for production deployment!" 