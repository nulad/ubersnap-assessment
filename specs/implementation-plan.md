# Coupon System Implementation Plan

## Overview
Build a high-concurrency REST API in Golang for a Flash Sale Coupon System with strict data consistency guarantees.

## Technology Decisions

### Database Choice: PostgreSQL
**Rationale:**
- Native ACID transaction support for atomic operations
- Row-level locking for concurrency control
- Composite unique constraints for (user_id, coupon_name)
- Better for high-concurrency write scenarios
- Simpler transaction semantics than MongoDB

### Framework: Gin (Web Framework)
**Rationale:**
- Fast, lightweight HTTP router
- Good middleware support
- Easy to test
- Minimal boilerplate

## Architecture

### Project Structure
```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   └── postgres.go          # Database connection setup
│   ├── models/
│   │   ├── coupon.go            # Coupon model
│   │   └── claim.go             # Claim model
│   ├── repository/
│   │   ├── coupon_repo.go       # Coupon data access layer
│   │   └── claim_repo.go        # Claim data access layer
│   ├── service/
│   │   └── coupon_service.go    # Business logic with transaction handling
│   └── handler/
│       └── coupon_handler.go    # HTTP handlers
├── tests/
│   ├── integration_test.go      # Integration tests
│   └── concurrency_test.go      # Flash Sale & Double Dip tests
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## Database Schema Design

### Table: coupons
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

### Table: claims
```sql
CREATE TABLE claims (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    coupon_name VARCHAR(255) NOT NULL,
    claimed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, coupon_name),
    FOREIGN KEY (coupon_name) REFERENCES coupons(name) ON DELETE CASCADE
);

CREATE INDEX idx_claims_user_id ON claims(user_id);
CREATE INDEX idx_claims_coupon_name ON claims(coupon_name);
```

**Key Design Points:**
- UNIQUE constraint on (user_id, coupon_name) prevents duplicate claims at DB level
- Foreign key ensures referential integrity
- Indexes on lookup columns for performance
- CHECK constraints ensure non-negative amounts

## API Implementation Strategy

### 1. POST /api/coupons - Create Coupon
**Flow:**
1. Validate input (name not empty, amount > 0)
2. Insert into coupons table with remaining_amount = amount
3. Return 201 Created on success
4. Return 400 if validation fails or name already exists

**Complexity:** LOW

### 2. POST /api/coupons/claim - Claim Coupon
**Flow (CRITICAL - Atomic Transaction Required):**
```go
BEGIN TRANSACTION;

  // 1. Lock the coupon row for update
  SELECT * FROM coupons WHERE name = ? FOR UPDATE;

  // 2. Check if remaining_amount > 0
  if remaining_amount <= 0 {
    ROLLBACK;
    return 400/409 "No stock available"
  }

  // 3. Try to insert claim (will fail if user already claimed)
  INSERT INTO claims (user_id, coupon_name) VALUES (?, ?);
  // If UNIQUE constraint violation -> user already claimed
  if constraint_violation {
    ROLLBACK;
    return 409 "Already claimed"
  }

  // 4. Decrement remaining_amount
  UPDATE coupons SET remaining_amount = remaining_amount - 1 WHERE name = ?;

COMMIT;
```

**Concurrency Strategy:**
- Use `SELECT ... FOR UPDATE` to lock the coupon row (pessimistic locking)
- This prevents race conditions during stock checks
- Database UNIQUE constraint prevents duplicate claims
- All operations in single transaction = atomic

**Error Handling:**
- Already claimed: Check for UNIQUE constraint violation (pq error code 23505)
- No stock: remaining_amount <= 0
- Return 409 for already claimed (preferred)
- Return 400 for no stock

**Complexity:** HIGH

### 3. GET /api/coupons/{name} - Get Coupon Details
**Flow:**
1. Query coupon by name
2. Query all claims for this coupon
3. Aggregate user_ids into array
4. Return JSON response

**Performance Consideration:**
- Use JOIN or separate queries
- For large claim lists, consider pagination (not required but good practice)

**Complexity:** MEDIUM

## Concurrency Testing Scenarios

### Test 1: Flash Sale Attack
```
Setup: Create coupon with amount = 5
Execute: 50 concurrent goroutines claiming the same coupon (50 different users)
Expected: Exactly 5 success (200/201), 45 failures (400), remaining_amount = 0
Verify: Count claims in database = 5
```

### Test 2: Double Dip Attack
```
Setup: Create coupon with amount = 100
Execute: 10 concurrent goroutines claiming with SAME user_id
Expected: Exactly 1 success (200/201), 9 failures (409), remaining_amount = 99
Verify: Count claims for this user = 1
```

**Test Implementation:**
- Use `sync.WaitGroup` to coordinate goroutines
- Collect all response codes
- Assert exact counts match expectations

## Docker Configuration

### Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /api .
EXPOSE 8080
CMD ["./api"]
```

### docker-compose.yml
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
        condition: service_healthy

volumes:
  postgres_data:
```

## Implementation Order

1. **Setup Phase**
   - Initialize Go module
   - Set up project structure
   - Add dependencies (gin, pq/pgx)

2. **Database Layer**
   - Database connection setup with retry logic
   - Migration scripts for table creation
   - Repository pattern implementation

3. **Service Layer**
   - Implement transaction-wrapped claim logic
   - Error handling for constraint violations
   - Business logic validation

4. **Handler Layer**
   - Route setup
   - Request/response DTOs
   - HTTP error code mapping

5. **Docker Setup**
   - Dockerfile multi-stage build
   - docker-compose.yml with healthchecks
   - Environment variable configuration

6. **Testing**
   - Unit tests for business logic
   - Integration tests for API endpoints
   - Concurrency stress tests

7. **Documentation**
   - README with setup instructions
   - Architecture notes
   - API documentation

## Key Success Criteria

✅ Atomic claim operations (no race conditions)
✅ UNIQUE constraint enforces one claim per user per coupon
✅ SELECT FOR UPDATE prevents overselling
✅ Flash Sale test: exactly 5 claims for 5-item stock
✅ Double Dip test: exactly 1 claim for same user
✅ Proper HTTP status codes (409 for already claimed, 400 for no stock)
✅ Docker Compose one-command startup
✅ Clean separation: Coupons and Claims in separate tables

## Potential Pitfalls to Avoid

❌ Don't use optimistic locking (check-then-act pattern) - race conditions
❌ Don't embed claims in coupon document - violates spec
❌ Don't forget FOR UPDATE lock - causes overselling
❌ Don't handle uniqueness in app code only - race conditions
❌ Don't forget to test actual concurrency scenarios
❌ Don't return wrong HTTP status codes - automated tests will fail

## Performance Considerations

- Database connection pooling (tune max connections)
- Indexes on frequently queried columns
- Transaction timeout configuration
- Consider READ COMMITTED isolation level (PostgreSQL default)
- Monitor lock contention under high load

## Timeline Estimate

- Setup & Database Layer: 2-3 hours
- Service & Handler Layer: 2-3 hours
- Docker Configuration: 1 hour
- Testing & Debugging: 2-3 hours
- Documentation: 1 hour

**Total: 8-11 hours of focused development**
