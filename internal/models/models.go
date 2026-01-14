package models

import "time"

// Coupon represents a coupon in the system with amount tracking
type Coupon struct {
	ID              int       `db:"id"`
	Name            string    `db:"name"`
	Amount          int       `db:"amount"`
	RemainingAmount int       `db:"remaining_amount"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// Claim represents a user's claim on a coupon
type Claim struct {
	ID         int       `db:"id"`
	UserID     string    `db:"user_id"`
	CouponName string    `db:"coupon_name"`
	ClaimedAt  time.Time `db:"claimed_at"`
}
