package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nulad/ubersnap-assessment/internal/repository"
	"github.com/nulad/ubersnap-assessment/internal/service"
	"github.com/stretchr/testify/assert"
)

// mockCouponService is a mock implementation of CouponService for testing
type mockCouponService struct {
	createCouponFunc     func(name string, amount int) error
	claimCouponFunc      func(userID, couponName string) error
	getCouponDetailsFunc func(name string) (*service.CouponDetails, error)
}

func (m *mockCouponService) CreateCoupon(name string, amount int) error {
	if m.createCouponFunc != nil {
		return m.createCouponFunc(name, amount)
	}
	return nil
}

func (m *mockCouponService) ClaimCoupon(userID, couponName string) error {
	if m.claimCouponFunc != nil {
		return m.claimCouponFunc(userID, couponName)
	}
	return nil
}

func (m *mockCouponService) GetCouponDetails(name string) (*service.CouponDetails, error) {
	if m.getCouponDetailsFunc != nil {
		return m.getCouponDetailsFunc(name)
	}
	return nil, nil
}

func TestCreateCoupon_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{
		createCouponFunc: func(name string, amount int) error {
			return nil
		},
	}
	h := NewCouponHandler(mockService)

	body, _ := json.Marshal(map[string]any{"name": "PROMO_SUPER", "amount": 100})
	req := httptest.NewRequest(http.MethodPost, "/api/coupons", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.Default()
	router.POST("/api/coupons", h.CreateCoupon)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateCoupon_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{}
	h := NewCouponHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/coupons", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.Default()
	router.POST("/api/coupons", h.CreateCoupon)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateCoupon_DuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{
		createCouponFunc: func(name string, amount int) error {
			return repository.ErrCouponExists
		},
	}
	h := NewCouponHandler(mockService)

	body, _ := json.Marshal(map[string]any{"name": "PROMO_SUPER", "amount": 100})
	req := httptest.NewRequest(http.MethodPost, "/api/coupons", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.Default()
	router.POST("/api/coupons", h.CreateCoupon)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClaimCoupon_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{
		claimCouponFunc: func(userID, couponName string) error {
			return nil // Success
		},
	}
	handler := NewCouponHandler(mockService)

	// Create request
	reqBody := map[string]string{
		"user_id":     "user_123",
		"coupon_name": "PROMO_SUPER",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.POST("/api/coupons/claim", handler.ClaimCoupon)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Coupon claimed successfully", response["message"])
}

func TestClaimCoupon_AlreadyClaimed(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{
		claimCouponFunc: func(userID, couponName string) error {
			return service.ErrAlreadyClaimed // User already claimed
		},
	}
	handler := NewCouponHandler(mockService)

	// Create request
	reqBody := map[string]string{
		"user_id":     "user_123",
		"coupon_name": "PROMO_SUPER",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.POST("/api/coupons/claim", handler.ClaimCoupon)
	router.ServeHTTP(w, req)

	// Assert - spec requirement: 409 for already claimed
	assert.Equal(t, http.StatusConflict, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "already claimed")
}

func TestClaimCoupon_NoStock(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{
		claimCouponFunc: func(userID, couponName string) error {
			return service.ErrNoStock // No stock available
		},
	}
	handler := NewCouponHandler(mockService)

	// Create request
	reqBody := map[string]string{
		"user_id":     "user_123",
		"coupon_name": "PROMO_SUPER",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.POST("/api/coupons/claim", handler.ClaimCoupon)
	router.ServeHTTP(w, req)

	// Assert - 400 for no stock
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "no stock")
}

func TestClaimCoupon_CouponNotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{
		claimCouponFunc: func(userID, couponName string) error {
			return service.ErrCouponNotFound // Coupon doesn't exist
		},
	}
	handler := NewCouponHandler(mockService)

	// Create request
	reqBody := map[string]string{
		"user_id":     "user_123",
		"coupon_name": "NONEXISTENT",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.POST("/api/coupons/claim", handler.ClaimCoupon)
	router.ServeHTTP(w, req)

	// Assert - 404 for not found
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "not found")
}

func TestClaimCoupon_EmptyUserID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{}
	handler := NewCouponHandler(mockService)

	// Create request with empty user_id
	reqBody := map[string]string{
		"user_id":     "",
		"coupon_name": "PROMO_SUPER",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.POST("/api/coupons/claim", handler.ClaimCoupon)
	router.ServeHTTP(w, req)

	// Assert - 400 for empty user_id
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["error"])
}

func TestClaimCoupon_EmptyCouponName(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{}
	handler := NewCouponHandler(mockService)

	// Create request with empty coupon_name
	reqBody := map[string]string{
		"user_id":     "user_123",
		"coupon_name": "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.POST("/api/coupons/claim", handler.ClaimCoupon)
	router.ServeHTTP(w, req)

	// Assert - 400 for empty coupon_name
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["error"])
}

func TestClaimCoupon_InvalidJSON(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{}
	handler := NewCouponHandler(mockService)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.POST("/api/coupons/claim", handler.ClaimCoupon)
	router.ServeHTTP(w, req)

	// Assert - 400 for invalid JSON
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Invalid request")
}

func TestClaimCoupon_InternalError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{
		claimCouponFunc: func(userID, couponName string) error {
			return errors.New("database error") // Some other error
		},
	}
	handler := NewCouponHandler(mockService)

	// Create request
	reqBody := map[string]string{
		"user_id":     "user_123",
		"coupon_name": "PROMO_SUPER",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/coupons/claim", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.POST("/api/coupons/claim", handler.ClaimCoupon)
	router.ServeHTTP(w, req)

	// Assert - 500 for internal errors
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Failed to claim coupon", response["error"])
}

func TestGetCoupon_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	expectedDetails := &service.CouponDetails{
		ID:              1,
		Name:            "PROMO_SUPER",
		Amount:          100,
		RemainingAmount: 50,
		ClaimedBy:       []string{"user_123", "user_456", "user_789"},
	}
	mockService := &mockCouponService{
		getCouponDetailsFunc: func(name string) (*service.CouponDetails, error) {
			assert.Equal(t, "PROMO_SUPER", name)
			return expectedDetails, nil
		},
	}
	handler := NewCouponHandler(mockService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/coupons/PROMO_SUPER", nil)
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.GET("/api/coupons/:name", handler.GetCoupon)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response service.CouponDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedDetails.ID, response.ID)
	assert.Equal(t, expectedDetails.Name, response.Name)
	assert.Equal(t, expectedDetails.Amount, response.Amount)
	assert.Equal(t, expectedDetails.RemainingAmount, response.RemainingAmount)
	assert.Equal(t, expectedDetails.ClaimedBy, response.ClaimedBy)
}

func TestGetCoupon_NotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{
		getCouponDetailsFunc: func(name string) (*service.CouponDetails, error) {
			return nil, service.ErrCouponNotFound
		},
	}
	handler := NewCouponHandler(mockService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/coupons/NONEXISTENT", nil)
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.GET("/api/coupons/:name", handler.GetCoupon)
	router.ServeHTTP(w, req)

	// Assert - 404 for not found
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Coupon not found")
}

func TestGetCoupon_EmptyName(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{}
	handler := NewCouponHandler(mockService)

	// Create request with empty name (this would need a custom route or test differently)
	// Since Gin doesn't match empty params, we'll test the handler directly
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/coupons/", nil)
	c.Params = gin.Params{gin.Param{Key: "name", Value: ""}}

	// Call handler directly
	handler.GetCoupon(c)

	// Assert - 400 for empty name
	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestGetCoupon_EmptyClaims(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	expectedDetails := &service.CouponDetails{
		ID:              2,
		Name:            "NEW_PROMO",
		Amount:          200,
		RemainingAmount: 200,
		ClaimedBy:       []string{}, // Empty claims array
	}
	mockService := &mockCouponService{
		getCouponDetailsFunc: func(name string) (*service.CouponDetails, error) {
			return expectedDetails, nil
		},
	}
	handler := NewCouponHandler(mockService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/coupons/NEW_PROMO", nil)
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.GET("/api/coupons/:name", handler.GetCoupon)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response service.CouponDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedDetails.ClaimedBy, response.ClaimedBy)
	assert.Empty(t, response.ClaimedBy) // Verify it's empty
}

func TestGetCoupon_InternalError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &mockCouponService{
		getCouponDetailsFunc: func(name string) (*service.CouponDetails, error) {
			return nil, errors.New("database error") // Some internal error
		},
	}
	handler := NewCouponHandler(mockService)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/coupons/PROMO_SUPER", nil)
	w := httptest.NewRecorder()

	// Setup router
	router := gin.Default()
	router.GET("/api/coupons/:name", handler.GetCoupon)
	router.ServeHTTP(w, req)

	// Assert - 500 for internal errors
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Failed to retrieve coupon", response["error"])
}
