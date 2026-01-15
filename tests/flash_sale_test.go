package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080/api"

type Coupon struct {
	Name            string   `json:"name"`
	Amount          int      `json:"amount"`
	RemainingAmount int      `json:"remaining_amount"`
	ClaimedBy       []string `json:"claimed_by"`
}

func createCoupon(t *testing.T, name string, amount int) {
	// Add retry logic for server startup
	var resp *http.Response
	var err error
	
	for i := 0; i < 10; i++ {
		payload := map[string]interface{}{
			"name":   name,
			"amount": amount,
		}
		body, _ := json.Marshal(payload)
		resp, err = http.Post(baseURL+"/coupons", "application/json", bytes.NewBuffer(body))
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("Failed to connect to API server. Is it running? Error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// If 409, maybe it already exists from a previous run? 
		// But we should probably fail if we expect a clean state.
		// Detailed error reading
		t.Fatalf("Failed to create coupon. Expected 201 Created, got %d", resp.StatusCode)
	}
}

func claimCoupon(userId, couponName string) *http.Response {
	payload := map[string]string{
		"user_id":     userId,
		"coupon_name": couponName,
	}
	body, _ := json.Marshal(payload)
	
	// Use a client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Post(baseURL+"/coupons/claim", "application/json", bytes.NewBuffer(body))
	if err != nil {
		// Return 500 response on connection error to satisfy test signature
		return &http.Response{
			StatusCode: 500,
			Status:     "500 Internal Server Error (Client Connection Failed)",
			Body:       http.NoBody,
		}
	}
	// We don't verify status code here, just return response
	resp.Body.Close()
	return resp
}

func getCoupon(t *testing.T, name string) Coupon {
	resp, err := http.Get(baseURL + "/coupons/" + name)
	if err != nil {
		t.Fatalf("Failed to get coupon: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	
	var coupon Coupon
	if err := json.NewDecoder(resp.Body).Decode(&coupon); err != nil {
		t.Fatalf("Failed to decode coupon: %v", err)
	}
	return coupon
}

func TestFlashSale(t *testing.T) {
	// This is an integration test that requires the API server to be running.
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8080", 500*time.Millisecond)
	if err != nil {
		t.Skipf("Skipping Flash Sale integration test: API server not reachable on localhost:8080: %v", err)
	}
	_ = conn.Close()

	// Setup: Create coupon with 5 stock with unique name to avoid conflicts
	couponName := fmt.Sprintf("FLASH_%d", time.Now().UnixNano())
	createCoupon(t, couponName, 5)

	// Execute: 50 concurrent goroutines
	var wg sync.WaitGroup
	results := make(chan int, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(userNum int) {
			defer wg.Done()

			// Each goroutine makes HTTP POST request
			resp := claimCoupon(
				fmt.Sprintf("user_%d", userNum),
				couponName,
			)
			results <- resp.StatusCode
		}(i)
	}

	// Wait for all to complete
	wg.Wait()
	close(results)

	// Count results
	successCount := 0
	failCount := 0
	for code := range results {
		if code == 200 || code == 201 {
			successCount++
		} else {
			failCount++
		}
	}

	// Assert: Exactly 5 success, 45 failures
	assert.Equal(t, 5, successCount, "Expected exactly 5 successful claims")
	assert.Equal(t, 45, failCount, "Expected exactly 45 failed claims")

	// Verify database state
	coupon := getCoupon(t, couponName)
	assert.Equal(t, 0, coupon.RemainingAmount, "Expected remaining amount to be 0")
	assert.Equal(t, 5, len(coupon.ClaimedBy), "Expected exactly 5 users in claimed_by list")
}

func TestDoubleDipConcurrency(t *testing.T) {
	// This test verifies exactly 1 claim succeeds when same user makes 10 concurrent requests
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8080", 500*time.Millisecond)
	if err != nil {
		t.Skipf("Skipping Double Dip concurrency test: API server not reachable on localhost:8080: %v", err)
	}
	_ = conn.Close()

	// Setup: Create coupon with amount=100 with unique name to avoid conflicts
	couponName := fmt.Sprintf("DOUBLE_DIP_%d", time.Now().UnixNano())
	createCoupon(t, couponName, 100)

	// Execute: Spawn 10 goroutines, ALL using SAME user_id
	var wg sync.WaitGroup
	results := make(chan int, 10)
	const userId = "double_dip_user"

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// All goroutines use the SAME user_id
			resp := claimCoupon(userId, couponName)
			results <- resp.StatusCode
		}()
	}

	// Wait for all to complete
	wg.Wait()
	close(results)

	// Collect and count results
	successCount := 0
	conflictCount := 0
	otherErrors := 0
	
	for code := range results {
		if code == 200 || code == 201 {
			successCount++
		} else if code == 409 {
			conflictCount++
		} else {
			otherErrors++
		}
	}

	// Assert: Exactly 1 success, 9 conflicts
	assert.Equal(t, 1, successCount, "Expected exactly 1 successful claim")
	assert.Equal(t, 9, conflictCount, "Expected exactly 9 claims to fail with 409 (already claimed)")
	assert.Equal(t, 0, otherErrors, "Expected no other errors")

	// Verify database state
	coupon := getCoupon(t, couponName)
	assert.Equal(t, 99, coupon.RemainingAmount, "Expected remaining amount to be 99 after 1 claim")
	assert.Equal(t, 1, len(coupon.ClaimedBy), "Expected exactly 1 user in claimed_by list")
	assert.Contains(t, coupon.ClaimedBy, userId, "Expected the double dip user to be in claimed_by list")
}
