# Database Package

This package provides database connection management with connection pooling, retry logic, and migrations for the Ubersnap assessment project.

## Features

- **Connection Pooling**: Configurable connection pool with proper limits
- **Retry Logic**: Automatic retry with exponential backoff for initial connections
- **Migrations**: Automatic schema migration execution
- **Graceful Shutdown**: Proper connection cleanup on application shutdown
- **Health Checks**: Built-in health check functionality
- **Context Support**: Context-aware database operations

## Configuration

The database connection is configured through environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL server host |
| `DB_PORT` | 5432 | PostgreSQL server port |
| `DB_USER` | postgres | Database username |
| `DB_PASSWORD` | (empty) | Database password |
| `DB_NAME` | ubersnap | Database name |
| `DB_SSLMODE` | disable | SSL mode for connection |
| `DB_MAX_OPEN_CONNS` | 25 | Maximum open connections |
| `DB_MAX_IDLE_CONNS` | 10 | Maximum idle connections |
| `DB_CONN_MAX_LIFETIME` | 5m | Connection maximum lifetime |

## Usage

### Basic Usage

```go
package main

import (
    "log"
    "github.com/nulad/ubersnap-assessment/internal/database"
)

func main() {
    // Initialize database with migrations
    db, err := database.InitDatabase()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Use the database
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM coupons").Scan(&count)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Total coupons: %d", count)
}
```

### Context-Aware Operations

```go
ctx := context.Background()
sqlCtx := db.WithContext(ctx)

rows, err := sqlCtx.QueryContext("SELECT * FROM coupons WHERE remaining_amount > $1", 0)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()
```

### Health Check

```go
if err := db.Health(); err != nil {
    log.Printf("Database unhealthy: %v", err)
} else {
    log.Println("Database is healthy")
}
```

### Connection Pool Statistics

```go
stats := db.Stats()
log.Printf("Open connections: %d", stats.OpenConnections)
log.Printf("In use: %d", stats.InUse)
log.Printf("Idle: %d", stats.Idle)
```

## Running with Docker

Start PostgreSQL using Docker Compose:

```bash
docker-compose up -d postgres
```

Then run the application:

```bash
export DB_HOST=localhost
export DB_PASSWORD=postgres
export DB_NAME=coupondb
go run cmd/server/main.go
```

## Migration

The package automatically runs migrations from `database/schema.sql` on initialization. The migration:

- Creates tables with proper constraints
- Adds indexes for performance
- Sets up triggers for timestamp updates
- Is idempotent (safe to run multiple times)

## Testing

Run the tests:

```bash
go test ./internal/database/...
```

The tests cover:
- Configuration loading from environment
- Connection string generation
- Default value handling
