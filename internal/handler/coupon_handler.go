package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nulad/ubersnap-assessment/internal/dto"
	"github.com/nulad/ubersnap-assessment/internal/repository"
	"github.com/nulad/ubersnap-assessment/internal/service"
)

// CouponHandler handles HTTP requests for coupon operations
type CouponHandler struct {
	couponService service.CouponService
}

// NewCouponHandler creates a new instance of CouponHandler
func NewCouponHandler(couponService service.CouponService) *CouponHandler {
	return &CouponHandler{
		couponService: couponService,
	}
}

// CreateCoupon handles POST /api/coupons
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req dto.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.couponService.CreateCoupon(req.Name, req.Amount)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidAmount) || errors.Is(err, repository.ErrCouponExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon"})
		return
	}

	c.Status(http.StatusCreated)
}

// ClaimCoupon handles POST /api/coupons/claim
// This is the CRITICAL CONCURRENCY ENDPOINT
func (h *CouponHandler) ClaimCoupon(c *gin.Context) {
	// Parse JSON body into ClaimCouponRequest struct
	var req dto.ClaimCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate: user_id and coupon_name not empty
	if req.UserID == "" || req.CouponName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id and coupon_name are required",
		})
		return
	}

	// Call service.ClaimCoupon
	err := h.couponService.ClaimCoupon(req.UserID, req.CouponName)
	if err != nil {
		// Map service errors to HTTP codes
		if errors.Is(err, service.ErrAlreadyClaimed) {
			// 409 Conflict - user already claimed this coupon (spec requirement)
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, service.ErrNoStock) {
			// 400 Bad Request - no stock available
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, service.ErrCouponNotFound) {
			// 404 Not Found - coupon doesn't exist
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Internal server error for any other errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to claim coupon",
		})
		return
	}

	// Success - 200 OK
	c.JSON(http.StatusOK, gin.H{
		"message": "Coupon claimed successfully",
	})
}

// GetCoupon handles GET /api/coupons/:name
func (h *CouponHandler) GetCoupon(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}
