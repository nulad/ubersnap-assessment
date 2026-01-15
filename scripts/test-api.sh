#!/bin/bash

# Flash Sale Coupon System - API Test Script
# This script demonstrates all API endpoints with expected responses

set -e

API_BASE="http://localhost:8080/api"
COUPON_NAME="TEST_DEMO_$(date +%s)"

echo "======================================"
echo "Flash Sale Coupon System - API Demo"
echo "======================================"
echo ""

# Check if API is running
echo "[1/6] Checking API availability..."
if ! curl -s "${API_BASE}/coupons/test" > /dev/null 2>&1; then
    echo "❌ Error: API not reachable at ${API_BASE}"
    echo "   Please ensure the server is running: docker-compose up"
    exit 1
fi
echo "✓ API is running"
echo ""

# 1. Create a coupon
echo "[2/6] Creating coupon '${COUPON_NAME}' with 10 units..."
CREATE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/coupons" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${COUPON_NAME}\",\"amount\":10}")

HTTP_STATUS=$(echo "$CREATE_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
RESPONSE_BODY=$(echo "$CREATE_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$HTTP_STATUS" = "201" ]; then
    echo "✓ Created successfully (HTTP 201)"
else
    echo "❌ Failed with HTTP ${HTTP_STATUS}"
    echo "   Response: ${RESPONSE_BODY}"
    exit 1
fi
echo ""

# 2. Claim with user1
echo "[3/6] Claiming coupon with user1..."
CLAIM1_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/coupons/claim" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"user1\",\"coupon_name\":\"${COUPON_NAME}\"}")

HTTP_STATUS=$(echo "$CLAIM1_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
RESPONSE_BODY=$(echo "$CLAIM1_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ Claimed successfully (HTTP 200)"
    echo "   Response: ${RESPONSE_BODY}"
else
    echo "❌ Failed with HTTP ${HTTP_STATUS}"
    echo "   Response: ${RESPONSE_BODY}"
    exit 1
fi
echo ""

# 3. Try to claim again with user1 (should fail with 409)
echo "[4/6] Attempting double claim with user1 (should fail)..."
CLAIM2_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/coupons/claim" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"user1\",\"coupon_name\":\"${COUPON_NAME}\"}")

HTTP_STATUS=$(echo "$CLAIM2_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
RESPONSE_BODY=$(echo "$CLAIM2_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$HTTP_STATUS" = "409" ]; then
    echo "✓ Correctly rejected double claim (HTTP 409)"
    echo "   Response: ${RESPONSE_BODY}"
else
    echo "❌ Unexpected status: HTTP ${HTTP_STATUS}"
    echo "   Expected: 409 Conflict"
    echo "   Response: ${RESPONSE_BODY}"
    exit 1
fi
echo ""

# 4. Claim with user2
echo "[5/6] Claiming coupon with user2..."
CLAIM3_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/coupons/claim" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"user2\",\"coupon_name\":\"${COUPON_NAME}\"}")

HTTP_STATUS=$(echo "$CLAIM3_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
RESPONSE_BODY=$(echo "$CLAIM3_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ Claimed successfully (HTTP 200)"
    echo "   Response: ${RESPONSE_BODY}"
else
    echo "❌ Failed with HTTP ${HTTP_STATUS}"
    echo "   Response: ${RESPONSE_BODY}"
    exit 1
fi
echo ""

# 5. Get coupon details
echo "[6/6] Fetching coupon details..."
GET_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET "${API_BASE}/coupons/${COUPON_NAME}")

HTTP_STATUS=$(echo "$GET_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
RESPONSE_BODY=$(echo "$GET_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ Retrieved successfully (HTTP 200)"
    echo "   Response:"
    echo "$RESPONSE_BODY" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE_BODY"
else
    echo "❌ Failed with HTTP ${HTTP_STATUS}"
    echo "   Response: ${RESPONSE_BODY}"
    exit 1
fi
echo ""

echo "======================================"
echo "✓ All API tests passed successfully!"
echo "======================================"
echo ""
echo "Summary:"
echo "  - Created coupon: ${COUPON_NAME} with 10 units"
echo "  - user1 claimed: ✓"
echo "  - user1 double claim: Rejected (409)"
echo "  - user2 claimed: ✓"
echo "  - Remaining: 8 units"
echo ""
