package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nulad/ubersnap-assessment/internal/database"
	"github.com/nulad/ubersnap-assessment/internal/repository"
)

// MockCouponRepository is a mock implementation of CouponRepository
type MockCouponRepository struct {
	mock.Mock
}

func (m *MockCouponRepository) Create(name string, amount int) error {
	args := m.Called(name, amount)
	return args.Error(0)
}

func (m *MockCouponRepository) GetByName(name string) (*repository.Coupon, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByNameForUpdate(tx *sql.Tx, name string) (*repository.Coupon, error) {
	args := m.Called(tx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Coupon), args.Error(1)
}

func (m *MockCouponRepository) DecrementStock(tx *sql.Tx, name string) error {
	args := m.Called(tx, name)
	return args.Error(0)
}

// MockClaimRepository is a mock implementation of ClaimRepository
type MockClaimRepository struct {
	mock.Mock
}

func (m *MockClaimRepository) Insert(tx *sql.Tx, userID, couponName string) error {
	args := m.Called(tx, userID, couponName)
	return args.Error(0)
}

func (m *MockClaimRepository) GetByCouponName(couponName string) ([]database.Claim, error) {
	args := m.Called(couponName)
	return args.Get(0).([]database.Claim), args.Error(1)
}

func TestCouponService_CreateCoupon(t *testing.T) {
	mockCouponRepo := new(MockCouponRepository)
	mockClaimRepo := new(MockClaimRepository)
	
	// Create a mock database connection
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	service := NewCouponService(mockCouponRepo, mockClaimRepo, db)

	t.Run("successful coupon creation", func(t *testing.T) {
		mockCouponRepo.On("Create", "TEST10", 10).Return(nil)

		err := service.CreateCoupon("TEST10", 10)

		assert.NoError(t, err)
		mockCouponRepo.AssertExpectations(t)
	})

	t.Run("invalid amount", func(t *testing.T) {
		err := service.CreateCoupon("INVALID", 0)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, repository.ErrInvalidAmount))
	})

	t.Run("negative amount", func(t *testing.T) {
		err := service.CreateCoupon("NEGATIVE", -5)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, repository.ErrInvalidAmount))
	})

	t.Run("coupon already exists", func(t *testing.T) {
		mockCouponRepo.On("Create", "DUPLICATE", 10).Return(repository.ErrCouponExists)

		err := service.CreateCoupon("DUPLICATE", 10)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, repository.ErrCouponExists))
		mockCouponRepo.AssertExpectations(t)
	})
}

func TestCouponService_ClaimCoupon(t *testing.T) {
	mockCouponRepo := new(MockCouponRepository)
	mockClaimRepo := new(MockClaimRepository)
	
	// Create a mock database connection
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	service := NewCouponService(mockCouponRepo, mockClaimRepo, db)

	t.Run("successful claim", func(t *testing.T) {
		coupon := &repository.Coupon{
			ID:             1,
			Name:           "TEST10",
			Amount:         10,
			RemainingAmount: 10,
		}

		// Mock transaction begin
		sqlMock.ExpectBegin()
		
		mockCouponRepo.On("GetByNameForUpdate", mock.AnythingOfType("*sql.Tx"), "TEST10").Return(coupon, nil)
		mockClaimRepo.On("Insert", mock.AnythingOfType("*sql.Tx"), "user123", "TEST10").Return(nil)
		mockCouponRepo.On("DecrementStock", mock.AnythingOfType("*sql.Tx"), "TEST10").Return(nil)
		
		// Mock transaction commit
		sqlMock.ExpectCommit()

		err := service.ClaimCoupon("user123", "TEST10")

		assert.NoError(t, err)
		mockCouponRepo.AssertExpectations(t)
		mockClaimRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("coupon not found", func(t *testing.T) {
		// Mock transaction begin and rollback
		sqlMock.ExpectBegin()
		sqlMock.ExpectRollback()
		
		mockCouponRepo.On("GetByNameForUpdate", mock.AnythingOfType("*sql.Tx"), "NOTFOUND").Return(nil, repository.ErrCouponNotFound)

		err := service.ClaimCoupon("user123", "NOTFOUND")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrCouponNotFound))
		mockCouponRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("no stock available", func(t *testing.T) {
		coupon := &repository.Coupon{
			ID:             1,
			Name:           "EMPTY",
			Amount:         10,
			RemainingAmount: 0,
		}

		// Mock transaction begin and rollback
		sqlMock.ExpectBegin()
		sqlMock.ExpectRollback()
		
		mockCouponRepo.On("GetByNameForUpdate", mock.AnythingOfType("*sql.Tx"), "EMPTY").Return(coupon, nil)

		err := service.ClaimCoupon("user123", "EMPTY")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNoStock))
		mockCouponRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("user already claimed", func(t *testing.T) {
		coupon := &repository.Coupon{
			ID:             1,
			Name:           "CLAIMED",
			Amount:         10,
			RemainingAmount: 10,
		}

		// Mock transaction begin and rollback
		sqlMock.ExpectBegin()
		sqlMock.ExpectRollback()
		
		mockCouponRepo.On("GetByNameForUpdate", mock.AnythingOfType("*sql.Tx"), "CLAIMED").Return(coupon, nil)
		mockClaimRepo.On("Insert", mock.AnythingOfType("*sql.Tx"), "user123", "CLAIMED").Return(database.ErrAlreadyClaimed)

		err := service.ClaimCoupon("user123", "CLAIMED")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrAlreadyClaimed))
		mockCouponRepo.AssertExpectations(t)
		mockClaimRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("database error during stock decrement", func(t *testing.T) {
		coupon := &repository.Coupon{
			ID:             1,
			Name:           "ERROR",
			Amount:         10,
			RemainingAmount: 10,
		}

		// Mock transaction begin and rollback
		sqlMock.ExpectBegin()
		sqlMock.ExpectRollback()
		
		mockCouponRepo.On("GetByNameForUpdate", mock.AnythingOfType("*sql.Tx"), "ERROR").Return(coupon, nil)
		mockClaimRepo.On("Insert", mock.AnythingOfType("*sql.Tx"), "user123", "ERROR").Return(nil)
		mockCouponRepo.On("DecrementStock", mock.AnythingOfType("*sql.Tx"), "ERROR").Return(repository.ErrNoStockAvailable)

		err := service.ClaimCoupon("user123", "ERROR")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNoStock))
		mockCouponRepo.AssertExpectations(t)
		mockClaimRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	})
}

func TestCouponService_GetCouponDetails(t *testing.T) {
	mockCouponRepo := new(MockCouponRepository)
	mockClaimRepo := new(MockClaimRepository)
	
	// Create a mock database connection
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	service := NewCouponService(mockCouponRepo, mockClaimRepo, db)

	t.Run("successful retrieval", func(t *testing.T) {
		coupon := &repository.Coupon{
			ID:             1,
			Name:           "TEST10",
			Amount:         10,
			RemainingAmount: 7,
		}

		mockCouponRepo.On("GetByName", "TEST10").Return(coupon, nil)

		details, err := service.GetCouponDetails("TEST10")

		assert.NoError(t, err)
		assert.Equal(t, &CouponDetails{
			ID:             1,
			Name:           "TEST10",
			Amount:         10,
			RemainingAmount: 7,
		}, details)
		mockCouponRepo.AssertExpectations(t)
	})

	t.Run("coupon not found", func(t *testing.T) {
		mockCouponRepo.On("GetByName", "NOTFOUND").Return(nil, repository.ErrCouponNotFound)

		details, err := service.GetCouponDetails("NOTFOUND")

		assert.Error(t, err)
		assert.Nil(t, details)
		assert.True(t, errors.Is(err, ErrCouponNotFound))
		mockCouponRepo.AssertExpectations(t)
	})
}
