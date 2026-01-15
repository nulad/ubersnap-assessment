# Flash Sale Coupon System

A scalable REST API built with Golang for handling high-concurrency coupon claims with strict data consistency guarantees. This system ensures that coupons cannot be over-claimed during flash sales and prevents duplicate claims by the same user.

## Table of Contents

- [Prerequisites](#prerequisites)
- [How to Run](#how-to-run)
- [How to Test](#how-to-test)
- [API Documentation](#api-documentation)
- [Architecture Notes](#architecture-notes)
- [Project Structure](#project-structure)

## Prerequisites

Before running this application, ensure you have the following installed:

- **Docker Desktop**: Version 20.10 or higher
- **Docker Compose**: Version 2.0 or higher (included with Docker Desktop)
- **Available Ports**:
  - Port 8080 for the API server
  - Port 5432 for PostgreSQL database

You can verify Docker installation with:
```bash
docker --version
docker-compose --version
```

## How to Run

1. Clone the repository:
```bash
git clone <repository-url>
cd ubersnap-assessment
```

2. Start the application using Docker Compose:
```bash
docker-compose up --build
```

3. Wait for the startup message:
```
Server listening on :8080
```

The API will be available at `http://localhost:8080/api`

To stop the application, press `Ctrl+C` or run:
```bash
docker-compose down
```

## How to Test

### Manual Testing with curl

#### Create a Coupon
```bash
curl -X POST http://localhost:8080/api/coupons \
  -H "Content-Type: application/json" \
  -d '{"name": "PROMO_SUPER", "amount": 100}'
```
Expected response: `201 Created`

#### Claim a Coupon
```bash
curl -X POST http://localhost:8080/api/coupons/claim \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user_12345", "coupon_name": "PROMO_SUPER"}'
```
Expected response: `200 OK` with message `{"message": "Coupon claimed successfully"}`

#### Get Coupon Details
```bash
curl http://localhost:8080/api/coupons/PROMO_SUPER
```
Expected response: `200 OK` with JSON:
```json
{
  "name": "PROMO_SUPER",
  "amount": 100,
  "remaining_amount": 99,
  "claimed_by": ["user_12345"]
}
```

### Quick API Demo Script

Run the included test script to see all endpoints in action:
```bash
./scripts/test-api.sh
```

This script will:
1. Create a new coupon with 10 units
2. Claim it with user1 (should succeed)
3. Try to claim again with user1 (should fail with 409 Conflict)
4. Claim with user2 (should succeed)
5. Display the final coupon state showing 8 remaining units

### Automated Tests

Run the full test suite with Docker Compose:
```bash
docker-compose exec api go test ./tests/... -v
```

#### Expected Test Results

**Flash Sale Test** (50 concurrent requests, 5 stock):
- Creates coupon "FLASH" with 5 items
- Spawns 50 concurrent goroutines attempting to claim
- Expected result: Exactly 5 successes, 45 failures
- Remaining amount: 0
- Claimed by: 5 unique users

**Double Dip Test** (10 concurrent requests, same user):
- Creates coupon with 100 items
- Spawns 10 concurrent goroutines with SAME user_id
- Expected result: Exactly 1 success, 9 conflicts (409)
- Remaining amount: 99
- Claimed by: 1 user only

## API Documentation

### 1. Create Coupon

Creates a new coupon in the system.

**Endpoint:** `POST /api/coupons`

**Request Body:**
```json
{
  "name": "PROMO_SUPER",
  "amount": 100
}
```

**Response Codes:**
- `201 Created` - Coupon created successfully
- `400 Bad Request` - Invalid request (duplicate name or invalid amount)
- `500 Internal Server Error` - Server error

---

### 2. Claim Coupon

Attempts to claim a coupon for a specific user. This endpoint is designed for high-concurrency scenarios.

**Endpoint:** `POST /api/coupons/claim`

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "user_id": "user_12345",
  "coupon_name": "PROMO_SUPER"
}
```

**Response Codes:**
- `200 OK` - Claim successful
- `400 Bad Request` - Invalid request or no stock available
- `409 Conflict` - User already claimed this coupon
- `404 Not Found` - Coupon does not exist
- `500 Internal Server Error` - Server error

**Success Response:**
```json
{
  "message": "Coupon claimed successfully"
}
```

**Error Response Examples:**
```json
{
  "error": "user has already claimed this coupon"
}
```
```json
{
  "error": "no stock available for this coupon"
}
```

---

### 3. Get Coupon Details

Retrieves detailed information about a coupon including who has claimed it.

**Endpoint:** `GET /api/coupons/{name}`

**Response Body:**
```json
{
  "name": "PROMO_SUPER",
  "amount": 100,
  "remaining_amount": 95,
  "claimed_by": ["user_1", "user_2", "user_3", "user_4", "user_5"]
}
```

**Response Codes:**
- `200 OK` - Coupon details retrieved successfully
- `404 Not Found` - Coupon does not exist
- `500 Internal Server Error` - Server error

## Architecture Notes

### Database Design

The system uses **PostgreSQL** with two separate tables:

#### Schema Overview

**coupons table:**
```sql
- id (SERIAL PRIMARY KEY)
- name (VARCHAR 255, UNIQUE)
- amount (INTEGER, CHECK >= 0)
- remaining_amount (INTEGER, CHECK >= 0)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

**claims table:**
```sql
- id (SERIAL PRIMARY KEY)
- user_id (VARCHAR 255)
- coupon_name (VARCHAR 255, FOREIGN KEY)
- claimed_at (TIMESTAMP)
- UNIQUE CONSTRAINT (user_id, coupon_name)
```

**Key Design Decisions:**

1. **Separation of Concerns**: Coupons and claims are stored in separate tables to maintain normalization and query flexibility.

2. **UNIQUE Constraint**: The `(user_id, coupon_name)` composite unique constraint at the database level prevents duplicate claims even under high concurrency.

3. **Foreign Key with CASCADE**: Claims reference coupon names with CASCADE delete to maintain referential integrity.

### Concurrency Strategy

The system handles race conditions through a combination of database-level and application-level mechanisms:

#### Transaction Flow

```
BEGIN TRANSACTION
    ↓
1. SELECT ... FOR UPDATE (Lock coupon row)
    ↓
2. Check remaining_amount > 0
    ↓
3. INSERT INTO claims (Enforced by UNIQUE constraint)
    ↓
4. UPDATE coupons SET remaining_amount = remaining_amount - 1
    ↓
COMMIT TRANSACTION
```

#### Concurrency Mechanisms

1. **SELECT FOR UPDATE**: Row-level pessimistic locking ensures only one transaction can modify a coupon at a time. Other transactions wait until the lock is released.

2. **UNIQUE Constraint**: Database-enforced uniqueness on `(user_id, coupon_name)` guarantees no duplicate claims, even if concurrent transactions pass the application check.

3. **Atomic Transactions**: All operations (check stock, insert claim, decrement stock) happen within a single ACID transaction. Either all succeed or all fail.

4. **Transaction Isolation**: PostgreSQL's default isolation level (Read Committed) is sufficient because SELECT FOR UPDATE provides stronger guarantees for the locked rows.

### Why PostgreSQL?

PostgreSQL was chosen for this system because:

1. **Row-Level Locking**: Native support for SELECT FOR UPDATE with robust locking mechanisms
2. **ACID Compliance**: Strong transactional guarantees prevent data inconsistencies
3. **UNIQUE Constraints**: Database-enforced uniqueness provides an additional safety layer
4. **Performance**: Excellent performance for concurrent writes with proper indexing
5. **Reliability**: Battle-tested in high-concurrency environments

### Error Handling

The application implements comprehensive error handling:

- **Repository Layer**: Returns domain-specific errors (ErrCouponNotFound, ErrAlreadyClaimed, etc.)
- **Service Layer**: Maps repository errors to business logic errors
- **Handler Layer**: Converts service errors to appropriate HTTP status codes
- **Transaction Rollback**: Automatic rollback on any error using defer pattern

### Performance Considerations

- **Indexes**: Strategic indexes on frequently queried columns (coupon name, user_id)
- **Connection Pooling**: Database connection pool configured for optimal throughput
- **Graceful Shutdown**: Server handles SIGTERM/SIGINT for clean database connection closure

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go           # Configuration loader
│   │   └── config_test.go
│   ├── database/
│   │   ├── database.go         # Database connection management
│   │   ├── claim_repository.go # Claims data access layer
│   │   └── ...
│   ├── dto/
│   │   └── dto.go              # Request/response data transfer objects
│   ├── handler/
│   │   ├── coupon_handler.go   # HTTP request handlers
│   │   └── coupon_handler_test.go
│   ├── models/
│   │   └── models.go           # Domain models (Coupon, Claim)
│   ├── repository/
│   │   ├── coupon.go           # Coupon repository implementation
│   │   ├── errors.go           # Repository error definitions
│   │   └── ...
│   └── service/
│       ├── coupon_service.go   # Business logic layer
│       ├── errors.go           # Service error definitions
│       └── coupon_service_test.go
├── tests/
│   ├── flash_sale_test.go      # Flash sale concurrency test
│   └── setup_test.go
├── database/
│   └── schema.sql              # PostgreSQL schema definition
├── docker-compose.yml           # Docker orchestration
├── Dockerfile                   # API container definition
└── README.md                    # This file
```

### Architecture Layers

1. **Handler Layer** (`internal/handler/`): HTTP request handling, request validation, response formatting
2. **Service Layer** (`internal/service/`): Business logic, transaction management, error handling
3. **Repository Layer** (`internal/repository/`, `internal/database/`): Data access, SQL queries
4. **Models** (`internal/models/`): Domain entities
5. **DTO** (`internal/dto/`): API contract definitions

This layered architecture ensures:
- Clear separation of concerns
- Easy testability at each layer
- Flexibility to change implementations
- Maintainable and scalable codebase

---

**Built with Go 1.23, PostgreSQL 15, Docker, and Gin Web Framework**
