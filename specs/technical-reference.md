# Technical Reference Card

Quick reference for critical implementation patterns and decisions.

---

## Database Schema (CRITICAL)

### Coupons Table
```sql
CREATE TABLE coupons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    amount INTEGER NOT NULL CHECK (amount >= 0),
    remaining_amount INTEGER NOT NULL CHECK (remaining_amount >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_coupons_name ON coupons(name);
```

### Claims Table
```sql
CREATE TABLE claims (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    coupon_name VARCHAR(255) NOT NULL,
    claimed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_coupon UNIQUE(user_id, coupon_name),
    CONSTRAINT fk_coupon FOREIGN KEY (coupon_name)
        REFERENCES coupons(name) ON DELETE CASCADE
);

CREATE INDEX idx_claims_user_id ON claims(user_id);
CREATE INDEX idx_claims_coupon_name ON claims(coupon_name);
```

**Why This Matters:**
- `UNIQUE(user_id, coupon_name)` prevents double-dipping at database level
- Separate tables (no embedding) per spec requirement
- Indexes improve query performance under load

---

## Atomic Claim Transaction (CRITICAL)

### The Pattern That Prevents Race Conditions

```go
func (s *CouponService) ClaimCoupon(userID, couponName string) error {
    // START TRANSACTION
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback() // Auto-rollback if not committed

    // STEP 1: Lock the coupon row (blocks other transactions)
    // This is SELECT FOR UPDATE - creates an exclusive lock
    query := `SELECT id, name, amount, remaining_amount
              FROM coupons
              WHERE name = $1
              FOR UPDATE`

    var coupon Coupon
    err = tx.QueryRow(query, couponName).Scan(
        &coupon.ID,
        &coupon.Name,
        &coupon.Amount,
        &coupon.RemainingAmount,
    )
    if err == sql.ErrNoRows {
        return ErrCouponNotFound
    }
    if err != nil {
        return err
    }

    // STEP 2: Check stock
    if coupon.RemainingAmount <= 0 {
        return ErrNoStock // Will trigger rollback
    }

    // STEP 3: Insert claim (UNIQUE constraint enforced here)
    _, err = tx.Exec(
        `INSERT INTO claims (user_id, coupon_name) VALUES ($1, $2)`,
        userID,
        couponName,
    )
    if err != nil {
        // Check for UNIQUE constraint violation
        if pqErr, ok := err.(*pq.Error); ok {
            if pqErr.Code == "23505" { // unique_violation
                return ErrAlreadyClaimed
            }
        }
        return err
    }

    // STEP 4: Decrement stock
    _, err = tx.Exec(
        `UPDATE coupons SET remaining_amount = remaining_amount - 1
         WHERE name = $1`,
        couponName,
    )
    if err != nil {
        return err
    }

    // COMMIT - all or nothing
    return tx.Commit()
}
```

**Why This Works:**
1. `FOR UPDATE` locks the coupon row - other transactions wait
2. All operations in ONE transaction - atomic
3. Any error causes rollback - no partial claims
4. Database enforces UNIQUE constraint - double safety
5. Lock released on COMMIT - next transaction can proceed

---

## HTTP Status Code Mapping

```go
// Service layer errors
var (
    ErrCouponNotFound  = errors.New("coupon not found")
    ErrNoStock         = errors.New("no stock available")
    ErrAlreadyClaimed  = errors.New("user already claimed this coupon")
)

// Handler mapping
switch err {
case nil:
    c.JSON(200, gin.H{"message": "claimed successfully"})
case ErrAlreadyClaimed:
    c.JSON(409, gin.H{"error": "already claimed"}) // ← 409 is preferred per spec
case ErrNoStock:
    c.JSON(400, gin.H{"error": "no stock available"})
case ErrCouponNotFound:
    c.JSON(404, gin.H{"error": "coupon not found"})
default:
    c.JSON(500, gin.H{"error": "internal server error"})
}
```

**Per Spec Requirements:**
- Already claimed: **409 Conflict** (preferred) or 400
- No stock: **400 Bad Request** or 409
- Success: **200 OK** or **201 Created**

---

## PostgreSQL Error Code Detection

```go
import "github.com/lib/pq"

// Check for UNIQUE constraint violation
if pqErr, ok := err.(*pq.Error); ok {
    switch pqErr.Code {
    case "23505": // unique_violation
        return ErrAlreadyClaimed
    case "23503": // foreign_key_violation
        return ErrCouponNotFound
    case "23514": // check_violation
        return errors.New("constraint violation")
    }
}
```

**Common PostgreSQL Error Codes:**
- `23505`: UNIQUE constraint violation
- `23503`: Foreign key violation
- `23514`: CHECK constraint violation

---

## Connection Pool Configuration

```go
import "database/sql"

func InitDB(connStr string) (*sql.DB, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    // Connection pool tuning
    db.SetMaxOpenConns(25)        // Max concurrent connections
    db.SetMaxIdleConns(10)        // Warm connections to keep
    db.SetConnMaxLifetime(5 * time.Minute) // Recycle old connections

    // Retry logic for initial connection
    var pingErr error
    for i := 0; i < 10; i++ {
        pingErr = db.Ping()
        if pingErr == nil {
            break
        }
        time.Sleep(time.Duration(i) * time.Second) // Exponential backoff
    }

    return db, pingErr
}
```

---

## Docker Compose Configuration

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: coupondb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: coupondb
    depends_on:
      postgres:
        condition: service_healthy  # ← Wait for healthcheck

volumes:
  postgres_data:
```

**Key Points:**
- Healthcheck ensures database is ready before API starts
- `depends_on` with `condition: service_healthy` is critical
- Named volume persists data between restarts

---

## Concurrency Test Pattern

### Flash Sale Test (50 concurrent, 5 stock)

```go
func TestFlashSale(t *testing.T) {
    // Setup: Create coupon with 5 stock
    createCoupon(t, "FLASH", 5)

    // Execute: 50 concurrent goroutines
    var wg sync.WaitGroup
    results := make(chan int, 50)

    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func(userNum int) {
            defer wg.Done()

            // Each goroutine makes HTTP POST request
            resp := claimCoupon(
                fmt.Sprintf("user_%d", userNum),
                "FLASH",
            )
            results <- resp.StatusCode
        }(i)
    }

    // Wait for all to complete
    wg.Wait()
    close(results)

    // Count results
    successCount := 0
    failCount := 0
    for code := range results {
        if code == 200 || code == 201 {
            successCount++
        } else {
            failCount++
        }
    }

    // Assert: Exactly 5 success, 45 failures
    assert.Equal(t, 5, successCount)
    assert.Equal(t, 45, failCount)

    // Verify database state
    coupon := getCoupon(t, "FLASH")
    assert.Equal(t, 0, coupon.RemainingAmount)
    assert.Equal(t, 5, len(coupon.ClaimedBy))
}
```

### Double Dip Test (10 concurrent, same user)

```go
func TestDoubleDip(t *testing.T) {
    // Setup: Create coupon with 100 stock
    createCoupon(t, "PROMO", 100)

    // Execute: 10 concurrent goroutines, SAME user
    var wg sync.WaitGroup
    results := make(chan int, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()

            // ALL use same user_id
            resp := claimCoupon("user_123", "PROMO")
            results <- resp.StatusCode
        }()
    }

    wg.Wait()
    close(results)

    // Count results
    successCount := 0
    conflictCount := 0
    for code := range results {
        if code == 200 || code == 201 {
            successCount++
        } else if code == 409 {
            conflictCount++
        }
    }

    // Assert: Exactly 1 success, 9 conflicts (409)
    assert.Equal(t, 1, successCount)
    assert.Equal(t, 9, conflictCount)

    // Verify database state
    coupon := getCoupon(t, "PROMO")
    assert.Equal(t, 99, coupon.RemainingAmount)
    assert.Equal(t, 1, len(coupon.ClaimedBy))
    assert.Contains(t, coupon.ClaimedBy, "user_123")
}
```

---

## API Contract (Exact Per Spec)

### 1. Create Coupon
```
POST /api/coupons
Content-Type: application/json

{
  "name": "PROMO_SUPER",
  "amount": 100
}

Response: 201 Created
```

### 2. Claim Coupon
```
POST /api/coupons/claim
Content-Type: application/json

{
  "user_id": "user_12345",
  "coupon_name": "PROMO_SUPER"
}

Response: 200/201 (success)
          409 (already claimed - preferred)
          400 (no stock)
```

### 3. Get Coupon Details
```
GET /api/coupons/{name}

Response: 200 OK
{
  "name": "PROMO_SUPER",
  "amount": 100,
  "remaining_amount": 0,
  "claimed_by": ["user_12345", "user_67890"]
}
```

---

## Common Pitfalls to Avoid

### ❌ DON'T: Use optimistic locking
```go
// WRONG - Race condition!
stock := GetStock(couponName)
if stock > 0 {
    InsertClaim()
    DecrementStock()
}
```
**Problem:** Multiple goroutines read stock=1, all think they can claim

### ❌ DON'T: Use application-level mutex
```go
// WRONG - Only works in single instance
var mu sync.Mutex
mu.Lock()
defer mu.Unlock()
ClaimCoupon()
```
**Problem:** Breaks when you scale to multiple API instances

### ❌ DON'T: Forget FOR UPDATE
```go
// WRONG - No lock!
SELECT * FROM coupons WHERE name = ?
```
**Problem:** No exclusive lock, race conditions occur

### ❌ DON'T: Split transaction
```go
// WRONG - Not atomic!
tx1.Begin()
CheckStock()
tx1.Commit()

tx2.Begin()
InsertClaim()
tx2.Commit()
```
**Problem:** Race condition between transactions

### ✅ DO: Use single atomic transaction with FOR UPDATE
```go
// CORRECT
tx.Begin()
SELECT ... FOR UPDATE  // Lock
CheckStock()
InsertClaim()
DecrementStock()
tx.Commit()  // All or nothing
```

---

## Environment Variables

```bash
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=coupondb

# Server
SERVER_PORT=8080
```

---

## Testing Commands

```bash
# Run all tests
go test ./...

# Run only concurrency tests
go test ./tests -run "Flash|Double"

# Run tests multiple times (verify determinism)
go test ./tests -run "Flash|Double" -count=5

# Run with race detector
go test ./... -race

# Run with verbose output
go test ./... -v
```

---

## Docker Commands

```bash
# Build and start
docker-compose up --build

# Clean start (remove volumes)
docker-compose down -v && docker-compose up --build

# View logs
docker-compose logs -f api

# Execute command in container
docker-compose exec api go test ./tests/...

# Stop services
docker-compose down

# Remove everything
docker-compose down -v --rmi all
```

---

## Quick Verification Checklist

Before considering any component "done":

- [ ] Code compiles without errors
- [ ] Manual testing passes
- [ ] Automated tests pass
- [ ] Concurrency tests are deterministic (run 5+ times)
- [ ] Docker deployment works
- [ ] API returns exact JSON structure per spec
- [ ] HTTP status codes match spec requirements
- [ ] Database has no orphaned data after tests
- [ ] Logs are clean (no error messages)
- [ ] README instructions work exactly as written

---

## Performance Targets

- **Startup time**: < 30 seconds (cold start)
- **API response time**: < 100ms (simple requests)
- **Concurrent claims**: Handle 50+ simultaneous requests
- **Database connections**: Pool of 25 max
- **Transaction duration**: < 10ms (typical)

---

## Success Metrics

The assessment will be evaluated on:

1. ✅ **Correctness**: Both concurrency tests pass 100%
2. ✅ **Code Quality**: Clean, readable, well-structured
3. ✅ **Compliance**: Exact API spec adherence
4. ✅ **Deployability**: docker-compose up works first try
5. ✅ **Documentation**: Clear README with all sections
6. ✅ **Architecture**: Sound concurrency strategy explained

**Critical**: If Flash Sale or Double Dip tests fail, submission fails.
