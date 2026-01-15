package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	apiBaseURL = "http://localhost:8080/api"
	serverAddr = "127.0.0.1:8080"
)


// Helper function to check if server is running
func isServerRunning(t *testing.T) bool {
	conn, err := net.DialTimeout("tcp", serverAddr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Helper function to wait for server to be ready
func waitForServer(t *testing.T) {
	for i := 0; i < 10; i++ {
		if isServerRunning(t) {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Skipf("Skipping integration test: API server not reachable on %s", serverAddr)
}

// Test cases for POST /api/coupons
func TestCreateCoupon(t *testing.T) {
	waitForServer(t)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		checkError     bool
		errorContains  string
	}{
		{
			name: "Success: Create coupon with valid data",
			payload: map[string]interface{}{
				"name":   fmt.Sprintf("VALID_%d", time.Now().UnixNano()),
				"amount": 100,
			},
			expectedStatus: http.StatusCreated,
			checkError:     false,
		},
		{
			name: "Fail: Invalid amount (zero)",
			payload: map[string]interface{}{
				"name":   fmt.Sprintf("ZERO_%d", time.Now().UnixNano()),
				"amount": 0,
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
			errorContains:  "", // Generic validation error is acceptable
		},
		{
			name: "Fail: Invalid amount (negative)",
			payload: map[string]interface{}{
				"name":   fmt.Sprintf("NEG_%d", time.Now().UnixNano()),
				"amount": -10,
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
			errorContains:  "", // Generic validation error is acceptable
		},
		{
			name: "Fail: Empty name",
			payload: map[string]interface{}{
				"name":   "",
				"amount": 50,
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			resp, err := http.Post(apiBaseURL+"/coupons", "application/json", bytes.NewBuffer(body))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkError {
				var errorResp map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp, "error")
				if tt.errorContains != "" {
					errorMsg := errorResp["error"].(string)
					assert.Contains(t, errorMsg, tt.errorContains)
				}
			}
		})
	}
}

func TestCreateCouponDuplicate(t *testing.T) {
	waitForServer(t)

	// Create a coupon first
	couponName := fmt.Sprintf("DUP_%d", time.Now().UnixNano())
	payload := map[string]interface{}{
		"name":   couponName,
		"amount": 50,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(apiBaseURL+"/coupons", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Try to create the same coupon again
	body, _ = json.Marshal(payload)
	resp, err = http.Post(apiBaseURL+"/coupons", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp, "error")
}

// Test cases for POST /api/coupons/claim
func TestClaimCoupon(t *testing.T) {
	waitForServer(t)

	tests := []struct {
		name           string
		setupCoupon    bool
		couponName     string
		couponAmount   int
		claimUserID    string
		expectedStatus int
		checkError     bool
		errorContains  string
	}{
		{
			name:           "Success: Claim with stock available",
			setupCoupon:    true,
			couponAmount:   10,
			claimUserID:    "user_success",
			expectedStatus: http.StatusOK,
			checkError:     false,
		},
		{
			name:           "Fail: Claim non-existent coupon",
			setupCoupon:    false,
			couponName:     "NONEXISTENT",
			claimUserID:    "user_fail",
			expectedStatus: http.StatusNotFound,
			checkError:     true,
			errorContains:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var couponName string
			if tt.setupCoupon {
				// Create a coupon for this test
				couponName = fmt.Sprintf("CLAIM_%d", time.Now().UnixNano())
				createPayload := map[string]interface{}{
					"name":   couponName,
					"amount": tt.couponAmount,
				}
				body, _ := json.Marshal(createPayload)
				resp, err := http.Post(apiBaseURL+"/coupons", "application/json", bytes.NewBuffer(body))
				require.NoError(t, err)
				defer resp.Body.Close()
				require.Equal(t, http.StatusCreated, resp.StatusCode)
			} else {
				couponName = tt.couponName
			}

			// Claim the coupon
			claimPayload := map[string]interface{}{
				"user_id":     tt.claimUserID,
				"coupon_name": couponName,
			}
			body, _ := json.Marshal(claimPayload)
			resp, err := http.Post(apiBaseURL+"/coupons/claim", "application/json", bytes.NewBuffer(body))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkError {
				var errorResp map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp, "error")
				if tt.errorContains != "" {
					errorMsg := errorResp["error"].(string)
					assert.Contains(t, errorMsg, tt.errorContains)
				}
			}
		})
	}
}

func TestClaimCouponNoStock(t *testing.T) {
	waitForServer(t)

	// Create a coupon with only 1 stock
	couponName := fmt.Sprintf("NOSTOCK_%d", time.Now().UnixNano())
	createPayload := map[string]interface{}{
		"name":   couponName,
		"amount": 1,
	}
	body, _ := json.Marshal(createPayload)
	resp, err := http.Post(apiBaseURL+"/coupons", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// First claim should succeed
	claimPayload := map[string]interface{}{
		"user_id":     "user1",
		"coupon_name": couponName,
	}
	body, _ = json.Marshal(claimPayload)
	resp, err = http.Post(apiBaseURL+"/coupons/claim", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Second claim should fail with no stock
	claimPayload = map[string]interface{}{
		"user_id":     "user2",
		"coupon_name": couponName,
	}
	body, _ = json.Marshal(claimPayload)
	resp, err = http.Post(apiBaseURL+"/coupons/claim", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp, "error")
}

func TestClaimCouponAlreadyClaimed(t *testing.T) {
	waitForServer(t)

	// Create a coupon
	couponName := fmt.Sprintf("ALREADY_%d", time.Now().UnixNano())
	createPayload := map[string]interface{}{
		"name":   couponName,
		"amount": 10,
	}
	body, _ := json.Marshal(createPayload)
	resp, err := http.Post(apiBaseURL+"/coupons", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// First claim should succeed
	userID := "duplicate_user"
	claimPayload := map[string]interface{}{
		"user_id":     userID,
		"coupon_name": couponName,
	}
	body, _ = json.Marshal(claimPayload)
	resp, err = http.Post(apiBaseURL+"/coupons/claim", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Second claim by same user should fail with 409 Conflict
	body, _ = json.Marshal(claimPayload)
	resp, err = http.Post(apiBaseURL+"/coupons/claim", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp, "error")
}

// Test cases for GET /api/coupons/{name}
func TestGetCoupon(t *testing.T) {
	waitForServer(t)

	tests := []struct {
		name           string
		setupCoupon    bool
		couponName     string
		couponAmount   int
		expectedStatus int
		checkData      bool
	}{
		{
			name:           "Success: Get existing coupon",
			setupCoupon:    true,
			couponAmount:   50,
			expectedStatus: http.StatusOK,
			checkData:      true,
		},
		{
			name:           "Fail: Get non-existent coupon",
			setupCoupon:    false,
			couponName:     "NONEXISTENT_GET",
			expectedStatus: http.StatusNotFound,
			checkData:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var couponName string
			if tt.setupCoupon {
				// Create a coupon for this test
				couponName = fmt.Sprintf("GET_%d", time.Now().UnixNano())
				createPayload := map[string]interface{}{
					"name":   couponName,
					"amount": tt.couponAmount,
				}
				body, _ := json.Marshal(createPayload)
				resp, err := http.Post(apiBaseURL+"/coupons", "application/json", bytes.NewBuffer(body))
				require.NoError(t, err)
				defer resp.Body.Close()
				require.Equal(t, http.StatusCreated, resp.StatusCode)
			} else {
				couponName = tt.couponName
			}

			// Get the coupon
			resp, err := http.Get(apiBaseURL + "/coupons/" + couponName)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkData {
				var coupon Coupon
				err := json.NewDecoder(resp.Body).Decode(&coupon)
				require.NoError(t, err)

				assert.Equal(t, couponName, coupon.Name)
				assert.Equal(t, tt.couponAmount, coupon.Amount)
				assert.Equal(t, tt.couponAmount, coupon.RemainingAmount)
				assert.NotNil(t, coupon.ClaimedBy)
				assert.Equal(t, 0, len(coupon.ClaimedBy))
			}
		})
	}
}

func TestGetCouponClaimedByFormat(t *testing.T) {
	waitForServer(t)

	// Create a coupon
	couponName := fmt.Sprintf("FORMAT_%d", time.Now().UnixNano())
	createPayload := map[string]interface{}{
		"name":   couponName,
		"amount": 5,
	}
	body, _ := json.Marshal(createPayload)
	resp, err := http.Post(apiBaseURL+"/coupons", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Claim by 3 different users
	users := []string{"user1", "user2", "user3"}
	for _, userID := range users {
		claimPayload := map[string]interface{}{
			"user_id":     userID,
			"coupon_name": couponName,
		}
		body, _ := json.Marshal(claimPayload)
		resp, err := http.Post(apiBaseURL+"/coupons/claim", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Get the coupon and verify claimed_by format
	resp, err = http.Get(apiBaseURL + "/coupons/" + couponName)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var coupon Coupon
	err = json.NewDecoder(resp.Body).Decode(&coupon)
	require.NoError(t, err)

	// Verify claimed_by is an array with correct users
	assert.NotNil(t, coupon.ClaimedBy)
	assert.Equal(t, 3, len(coupon.ClaimedBy))
	assert.Equal(t, 2, coupon.RemainingAmount)

	// Verify all users are in the claimed_by list
	for _, userID := range users {
		assert.Contains(t, coupon.ClaimedBy, userID)
	}
}
