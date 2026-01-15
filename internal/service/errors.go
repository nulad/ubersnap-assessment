package service

import "errors"

// Service layer errors that map to HTTP status codes
var (
	// ErrAlreadyClaimed is returned when a user has already claimed a coupon (HTTP 409)
	ErrAlreadyClaimed = errors.New("user has already claimed this coupon")

	// ErrNoStock is returned when a coupon has no remaining stock (HTTP 400)
	ErrNoStock = errors.New("no stock available for this coupon")

	// ErrCouponNotFound is returned when a coupon doesn't exist (HTTP 404)
	ErrCouponNotFound = errors.New("coupon not found")

	// ErrCouponExists is returned when trying to create a duplicate coupon (HTTP 409)
	ErrCouponExists = errors.New("coupon already exists")

	// ErrInvalidAmount is returned when the coupon amount is invalid (HTTP 400)
	ErrInvalidAmount = errors.New("invalid coupon amount")
)
