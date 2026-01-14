package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nulad/ubersnap-assessment/internal/service"
)

type CouponHandler struct {
	service service.CouponService
}

func NewCouponHandler(service service.CouponService) *CouponHandler {
	return &CouponHandler{
		service: service,
	}
}

// CreateCoupon handles POST /api/coupons
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// ClaimCoupon handles POST /api/coupons/claim
func (h *CouponHandler) ClaimCoupon(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// GetCoupon handles GET /api/coupons/:name
func (h *CouponHandler) GetCoupon(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}
