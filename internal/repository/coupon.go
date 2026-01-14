package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// Coupon represents a coupon in the database
type Coupon struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Amount         int       `json:"amount"`
	RemainingAmount int      `json:"remaining_amount"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CouponRepository defines the interface for coupon CRUD operations
type CouponRepository interface {
	Create(name string, amount int) error
	GetByName(name string) (*Coupon, error)
	GetByNameForUpdate(tx *sql.Tx, name string) (*Coupon, error)
	DecrementStock(tx *sql.Tx, name string) error
}

// couponRepository implements CouponRepository
type couponRepository struct {
	db *sql.DB
}

// NewCouponRepository creates a new instance of CouponRepository
func NewCouponRepository(db *sql.DB) CouponRepository {
	return &couponRepository{db: db}
}

// Create inserts a new coupon into the database
func (r *couponRepository) Create(name string, amount int) error {
	if amount < 0 {
		return ErrInvalidAmount
	}
	
	query := `INSERT INTO coupons (name, amount, remaining_amount) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, name, amount, amount)
	if err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			return ErrCouponExists
		}
		return fmt.Errorf("failed to create coupon: %w", err)
	}
	return nil
}

// GetByName retrieves a coupon by its name
func (r *couponRepository) GetByName(name string) (*Coupon, error) {
	query := `SELECT id, name, amount, remaining_amount, created_at, updated_at FROM coupons WHERE name = $1`
	
	var coupon Coupon
	err := r.db.QueryRow(query, name).Scan(
		&coupon.ID,
		&coupon.Name,
		&coupon.Amount,
		&coupon.RemainingAmount,
		&coupon.CreatedAt,
		&coupon.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCouponNotFound
		}
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	
	return &coupon, nil
}

// GetByNameForUpdate retrieves a coupon by its name with a row lock
func (r *couponRepository) GetByNameForUpdate(tx *sql.Tx, name string) (*Coupon, error) {
	query := `SELECT id, name, amount, remaining_amount, created_at, updated_at FROM coupons WHERE name = $1 FOR UPDATE`
	
	var coupon Coupon
	err := tx.QueryRow(query, name).Scan(
		&coupon.ID,
		&coupon.Name,
		&coupon.Amount,
		&coupon.RemainingAmount,
		&coupon.CreatedAt,
		&coupon.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCouponNotFound
		}
		return nil, fmt.Errorf("failed to get coupon for update: %w", err)
	}
	
	return &coupon, nil
}

// DecrementStock decrements the remaining amount of a coupon by 1
func (r *couponRepository) DecrementStock(tx *sql.Tx, name string) error {
	query := `UPDATE coupons SET remaining_amount = remaining_amount - 1 WHERE name = $1 AND remaining_amount > 0`
	
	result, err := tx.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to decrement stock: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return ErrNoStockAvailable
	}
	
	return nil
}
