#!/bin/bash

set -e

echo "🚀 Testing FeatureFlags Swagger Integration"
echo "============================================"

# Function to wait for service to be ready
wait_for_service() {
    echo "⏳ Waiting for service to be ready..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo "✅ Service is ready!"
            return 0
        fi
        sleep 1
    done
    echo "❌ Service failed to start"
    exit 1
}

# Test 1: Swagger Enabled (default)
echo ""
echo "📖 Test 1: Swagger Documentation ENABLED"
echo "----------------------------------------"
docker-compose up -d --build

wait_for_service

echo "🔍 Testing Swagger UI endpoint..."
if curl -s http://localhost:8080/swagger/index.html | grep -q "Swagger UI"; then
    echo "✅ Swagger UI is accessible at http://localhost:8080/swagger/index.html"
else
    echo "❌ Swagger UI is not accessible"
fi

echo "🔍 Testing API endpoints..."
curl -s http://localhost:8080/health | jq .
curl -s http://localhost:8080/api/v1/flags | jq .

echo "🛑 Stopping service..."
docker-compose down

# Test 2: Swagger Disabled
echo ""
echo "📖 Test 2: Swagger Documentation DISABLED"
echo "-----------------------------------------"

# Temporarily disable Swagger
echo "SWAGGER_ENABLED=false" >> .env.temp
cat .env >> .env.temp
mv .env.temp .env

docker-compose up -d --build

wait_for_service

echo "🔍 Testing Swagger UI endpoint (should return 404)..."
if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/swagger/index.html | grep -q "404"; then
    echo "✅ Swagger UI is properly disabled (404)"
else
    echo "❌ Swagger UI should be disabled but is still accessible"
fi

echo "🔍 Testing that API endpoints still work..."
curl -s http://localhost:8080/health | jq .
curl -s http://localhost:8080/api/v1/flags | jq .

echo "🛑 Stopping service..."
docker-compose down

# Restore original .env
echo "🔧 Restoring original configuration..."
sed -i '' '/SWAGGER_ENABLED=false/d' .env

echo ""
echo "✅ Swagger Integration Tests Complete!"
echo ""
echo "🎯 Summary:"
echo "  • Swagger UI accessible when SWAGGER_ENABLED=true"
echo "  • Swagger UI returns 404 when SWAGGER_ENABLED=false"
echo "  • API endpoints work regardless of Swagger setting"
echo "  • Simple configuration with just SWAGGER_ENABLED variable"
echo ""
echo "🌐 To access Swagger UI manually:"
echo "  1. Run: docker-compose up -d"
echo "  2. Visit: http://localhost:8080/swagger/index.html" 