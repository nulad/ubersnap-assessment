package database

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaimRepository_Insert_Success(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewClaimRepository(db)

	// Set up expectations
	mock.ExpectExec(`INSERT INTO claims \(user_id, coupon_name\) VALUES \(\$1, \$2\)`).
		WithArgs("user123", "test-coupon").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute test
	err = repo.Insert(nil, "user123", "test-coupon")

	// Assertions
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimRepository_Insert_AlreadyClaimed(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewClaimRepository(db)

	// Set up expectations for UNIQUE constraint violation
	pqErr := &pq.Error{
		Code: "23505", // unique_violation
	}
	mock.ExpectExec(`INSERT INTO claims \(user_id, coupon_name\) VALUES \(\$1, \$2\)`).
		WithArgs("user123", "test-coupon").
		WillReturnError(pqErr)

	// Execute test
	err = repo.Insert(nil, "user123", "test-coupon")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, ErrAlreadyClaimed, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClaimRepository_GetByCouponName_Success(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewClaimRepository(db)

	// Sample data
	expectedTime := time.Now().Format(time.RFC3339)
	rows := sqlmock.NewRows([]string{"id", "user_id", "coupon_name", "claimed_at"}).
		AddRow(1, "user1", "test-coupon", expectedTime).
		AddRow(2, "user2", "test-coupon", expectedTime)

	// Set up expectations
	mock.ExpectQuery(`SELECT id, user_id, coupon_name, claimed_at FROM claims WHERE coupon_name = \$1 ORDER BY claimed_at DESC`).
		WithArgs("test-coupon").
		WillReturnRows(rows)

	// Execute test
	claims, err := repo.GetByCouponName("test-coupon")

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, claims, 2)
	assert.Equal(t, 1, claims[0].ID)
	assert.Equal(t, "user1", claims[0].UserID)
	assert.Equal(t, "test-coupon", claims[0].CouponName)
	assert.Equal(t, expectedTime, claims[0].ClaimedAt)
	assert.Equal(t, 2, claims[1].ID)
	assert.Equal(t, "user2", claims[1].UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
