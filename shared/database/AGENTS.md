# Shared Database Module

This module provides the database foundation for FMS using SQLc, pgx, and Atlas.

## Architecture

```
shared/database/
├── sql/
│   ├── schema.sql           # Desired schema (hand-written)
│   └── queries/             # SQLc query files
│       ├── vehicle.sql
│       ├── telemetry.sql
│       └── alert.sql
├── migrations/
│   ├── atlas.hcl           # Atlas configuration
│   └── *.sql               # Migration files
├── sqlc/                    # Generated code (do not edit)
│   ├── models.go           # Database models
│   ├── db.go               # Queries struct and DBTX interface
│   └── *.sql.go            # Query implementations
├── database.go             # Connection factories
├── errors.go               # Domain error types
└── migrate.sh              # Migration script
```

## Connection Factories

### Pool Factory (Production/Load Testing)

For scenarios requiring connection pooling:

```go
ctx := context.Background()
pool, err := database.NewPoolFactory(ctx, dsn, 25) // 25 max connections
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

queries := sqlc.New(pool)
vehicles, err := queries.GetAllVehicles(ctx)
```

### Single Connection Factory (Testing/Debugging)

For unit tests with transaction rollback:

```go
ctx := context.Background()
conn, err := database.NewConnFactory(ctx, dsn)
if err != nil {
    log.Fatal(err)
}
defer conn.Close(ctx)

// Use in transactions
tx, err := conn.Begin(ctx)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback(ctx) // Automatic rollback in tests

queries := sqlc.New(tx)
err = queries.CreateVehicle(ctx, ...)
```

## Usage Pattern: Domain Adapters

Services should NOT depend directly on SQLc-generated code. Instead:

1. **Services define domain interfaces**:

```go
// services/vehicle/service.go
package vehicle

type Querier interface {
    CreateVehicle(ctx context.Context, plate, model string) (id string, err error)
    GetVehicle(ctx context.Context, id string) (*Vehicle, error)
    GetAllVehicles(ctx context.Context) ([]*Vehicle, error)
}

type Service struct {
    querier Querier
}
```

2. **cmd/api adapts SQLc Querier to domain interfaces**:

```go
// cmd/api/main.go
package main

pool, err := database.NewPoolFactory(ctx, cfg.DatabaseURL, 25)
queries := sqlc.New(pool)

// Create service with adapted querier
vehicleService := vehicle.NewService(&VehicleQuerierAdapter{queries})
```

3. **Error Handling**:

Queries return pgx errors, which should be wrapped in domain errors:

```go
vehicle, err := queries.GetVehicleByID(ctx, id)
if err != nil {
    return nil, database.WrapError(err)
}

// In handlers:
if database.IsNotFound(err) {
    w.WriteHeader(http.StatusNotFound)
    return
}
```

## Schema Management

### Writing Queries

Queries are in `sql/queries/<domain>/<query-file>.sql` using SQLc syntax:

```sql
-- name: GetVehicleByID :one
SELECT id, plate, model, created_at 
FROM vehicles 
WHERE id = $1;

-- name: CreateVehicle :exec
INSERT INTO vehicles (id, plate, model, created_at) 
VALUES ($1, $2, $3, $4);
```

Query directives:
- `:one` - returns single row
- `:many` - returns multiple rows
- `:exec` - executes without returning rows

### Regenerating Code

After writing or modifying queries:

```bash
cd /home/aj/codes/www/FMS
sqlc generate
```

### Managing Schema Changes

1. **Update `schema.sql`** with your desired schema
2. **Generate migration** (when database is available):
   ```bash
   cd shared/database/migrations
   atlas migrate diff --env local my_description
   ```
3. **Or manually create migration file** with format: `YYYYMMDD_NNN_description.sql`
4. **Apply migrations**:
   ```bash
   shared/database/migrate.sh up "postgres://user:pass@localhost/db"
   ```

## Testing

For unit tests with transaction rollback:

```go
// test_db.go
func setupTestDB(t *testing.T) (pgx.Conn, context.Context) {
    ctx := context.Background()
    conn, err := database.NewConnFactory(ctx, testDSN)
    if err != nil {
        t.Fatalf("failed to create test connection: %v", err)
    }
    return conn, ctx
}

func TestGetVehicle(t *testing.T) {
    conn, ctx := setupTestDB(t)
    defer conn.Close(ctx)
    
    tx, _ := conn.Begin(ctx)
    defer tx.Rollback(ctx)
    
    queries := sqlc.New(tx)
    // Test with rollback
}
```

## Common Errors

### "record not found"
Queries return `pgx.ErrNoRows` which should be wrapped:
```go
vehicle, err := queries.GetVehicleByID(ctx, id)
if database.IsNotFound(database.WrapError(err)) {
    // Handle not found
}
```

### Connection refused
Ensure PostgreSQL is running and DSN is correct:
```
postgres://[user[:password]@]localhost[:5432]/[database]?sslmode=disable
```

### Stale generated code
After modifying queries, always run `sqlc generate` before building.

## Local Development

### Starting PostgreSQL

Run the following command to start a local PostgreSQL instance:

```bash
docker compose up -d
```

This starts `postgres:16-alpine` on port 5432 with:
- Database: `fleet_dev`
- User: `fleet`
- Password: `fleet`

To stop:

```bash
docker compose down
```

To view logs:

```bash
docker compose logs -f postgres
```

Connection string: `postgres://fleet:fleet@localhost:5432/fleet_dev?sslmode=disable`

### Running Integration Tests

Integration tests use testcontainers to spin up a fresh PostgreSQL container per test package.

**Prerequisites:**
- Docker must be running
- Go 1.26+ installed

**Run all tests:**

```bash
go test ./shared/database/...
```

**Run with verbose output:**

```bash
go test -v ./shared/database/...
```

Tests take ~5 seconds to run (includes container startup). Each test package shares one container via `TestMain`, and individual tests use transaction rollback for isolation.

## See Also

- [SQLc Documentation](https://docs.sqlc.dev)
- [pgx Documentation](https://github.com/jackc/pgx)
- [Atlas Documentation](https://atlasgo.io)
