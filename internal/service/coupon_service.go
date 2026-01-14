package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/nulad/ubersnap-assessment/internal/database"
	"github.com/nulad/ubersnap-assessment/internal/repository"
)

// CouponDetails represents detailed information about a coupon
type CouponDetails struct {
	Name            string `json:"name"`
	Amount          int    `json:"amount"`
	RemainingAmount int    `json:"remaining_amount"`
	Claims          []ClaimInfo `json:"claims,omitempty"`
}

// ClaimInfo represents information about a claim
type ClaimInfo struct {
	UserID    string `json:"user_id"`
	ClaimedAt string `json:"claimed_at"`
}

// CouponService defines the interface for coupon business logic
type CouponService interface {
	CreateCoupon(name string, amount int) error
	ClaimCoupon(userID, couponName string) error
	GetCouponDetails(name string) (*CouponDetails, error)
}

// couponService implements CouponService
type couponService struct {
	db             *sql.DB
	couponRepo     repository.CouponRepository
	claimRepo      database.ClaimRepository
}

// NewCouponService creates a new instance of CouponService
func NewCouponService(db *sql.DB, couponRepo repository.CouponRepository, claimRepo database.ClaimRepository) CouponService {
	return &couponService{
		db:         db,
		couponRepo: couponRepo,
		claimRepo:  claimRepo,
	}
}

// CreateCoupon creates a new coupon
func (s *couponService) CreateCoupon(name string, amount int) error {
	if amount < 0 {
		return ErrInvalidAmount
	}
	
	err := s.couponRepo.Create(name, amount)
	if err != nil {
		// Map repository errors to service errors
		if errors.Is(err, repository.ErrCouponExists) {
			return ErrCouponExists
		}
		if errors.Is(err, repository.ErrInvalidAmount) {
			return ErrInvalidAmount
		}
		return fmt.Errorf("failed to create coupon: %w", err)
	}
	
	return nil
}

// ClaimCoupon implements the atomic claim transaction logic
// This is THE CRITICAL COMPONENT that prevents race conditions
func (s *couponService) ClaimCoupon(userID, couponName string) error {
	// BEGIN TRANSACTION
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Ensure transaction is rolled back on error
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()
	
	// Step 1: Lock coupon row with SELECT FOR UPDATE
	// This is what prevents race conditions!
	coupon, err := s.couponRepo.GetByNameForUpdate(tx, couponName)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, repository.ErrCouponNotFound) {
			return ErrCouponNotFound
		}
		return fmt.Errorf("failed to get coupon for update: %w", err)
	}
	
	// Step 2: Check stock availability
	if coupon.RemainingAmount <= 0 {
		tx.Rollback()
		return ErrNoStock
	}
	
	// Step 3: Insert claim record
	// This will fail with UNIQUE constraint violation if user already claimed
	err = s.claimRepo.Insert(tx, userID, couponName)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, database.ErrAlreadyClaimed) {
			return ErrAlreadyClaimed
		}
		return fmt.Errorf("failed to insert claim: %w", err)
	}
	
	// Step 4: Decrement stock
	err = s.couponRepo.DecrementStock(tx, couponName)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, repository.ErrNoStockAvailable) {
			return ErrNoStock
		}
		return fmt.Errorf("failed to decrement stock: %w", err)
	}
	
	// COMMIT TRANSACTION
	// Only commits if all steps succeeded
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	log.Printf("Successfully claimed coupon '%s' for user '%s'", couponName, userID)
	return nil
}

// GetCouponDetails retrieves detailed information about a coupon
func (s *couponService) GetCouponDetails(name string) (*CouponDetails, error) {
	// Get coupon information
	coupon, err := s.couponRepo.GetByName(name)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	
	// Get claims for this coupon
	claims, err := s.claimRepo.GetByCouponName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get claims: %w", err)
	}
	
	// Convert to ClaimInfo
	claimInfos := make([]ClaimInfo, len(claims))
	for i, claim := range claims {
		claimInfos[i] = ClaimInfo{
			UserID:    claim.UserID,
			ClaimedAt: claim.ClaimedAt,
		}
	}
	
	return &CouponDetails{
		Name:            coupon.Name,
		Amount:          coupon.Amount,
		RemainingAmount: coupon.RemainingAmount,
		Claims:          claimInfos,
	}, nil
}
