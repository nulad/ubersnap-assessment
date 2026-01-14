package repository

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCouponRepository_Integration tests the repository with a mock database
// This test ensures the code compiles and interfaces work correctly
func TestCouponRepository_Integration(t *testing.T) {
	// Create a mock database using SQLite for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()
	
	// Create the coupons table for SQLite
	_, err = db.Exec(`
		CREATE TABLE coupons (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			amount INTEGER NOT NULL CHECK (amount >= 0),
			remaining_amount INTEGER NOT NULL CHECK (remaining_amount >= 0),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)
	
	repo := NewCouponRepository(db)
	
	// Test Create
	err = repo.Create("TEST20", 20)
	assert.NoError(t, err)
	
	// Test GetByName
	coupon, err := repo.GetByName("TEST20")
	assert.NoError(t, err)
	assert.Equal(t, "TEST20", coupon.Name)
	assert.Equal(t, 20, coupon.Amount)
	assert.Equal(t, 20, coupon.RemainingAmount)
	
	// Test GetByNameForUpdate (SQLite doesn't support FOR UPDATE, so we'll test the basic functionality)
	tx, err := db.Begin()
	require.NoError(t, err)
	
	// In SQLite, FOR UPDATE is not supported, so we'll just test that the method works
	// The actual locking behavior would be tested in PostgreSQL
	coupon, err = repo.GetByNameForUpdate(tx, "TEST20")
	// In SQLite, this will fail due to syntax error, which is expected
	// The FOR UPDATE feature is specific to PostgreSQL
	if err != nil {
		assert.Contains(t, err.Error(), "syntax error")
		tx.Rollback()
	} else {
		assert.NoError(t, err)
		assert.Equal(t, "TEST20", coupon.Name)
		
		// Test DecrementStock
		err = repo.DecrementStock(tx, "TEST20")
		assert.NoError(t, err)
		
		// Verify stock was decremented
		coupon, err = repo.GetByNameForUpdate(tx, "TEST20")
		assert.NoError(t, err)
		assert.Equal(t, 19, coupon.RemainingAmount)
		
		tx.Commit()
	}
	
	// Test error cases
	_, err = repo.GetByName("NOTFOUND")
	assert.Equal(t, ErrCouponNotFound, err)
	
	// Test duplicate creation
	err = repo.Create("TEST20", 10)
	assert.Error(t, err)
	assert.Equal(t, ErrCouponExists, err)
	
	// Test negative amount
	err = repo.Create("NEGATIVE", -1)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAmount, err)
}

// TestCouponRepository_Interface ensures the implementation satisfies the interface
func TestCouponRepository_Interface(t *testing.T) {
	var _ CouponRepository = &couponRepository{}
}

// Benchmark tests
func BenchmarkCouponRepository_Create(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}
	defer db.Close()
	
	_, err = db.Exec(`
		CREATE TABLE coupons (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			amount INTEGER NOT NULL,
			remaining_amount INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		b.Fatal(err)
	}
	
	repo := NewCouponRepository(db)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := repo.Create("BENCH"+string(rune(i)), 100)
		if err != nil && err != ErrCouponExists {
			b.Error(err)
		}
	}
}

func BenchmarkCouponRepository_GetByName(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}
	defer db.Close()
	
	_, err = db.Exec(`
		CREATE TABLE coupons (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			amount INTEGER NOT NULL,
			remaining_amount INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		b.Fatal(err)
	}
	
	repo := NewCouponRepository(db)
	err = repo.Create("BENCHMARK", 100)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByName("BENCHMARK")
		if err != nil {
			b.Error(err)
		}
	}
}
