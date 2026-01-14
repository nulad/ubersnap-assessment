package repository

import (
	"errors"
	"strings"
)

// Custom error types for coupon repository
var (
	// ErrCouponNotFound is returned when a coupon is not found
	ErrCouponNotFound = errors.New("coupon not found")
	
	// ErrCouponExists is returned when trying to create a duplicate coupon
	ErrCouponExists = errors.New("coupon already exists")
	
	// ErrInvalidAmount is returned when the amount is invalid
	ErrInvalidAmount = errors.New("invalid coupon amount")
	
	// ErrNoStockAvailable is returned when trying to decrement stock of an empty coupon
	ErrNoStockAvailable = errors.New("no stock available")
)

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	// Check for PostgreSQL unique violation error codes
	return strings.Contains(errStr, "UNIQUE constraint violated") ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		// SQLite error message
		strings.Contains(errStr, "UNIQUE constraint failed")
}
