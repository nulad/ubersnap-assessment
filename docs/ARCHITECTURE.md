# Ubersnap Assessment - Architecture Documentation

## Table of Contents
1. [System Overview](#system-overview)
2. [Database Design](#database-design)
3. [Concurrency Control](#concurrency-control)
4. [Error Handling](#error-handling)
5. [Scalability Considerations](#scalability-considerations)

## System Overview

### High-Level Architecture

The Ubersnap assessment is a coupon management system built with Go that provides RESTful APIs for creating, claiming, and retrieving coupons. The system follows a clean architecture pattern with clear separation of concerns.

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   HTTP Client   │────▶│   HTTP Server   │────▶│   Handler Layer │
│                 │     │   (Gin Router)  │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                        │
                                                        ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   PostgreSQL    │◀────│  Database Layer │◀────│  Service Layer  │
│     Database    │     │   (sql.DB)      │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │ Repository Layer│
                       │                 │
                       │ • CouponRepo    │
                       │ • ClaimRepo     │
                       └─────────────────┘
```

### Component Interaction Flow

1. **HTTP Request Flow**
   - Client sends HTTP request to one of three endpoints
   - Gin router routes request to appropriate handler method
   - Handler validates request and calls service layer
   - Service layer implements business logic using repositories
   - Repository layer executes database operations within transactions
   - Response flows back through the layers with appropriate HTTP status codes

2. **API Endpoints**
   - `POST /api/coupons` - Create a new coupon
   - `POST /api/coupons/claim` - Claim a coupon (critical concurrency endpoint)
   - `GET /api/coupons/{name}` - Retrieve coupon details with claim history

### Technology Stack Rationale

- **Go**: Chosen for performance, strong typing, and excellent concurrency support
- **Gin Web Framework**: Lightweight, high-performance HTTP router with middleware support
- **PostgreSQL**: Reliable relational database with strong ACID compliance and row-level locking
- **lib/pq**: Pure Go PostgreSQL driver with full feature support

## Database Design

### Schema Design Decisions

The database schema consists of two main tables: `coupons` and `claims`. This separation was chosen for:

1. **Normalization**: Avoids redundant data storage
2. **Scalability**: Claim history can grow independently of coupon data
3. **Performance**: Queries can be optimized for specific use cases

### Table Structures

#### Coupons Table
```sql
CREATE TABLE coupons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    amount INTEGER NOT NULL CHECK (amount >= 0),
    remaining_amount INTEGER NOT NULL CHECK (remaining_amount >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

- **Unique name constraint**: Ensures coupon names are unique identifiers
- **Amount vs Remaining Amount**: Tracks total initial amount vs current availability
- **Check constraints**: Enforces business rules at database level

#### Claims Table
```sql
CREATE TABLE claims (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    coupon_name VARCHAR(255) NOT NULL,
    claimed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_coupon UNIQUE (user_id, coupon_name),
    CONSTRAINT fk_coupon FOREIGN KEY (coupon_name) REFERENCES coupons(name) ON DELETE CASCADE
);
```

- **Composite unique constraint**: Prevents duplicate claims by same user for same coupon
- **Foreign key with cascade**: Maintains referential integrity
- **Timestamp**: Tracks when claims were made for audit purposes

### Index Strategy

Strategic indexes optimize common query patterns:
- `idx_coupons_name`: Fast coupon lookups by name
- `idx_claims_user_id`: Efficient user-specific queries
- `idx_claims_coupon_name`: Quick claim retrieval per coupon
- `idx_claims_user_claimed_at`: Optimizes user claim history queries

### Constraint Enforcement

Database-level constraints ensure data integrity:
- **UNIQUE constraints**: Prevent duplicate coupons and claims
- **CHECK constraints**: Validate amounts are non-negative
- **FOREIGN KEY constraints**: Maintain referential integrity
- **Trigger**: Automatically updates `updated_at` timestamp

## Concurrency Control

### Why Pessimistic Locking Over Optimistic

The coupon claiming system uses **pessimistic locking** (`SELECT FOR UPDATE`) instead of optimistic locking for critical reasons:

1. **Race Condition Prevention**: Multiple users claiming the same coupon simultaneously must be prevented
2. **Stock Accuracy**: The remaining amount must never go negative
3. **Business Criticality**: A lost claim results in direct business impact

### SELECT FOR UPDATE Explanation

`SELECT FOR UPDATE` locks the selected rows until the transaction commits or rolls back:

```sql
SELECT id, name, amount, remaining_amount 
FROM coupons 
WHERE name = $1 
FOR UPDATE
```

- **Row-level lock**: Only locks the specific coupon being claimed
- **Blocking**: Other transactions wait for the lock to release
- **Deadlock prevention**: Consistent lock ordering prevents circular waits

### Transaction Isolation Level

The system uses PostgreSQL's default **Read Committed** isolation level:
- **Read Committed**: Prevents dirty reads while allowing good concurrency
- **Transaction Boundaries**: All claim operations are atomic
- **Rollback on Failure**: Any error rolls back the entire transaction

### Race Condition Prevention

The claim process prevents race conditions through:

1. **Atomic Transaction**: All operations succeed or fail together
2. **Row Locking**: Prevents simultaneous modifications
3. **Stock Check Before Decrement**: Verifies availability before claiming
4. **Unique Constraint**: Database enforces no duplicate claims

### Deadlock Avoidance Strategy

Deadlocks are prevented through:
1. **Consistent Lock Order**: Always lock coupons first, then insert claims
2. **Short Transactions**: Minimize lock hold time
3. **No User Input in Transaction**: All validation happens before locking
4. **Retry Logic**: Application layer can retry on deadlock

### Claim Operation Flow Diagram

```
User Request
     │
     ▼
┌─────────────┐
│ Begin TX    │
└─────────────┘
     │
     ▼
┌─────────────┐     ┌─────────────────┐
│ Lock Coupon │────▶│ Wait if locked  │
│ FOR UPDATE  │     │ by other TX     │
└─────────────┘     └─────────────────┘
     │
     ▼
┌─────────────┐
│ Check Stock │
│ > 0?        │
└─────────────┘
     │
     ▼
┌─────────────┐     ┌─────────────────┐
│ Insert Claim│────▶│ Unique Violation│
│ (user, name)│     │ → Already Claimed│
└─────────────┘     └─────────────────┘
     │
     ▼
┌─────────────┐
│ Decrement   │
│ Stock       │
└─────────────┘
     │
     ▼
┌─────────────┐
│ Commit TX   │
└─────────────┘
     │
     ▼
Success Response
```

## Error Handling

### Custom Error Types

The system uses layered error handling with custom error types:

#### Repository Layer Errors
```go
var (
    ErrCouponNotFound     = errors.New("coupon not found")
    ErrCouponExists       = errors.New("coupon already exists")
    ErrInvalidAmount      = errors.New("invalid coupon amount")
    ErrNoStockAvailable   = errors.New("no stock available")
)
```

#### Service Layer Errors
```go
var (
    ErrAlreadyClaimed     = errors.New("user has already claimed this coupon")
    ErrNoStock           = errors.New("no stock available for this coupon")
    ErrCouponNotFound    = errors.New("coupon not found")
)
```

### HTTP Status Code Mapping

Errors are mapped to appropriate HTTP status codes:

- **400 Bad Request**: Invalid input, no stock available
- **404 Not Found**: Coupon doesn't exist
- **409 Conflict**: User already claimed coupon
- **500 Internal Server Error**: Unexpected system errors

### Database Constraint Violation Handling

Database violations are caught and converted to meaningful errors:
- **UNIQUE violations**: Mapped to `ErrCouponExists` or `ErrAlreadyClaimed`
- **CHECK violations**: Mapped to `ErrInvalidAmount`
- **FOREIGN KEY violations**: Mapped to `ErrCouponNotFound`

### Error Propagation Pattern

Errors bubble up through layers with context:
1. **Database**: Raw SQL errors
2. **Repository**: Converts to domain-specific errors
3. **Service**: Adds business logic context
4. **Handler**: Maps to HTTP responses

## Scalability Considerations

### Connection Pooling Configuration

Database connection pooling is configured for optimal performance:
```go
MaxOpenConns: 25     // Maximum concurrent connections
MaxIdleConns: 10     // Idle connections to keep
ConnMaxLifetime: 5m  // Connection reuse duration
```

- **Balanced Pool Size**: Prevents database overload
- **Idle Reuse**: Reduces connection overhead
- **Connection Lifetime**: Prevents stale connections

### Performance Under Load

The system is designed for high concurrency:
1. **Row-level Locking**: Allows concurrent claims on different coupons
2. **Indexed Queries**: Fast lookups even with large datasets
3. **Connection Pooling**: Efficient database resource usage
4. **Minimal Transaction Scope**: Reduces lock contention

### Horizontal Scaling Possibilities

The architecture supports horizontal scaling:

1. **Stateless Application**: Multiple app instances can run behind a load balancer
2. **Database Scaling**:
   - **Read Replicas**: For read-heavy workloads
   - **Partitioning**: By coupon name or date ranges
   - **Sharding**: For massive scale requirements

3. **Caching Layer** (future enhancement):
   - **Redis**: For frequently accessed coupons
   - **Cache Aside Pattern**: Reduces database load

### Bottleneck Analysis

Current and potential bottlenecks:
1. **Database**: Primary bottleneck under extreme load
   - Mitigation: Connection pooling, indexing, read replicas
2. **Single Coupon Contention**: Popular coupons cause lock contention
   - Mitigation: Batch processing, queue-based claiming
3. **Memory Usage**: Claim history can grow large
   - Mitigation: Pagination, archival of old claims

### Monitoring and Metrics

Key metrics to monitor:
- Connection pool utilization
- Transaction duration
- Lock wait times
- Error rates by type
- Concurrent claim attempts

### Future Enhancements

For massive scale, consider:
1. **Message Queue**: Async claim processing
2. **Event Sourcing**: Audit trail and replay capability
3. **CQRS**: Separate read/write models
4. **Microservices**: Domain separation
