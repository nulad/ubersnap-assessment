package database

import (
"database/sql"
"errors"

"github.com/lib/pq"
)

// Claim represents a claim record in the database
type Claim struct {
ID          int    `json:"id"`
UserID      string `json:"user_id"`
CouponName  string `json:"coupon_name"`
ClaimedAt   string `json:"claimed_at"`
}

// ErrAlreadyClaimed is returned when a user tries to claim a coupon they've already claimed
var ErrAlreadyClaimed = errors.New("user has already claimed this coupon")

// ClaimRepository provides database operations for claims
type ClaimRepository interface {
Insert(tx *sql.Tx, userID, couponName string) error
GetByCouponName(couponName string) ([]Claim, error)
}

// claimRepository implements ClaimRepository
type claimRepository struct {
db *sql.DB
}

// NewClaimRepository creates a new ClaimRepository instance
func NewClaimRepository(db *sql.DB) ClaimRepository {
return &claimRepository{db: db}
}

// Insert inserts a new claim record
// Returns ErrAlreadyClaimed if the user has already claimed this coupon
func (r *claimRepository) Insert(tx *sql.Tx, userID, couponName string) error {
query := `INSERT INTO claims (user_id, coupon_name) VALUES ($1, $2)`

var err error
if tx != nil {
_, err = tx.Exec(query, userID, couponName)
} else {
_, err = r.db.Exec(query, userID, couponName)
}

if err != nil {
// Check for UNIQUE constraint violation
if pqErr, ok := err.(*pq.Error); ok {
if pqErr.Code == "23505" { // unique_violation
return ErrAlreadyClaimed
}
}
return err
}

return nil
}

// GetByCouponName retrieves all claims for a specific coupon
func (r *claimRepository) GetByCouponName(couponName string) ([]Claim, error) {
query := `SELECT id, user_id, coupon_name, claimed_at FROM claims WHERE coupon_name = $1 ORDER BY claimed_at DESC`

rows, err := r.db.Query(query, couponName)
if err != nil {
return nil, err
}
defer rows.Close()

var claims []Claim
for rows.Next() {
var claim Claim
err := rows.Scan(&claim.ID, &claim.UserID, &claim.CouponName, &claim.ClaimedAt)
if err != nil {
return nil, err
}
claims = append(claims, claim)
}

if err = rows.Err(); err != nil {
return nil, err
}

return claims, nil
}
