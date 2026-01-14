package dto

// CreateCouponRequest represents the request body for creating a new coupon
type CreateCouponRequest struct {
	Name   string `json:"name" binding:"required"`
	Amount int    `json:"amount" binding:"required,gt=0"`
}

// ClaimCouponRequest represents the request body for claiming a coupon
type ClaimCouponRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	CouponName string `json:"coupon_name" binding:"required"`
}

// CouponDetailsResponse represents the response for coupon details
type CouponDetailsResponse struct {
	Name            string   `json:"name"`
	Amount          int      `json:"amount"`
	RemainingAmount int      `json:"remaining_amount"`
	ClaimedBy       []string `json:"claimed_by"`
}
