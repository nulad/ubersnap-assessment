#!/bin/bash

# Flash Sale Coupon System - Automated Test Runner
# This script starts the services, runs all tests, and reports results

set -e

echo "======================================"
echo "Running Automated Test Suite"
echo "======================================"
echo ""

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Error: docker-compose is not installed"
    echo "   Please install Docker Desktop which includes docker-compose"
    exit 1
fi

# Clean up any existing containers
echo "[1/5] Cleaning up existing containers..."
docker-compose down -v > /dev/null 2>&1 || true
echo "✓ Cleanup complete"
echo ""

# Start services in detached mode
echo "[2/5] Starting services (docker-compose up -d --build)..."
docker-compose up -d --build

if [ $? -ne 0 ]; then
    echo "❌ Failed to start services"
    exit 1
fi
echo "✓ Services started"
echo ""

# Wait for services to be healthy
echo "[3/5] Waiting for services to be ready..."
MAX_WAIT=60
WAITED=0
INTERVAL=2

while [ $WAITED -lt $MAX_WAIT ]; do
    # Check if API is responding
    if curl -s http://localhost:8080/api/coupons/healthcheck > /dev/null 2>&1; then
        echo "✓ API is ready"
        break
    fi

    # Alternative: check if postgres is ready
    if docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
        # Wait a bit more for API to start
        sleep 3
        if curl -s http://localhost:8080/api/coupons/healthcheck > /dev/null 2>&1 || curl -s http://localhost:8080/ > /dev/null 2>&1; then
            echo "✓ Services are ready"
            break
        fi
    fi

    echo "   Waiting for services... (${WAITED}s/${MAX_WAIT}s)"
    sleep $INTERVAL
    WAITED=$((WAITED + INTERVAL))
done

if [ $WAITED -ge $MAX_WAIT ]; then
    echo "❌ Services did not start within ${MAX_WAIT} seconds"
    echo "   Showing logs:"
    docker-compose logs --tail=50
    docker-compose down
    exit 1
fi
echo ""

# Run integration tests from host machine
echo "[4/5] Running integration tests..."
echo "----------------------------------------"

# Check if Go is installed on host
if ! command -v go &> /dev/null; then
    echo "⚠️  Warning: Go is not installed on the host machine."
    echo "   Integration tests require Go to run from the host."
    echo ""
    echo "   Alternative: Run the manual API test script instead:"
    echo "   ./scripts/test-api.sh"
    echo ""
    TEST_EXIT=1
else
    # Run tests from the host against the running API
    go test ./tests/... -v
    TEST_EXIT=$?
fi

echo "----------------------------------------"
echo ""

# Report results
echo "[5/5] Test Results:"
if [ $TEST_EXIT -eq 0 ]; then
    echo "✓ All tests PASSED"
    echo ""
    echo "======================================"
    echo "Test Suite: SUCCESS"
    echo "======================================"
else
    echo "❌ Some tests FAILED"
    echo ""
    echo "======================================"
    echo "Test Suite: FAILURE"
    echo "======================================"
fi
echo ""

# Optional: Keep services running or tear down
echo "Services are still running. To view logs:"
echo "  docker-compose logs -f"
echo ""
echo "To stop services:"
echo "  docker-compose down"
echo ""

exit $TEST_EXIT
