package repository

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// For testing, we'll use an in-memory database or test database
	// This is a placeholder - in a real scenario, you'd connect to a test database
	db, err := sql.Open("postgres", "postgres://test:test@localhost/ubersnap_test?sslmode=disable")
	if err != nil {
		t.Skipf("Skipping integration tests: %v", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Skipf("Skipping integration tests: %v", err)
	}
	
	// Clean up before each test
	_, err = db.Exec("DELETE FROM coupons")
	require.NoError(t, err)
	
	return db
}

func TestCouponRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewCouponRepository(db)
	
	t.Run("success", func(t *testing.T) {
		err := repo.Create("TEST10", 10)
		assert.NoError(t, err)
		
		// Verify the coupon was created
		coupon, err := repo.GetByName("TEST10")
		assert.NoError(t, err)
		assert.Equal(t, "TEST10", coupon.Name)
		assert.Equal(t, 10, coupon.Amount)
		assert.Equal(t, 10, coupon.RemainingAmount)
	})
	
	t.Run("duplicate coupon", func(t *testing.T) {
		err := repo.Create("TEST10", 5)
		assert.NoError(t, err)
		
		// Try to create the same coupon again
		err = repo.Create("TEST10", 15)
		assert.Error(t, err)
		assert.Equal(t, ErrCouponExists, err)
	})
	
	t.Run("negative amount", func(t *testing.T) {
		err := repo.Create("NEGATIVE", -5)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAmount, err)
	})
}

func TestCouponRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewCouponRepository(db)
	
	// Setup test data
	err := repo.Create("FINDME", 20)
	require.NoError(t, err)
	
	t.Run("success", func(t *testing.T) {
		coupon, err := repo.GetByName("FINDME")
		assert.NoError(t, err)
		assert.Equal(t, "FINDME", coupon.Name)
		assert.Equal(t, 20, coupon.Amount)
		assert.Equal(t, 20, coupon.RemainingAmount)
		assert.NotZero(t, coupon.CreatedAt)
		assert.NotZero(t, coupon.UpdatedAt)
	})
	
	t.Run("not found", func(t *testing.T) {
		coupon, err := repo.GetByName("NOTFOUND")
		assert.Error(t, err)
		assert.Nil(t, coupon)
		assert.Equal(t, ErrCouponNotFound, err)
	})
}

func TestCouponRepository_GetByNameForUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewCouponRepository(db)
	
	// Setup test data
	err := repo.Create("LOCKME", 15)
	require.NoError(t, err)
	
	t.Run("success", func(t *testing.T) {
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		
		coupon, err := repo.GetByNameForUpdate(tx, "LOCKME")
		assert.NoError(t, err)
		assert.Equal(t, "LOCKME", coupon.Name)
		assert.Equal(t, 15, coupon.Amount)
	})
	
	t.Run("not found", func(t *testing.T) {
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		
		coupon, err := repo.GetByNameForUpdate(tx, "NOTFOUND")
		assert.Error(t, err)
		assert.Nil(t, coupon)
		assert.Equal(t, ErrCouponNotFound, err)
	})
}

func TestCouponRepository_DecrementStock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewCouponRepository(db)
	
	t.Run("success", func(t *testing.T) {
		// Create a coupon with stock
		err := repo.Create("STOCK", 5)
		require.NoError(t, err)
		
		tx, err := db.Begin()
		require.NoError(t, err)
		
		// Decrement stock
		err = repo.DecrementStock(tx, "STOCK")
		assert.NoError(t, err)
		
		// Verify stock was decremented
		coupon, err := repo.GetByNameForUpdate(tx, "STOCK")
		assert.NoError(t, err)
		assert.Equal(t, 4, coupon.RemainingAmount)
		
		tx.Commit()
	})
	
	t.Run("no stock available", func(t *testing.T) {
		// Create a coupon with no stock
		err := repo.Create("EMPTY", 0)
		require.NoError(t, err)
		
		tx, err := db.Begin()
		require.NoError(t, err)
		
		// Try to decrement stock
		err = repo.DecrementStock(tx, "EMPTY")
		assert.Error(t, err)
		assert.Equal(t, ErrNoStockAvailable, err)
		
		tx.Rollback()
	})
	
	t.Run("coupon not found", func(t *testing.T) {
		tx, err := db.Begin()
		require.NoError(t, err)
		
		err = repo.DecrementStock(tx, "NOTFOUND")
		assert.Error(t, err)
		// Note: This might return a different error depending on the database
		// In PostgreSQL, this would affect 0 rows
		
		tx.Rollback()
	})
	
	t.Run("concurrent access", func(t *testing.T) {
		// Create a coupon with 1 stock
		err := repo.Create("SINGLE", 1)
		require.NoError(t, err)
		
		// Start two transactions
		tx1, err := db.Begin()
		require.NoError(t, err)
		
		tx2, err := db.Begin()
		require.NoError(t, err)
		
		// First transaction locks and decrements
		coupon1, err := repo.GetByNameForUpdate(tx1, "SINGLE")
		assert.NoError(t, err)
		assert.Equal(t, 1, coupon1.RemainingAmount)
		
		err = repo.DecrementStock(tx1, "SINGLE")
		assert.NoError(t, err)
		
		// Second transaction should wait for lock (in real scenario)
		// For this test, we'll just show that the stock is now 0
		tx1.Commit()
		
		// Now try to decrement with second transaction
		coupon2, err := repo.GetByNameForUpdate(tx2, "SINGLE")
		assert.NoError(t, err)
		assert.Equal(t, 0, coupon2.RemainingAmount)
		
		err = repo.DecrementStock(tx2, "SINGLE")
		assert.Error(t, err)
		assert.Equal(t, ErrNoStockAvailable, err)
		
		tx2.Rollback()
	})
}

// Mock tests for when database is not available
func TestCouponRepository_MockTests(t *testing.T) {
	// These tests can run without a database
	t.Run("error detection", func(t *testing.T) {
		assert.True(t, isUniqueViolation(sql.ErrNoRows) == false)
		assert.True(t, isUniqueViolation(nil) == false)
	})
	
	t.Run("custom errors", func(t *testing.T) {
		assert.Equal(t, "coupon not found", ErrCouponNotFound.Error())
		assert.Equal(t, "coupon already exists", ErrCouponExists.Error())
		assert.Equal(t, "invalid coupon amount", ErrInvalidAmount.Error())
		assert.Equal(t, "no stock available", ErrNoStockAvailable.Error())
	})
}
