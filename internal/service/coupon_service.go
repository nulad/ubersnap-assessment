package service

import (
	"database/sql"
	"errors"

	"github.com/nulad/ubersnap-assessment/internal/database"
	"github.com/nulad/ubersnap-assessment/internal/repository"
)

// Custom error types for service layer
var (
	// ErrAlreadyClaimed is returned when a user tries to claim a coupon they've already claimed
	ErrAlreadyClaimed = errors.New("user has already claimed this coupon")
	
	// ErrNoStock is returned when a coupon has no remaining stock
	ErrNoStock = errors.New("coupon has no stock available")
	
	// ErrCouponNotFound is returned when a coupon is not found
	ErrCouponNotFound = errors.New("coupon not found")
)

// CouponDetails represents the detailed information about a coupon
type CouponDetails struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Amount         int    `json:"amount"`
	RemainingAmount int    `json:"remaining_amount"`
}

// CouponService defines the interface for coupon business logic
type CouponService interface {
	CreateCoupon(name string, amount int) error
	ClaimCoupon(userID, couponName string) error
	GetCouponDetails(name string) (*CouponDetails, error)
}

// couponService implements CouponService
type couponService struct {
	couponRepo repository.CouponRepository
	claimRepo  database.ClaimRepository
	db         *sql.DB
}

// NewCouponService creates a new instance of CouponService
func NewCouponService(couponRepo repository.CouponRepository, claimRepo database.ClaimRepository, db *sql.DB) CouponService {
	return &couponService{
		couponRepo: couponRepo,
		claimRepo:  claimRepo,
		db:         db,
	}
}

// CreateCoupon creates a new coupon with the given name and amount
func (s *couponService) CreateCoupon(name string, amount int) error {
	if amount <= 0 {
		return repository.ErrInvalidAmount
	}
	
	return s.couponRepo.Create(name, amount)
}

// ClaimCoupon processes a coupon claim with atomic transaction
// CRITICAL: This method implements the atomic transaction flow to prevent race conditions
func (s *couponService) ClaimCoupon(userID, couponName string) error {
	// Begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	
	// Track whether we need to rollback
	shouldRollback := true
	defer func() {
		if shouldRollback {
			tx.Rollback()
		}
	}()
	
	// 1. Lock coupon with SELECT FOR UPDATE
	coupon, err := s.couponRepo.GetByNameForUpdate(tx, couponName)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	
	// 2. Check stock availability
	if coupon.RemainingAmount <= 0 {
		return ErrNoStock
	}
	
	// 3. Insert claim record
	err = s.claimRepo.Insert(tx, userID, couponName)
	if err != nil {
		if errors.Is(err, database.ErrAlreadyClaimed) {
			return ErrAlreadyClaimed
		}
		return err
	}
	
	// 4. Decrement stock
	err = s.couponRepo.DecrementStock(tx, couponName)
	if err != nil {
		if errors.Is(err, repository.ErrNoStockAvailable) {
			return ErrNoStock
		}
		return err
	}
	
	// 5. Commit transaction if all steps succeeded
	// Mark as successful before commit to avoid rollback on commit failure
	shouldRollback = false
	err = tx.Commit()
	if err != nil {
		return err
	}
	
	return nil
}

// GetCouponDetails retrieves coupon details by name
func (s *couponService) GetCouponDetails(name string) (*CouponDetails, error) {
	coupon, err := s.couponRepo.GetByName(name)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}
	
	return &CouponDetails{
		ID:             coupon.ID,
		Name:           coupon.Name,
		Amount:         coupon.Amount,
		RemainingAmount: coupon.RemainingAmount,
	}, nil
}
