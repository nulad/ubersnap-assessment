#!/bin/bash

# Flash Sale Coupon System - API Test Script
# This script demonstrates all API endpoints with expected responses

set -e

API_BASE="http://localhost:8080/api"
COUPON_NAME="TEST_DEMO_$(date +%s)"

# Test result tracking
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Test results storage
CREATE_RESULT=""
CLAIM1_RESULT=""
CLAIM2_RESULT=""
CLAIM3_RESULT=""
GET_RESULT=""

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

TESTS_TOTAL=$((TESTS_TOTAL + 1))
if [ "$HTTP_STATUS" = "201" ]; then
    echo "✓ Created successfully (HTTP 201)"
    CREATE_RESULT="✓ PASS"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo "❌ Failed with HTTP ${HTTP_STATUS}"
    echo "   Response: ${RESPONSE_BODY}"
    CREATE_RESULT="✗ FAIL (HTTP ${HTTP_STATUS})"
    TESTS_FAILED=$((TESTS_FAILED + 1))
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

TESTS_TOTAL=$((TESTS_TOTAL + 1))
if [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ Claimed successfully (HTTP 200)"
    echo "   Response: ${RESPONSE_BODY}"
    CLAIM1_RESULT="✓ PASS"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo "❌ Failed with HTTP ${HTTP_STATUS}"
    echo "   Response: ${RESPONSE_BODY}"
    CLAIM1_RESULT="✗ FAIL (HTTP ${HTTP_STATUS})"
    TESTS_FAILED=$((TESTS_FAILED + 1))
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

TESTS_TOTAL=$((TESTS_TOTAL + 1))
if [ "$HTTP_STATUS" = "409" ]; then
    echo "✓ Correctly rejected double claim (HTTP 409)"
    echo "   Response: ${RESPONSE_BODY}"
    CLAIM2_RESULT="✓ PASS"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo "❌ Unexpected status: HTTP ${HTTP_STATUS}"
    echo "   Expected: 409 Conflict"
    echo "   Response: ${RESPONSE_BODY}"
    CLAIM2_RESULT="✗ FAIL (HTTP ${HTTP_STATUS})"
    TESTS_FAILED=$((TESTS_FAILED + 1))
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

TESTS_TOTAL=$((TESTS_TOTAL + 1))
if [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ Claimed successfully (HTTP 200)"
    echo "   Response: ${RESPONSE_BODY}"
    CLAIM3_RESULT="✓ PASS"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo "❌ Failed with HTTP ${HTTP_STATUS}"
    echo "   Response: ${RESPONSE_BODY}"
    CLAIM3_RESULT="✗ FAIL (HTTP ${HTTP_STATUS})"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    exit 1
fi
echo ""

# 5. Get coupon details
echo "[6/6] Fetching coupon details..."
GET_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET "${API_BASE}/coupons/${COUPON_NAME}")

HTTP_STATUS=$(echo "$GET_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
RESPONSE_BODY=$(echo "$GET_RESPONSE" | sed '/HTTP_STATUS/d')

TESTS_TOTAL=$((TESTS_TOTAL + 1))
if [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ Retrieved successfully (HTTP 200)"
    echo "   Response:"
    echo "$RESPONSE_BODY" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE_BODY"
    GET_RESULT="✓ PASS"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo "❌ Failed with HTTP ${HTTP_STATUS}"
    echo "   Response: ${RESPONSE_BODY}"
    GET_RESULT="✗ FAIL (HTTP ${HTTP_STATUS})"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    exit 1
fi
echo ""

echo "======================================"
if [ $TESTS_FAILED -eq 0 ]; then
    echo "✓ All API tests passed successfully!"
else
    echo "✗ Some tests failed!"
fi
echo "======================================"
echo ""
echo "Test Summary:"
echo "  Total Tests:  ${TESTS_TOTAL}"
echo "  Passed:       ${TESTS_PASSED}"
echo "  Failed:       ${TESTS_FAILED}"
echo ""
echo "Detailed Results:"
echo "  1. Create coupon:      ${CREATE_RESULT}"
echo "  2. user1 claim:        ${CLAIM1_RESULT}"
echo "  3. user1 double claim: ${CLAIM2_RESULT}"
echo "  4. user2 claim:        ${CLAIM3_RESULT}"
echo "  5. Get coupon details: ${GET_RESULT}"
echo ""

# ========================================
# ATTACK SCENARIO TESTS
# ========================================

echo "======================================"
echo "Attack Scenario Tests"
echo "======================================"
echo ""

# Flash Sale Attack Test
echo "[Flash Sale Attack] 50 concurrent requests for coupon with 5 units..."
FLASH_COUPON_NAME="FLASH_SALE_$(date +%s)"

# Create coupon with 5 units
FLASH_CREATE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/coupons" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${FLASH_COUPON_NAME}\",\"amount\":5}")

FLASH_HTTP_STATUS=$(echo "$FLASH_CREATE_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)

if [ "$FLASH_HTTP_STATUS" != "201" ]; then
    echo "❌ Failed to create flash sale coupon"
    exit 1
fi

# Launch 50 concurrent requests
SUCCESS_COUNT=0
FAIL_COUNT=0
TEMP_DIR=$(mktemp -d)

for i in $(seq 1 50); do
    (
        RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/coupons/claim" \
          -H "Content-Type: application/json" \
          -d "{\"user_id\":\"flash_user_${i}\",\"coupon_name\":\"${FLASH_COUPON_NAME}\"}")
        
        HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
        echo "$HTTP_CODE" > "${TEMP_DIR}/result_${i}.txt"
    ) &
done

# Wait for all background processes
wait

# Count results
for i in $(seq 1 50); do
    if [ -f "${TEMP_DIR}/result_${i}.txt" ]; then
        CODE=$(cat "${TEMP_DIR}/result_${i}.txt")
        if [ "$CODE" = "200" ] || [ "$CODE" = "201" ]; then
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        else
            FAIL_COUNT=$((FAIL_COUNT + 1))
        fi
    fi
done

rm -rf "$TEMP_DIR"

# Verify final state
GET_FLASH_RESPONSE=$(curl -s -X GET "${API_BASE}/coupons/${FLASH_COUPON_NAME}")
REMAINING=$(echo "$GET_FLASH_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('remaining_amount', 0))" 2>/dev/null || echo "0")

echo "  Results: ${SUCCESS_COUNT} successful, ${FAIL_COUNT} failed"
echo "  Remaining units: ${REMAINING}"

if [ "$SUCCESS_COUNT" = "5" ] && [ "$FAIL_COUNT" = "45" ] && [ "$REMAINING" = "0" ]; then
    echo "✓ Flash Sale Attack test PASSED"
    FLASH_SALE_RESULT="✓ PASS"
else
    echo "✗ Flash Sale Attack test FAILED"
    echo "  Expected: 5 success, 45 fail, 0 remaining"
    echo "  Got: ${SUCCESS_COUNT} success, ${FAIL_COUNT} fail, ${REMAINING} remaining"
    FLASH_SALE_RESULT="✗ FAIL"
fi
echo ""

# Double Dip Attack Test
echo "[Double Dip Attack] 10 concurrent requests from same user..."
DOUBLE_DIP_COUPON_NAME="DOUBLE_DIP_$(date +%s)"

# Create coupon with 100 units
DOUBLE_DIP_CREATE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/coupons" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${DOUBLE_DIP_COUPON_NAME}\",\"amount\":100}")

DOUBLE_DIP_HTTP_STATUS=$(echo "$DOUBLE_DIP_CREATE_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)

if [ "$DOUBLE_DIP_HTTP_STATUS" != "201" ]; then
    echo "❌ Failed to create double dip coupon"
    exit 1
fi

# Launch 10 concurrent requests from the SAME user
DD_SUCCESS_COUNT=0
DD_CONFLICT_COUNT=0
DD_OTHER_COUNT=0
TEMP_DIR_DD=$(mktemp -d)

for i in $(seq 1 10); do
    (
        RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/coupons/claim" \
          -H "Content-Type: application/json" \
          -d "{\"user_id\":\"double_dip_user\",\"coupon_name\":\"${DOUBLE_DIP_COUPON_NAME}\"}")
        
        HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
        echo "$HTTP_CODE" > "${TEMP_DIR_DD}/result_${i}.txt"
    ) &
done

# Wait for all background processes
wait

# Count results
for i in $(seq 1 10); do
    if [ -f "${TEMP_DIR_DD}/result_${i}.txt" ]; then
        CODE=$(cat "${TEMP_DIR_DD}/result_${i}.txt")
        if [ "$CODE" = "200" ] || [ "$CODE" = "201" ]; then
            DD_SUCCESS_COUNT=$((DD_SUCCESS_COUNT + 1))
        elif [ "$CODE" = "409" ]; then
            DD_CONFLICT_COUNT=$((DD_CONFLICT_COUNT + 1))
        else
            DD_OTHER_COUNT=$((DD_OTHER_COUNT + 1))
        fi
    fi
done

rm -rf "$TEMP_DIR_DD"

# Verify final state
GET_DD_RESPONSE=$(curl -s -X GET "${API_BASE}/coupons/${DOUBLE_DIP_COUPON_NAME}")
DD_REMAINING=$(echo "$GET_DD_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('remaining_amount', 0))" 2>/dev/null || echo "0")
DD_CLAIMED_COUNT=$(echo "$GET_DD_RESPONSE" | python3 -c "import sys, json; print(len(json.load(sys.stdin).get('claimed_by', [])))" 2>/dev/null || echo "0")

echo "  Results: ${DD_SUCCESS_COUNT} success, ${DD_CONFLICT_COUNT} conflict, ${DD_OTHER_COUNT} other"
echo "  Remaining units: ${DD_REMAINING}"
echo "  Claimed by users: ${DD_CLAIMED_COUNT}"

if [ "$DD_SUCCESS_COUNT" = "1" ] && [ "$DD_CONFLICT_COUNT" = "9" ] && [ "$DD_OTHER_COUNT" = "0" ] && [ "$DD_REMAINING" = "99" ] && [ "$DD_CLAIMED_COUNT" = "1" ]; then
    echo "✓ Double Dip Attack test PASSED"
    DOUBLE_DIP_RESULT="✓ PASS"
else
    echo "✗ Double Dip Attack test FAILED"
    echo "  Expected: 1 success, 9 conflict, 0 other, 99 remaining, 1 claimed"
    echo "  Got: ${DD_SUCCESS_COUNT} success, ${DD_CONFLICT_COUNT} conflict, ${DD_OTHER_COUNT} other, ${DD_REMAINING} remaining, ${DD_CLAIMED_COUNT} claimed"
    DOUBLE_DIP_RESULT="✗ FAIL"
fi
echo ""

# Final Summary
echo "======================================"
echo "Final Test Summary"
echo "======================================"
echo ""
echo "Basic API Tests:"
echo "  1. Create coupon:      ${CREATE_RESULT}"
echo "  2. user1 claim:        ${CLAIM1_RESULT}"
echo "  3. user1 double claim: ${CLAIM2_RESULT}"
echo "  4. user2 claim:        ${CLAIM3_RESULT}"
echo "  5. Get coupon details: ${GET_RESULT}"
echo ""
echo "Attack Scenario Tests:"
echo "  6. Flash Sale Attack:  ${FLASH_SALE_RESULT}"
echo "  7. Double Dip Attack:  ${DOUBLE_DIP_RESULT}"
echo ""
