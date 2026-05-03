## 1. Directory and Module Setup

- [x] 1.1 Create `shared/database/` directory structure with `sql/`, `migrations/`, and `sqlc/` subdirs
- [x] 1.2 Create `shared/database/go.mod` with pgx and atlas dependencies
- [x] 1.3 Update root `go.work` to include `shared/database` module

## 2. Schema Definition

- [x] 2.1 Create `shared/database/sql/schema.sql` with Vehicle, TelemetryEvent, Alert table definitions
- [x] 2.2 Add indexes on vehicle_id and timestamp in schema.sql for query performance
- [x] 2.3 Define foreign key relationships (TelemetryEvent and Alert reference Vehicle)

## 3. Atlas Configuration and Migration Generation

- [x] 3.1 Create `shared/database/migrations/atlas.hcl` to configure Atlas for local environment
- [x] 3.2 Run `atlas migrate diff --env local` to generate initial migration file
- [x] 3.3 Verify migration file creates correct schema in `shared/database/migrations/`

## 4. SQLc Configuration

- [x] 4.1 Create `sqlc.yaml` at project root with correct paths and PostgreSQL engine
- [x] 4.2 Verify sqlc.yaml paths point to `./shared/database/sql/` for queries

## 5. Query SQL Files

- [x] 5.1 Create `shared/database/sql/vehicle/vehicle.sql` with queries (create, getByID, getAll)
- [x] 5.2 Create `shared/database/sql/telemetry/telemetry.sql` with queries (create, getByVehicleID, getLastPosition)
- [x] 5.3 Create `shared/database/sql/alert/alert.sql` with queries (create, getByVehicleID, getAll)

## 6. SQLc Code Generation

- [x] 6.1 Run `sqlc generate` to create type-safe Go code in `shared/database/sqlc/`
- [x] 6.2 Verify generated files (models.go, db.go, and domain-specific files)
- [x] 6.3 Ensure generated Querier interface contains all query methods

## 7. Database Connection Factories

- [x] 7.1 Create `shared/database/database.go` with `NewPoolFactory()` function
- [x] 7.2 Implement `NewConnFactory()` function for single connections
- [x] 7.3 Add DSN parsing and connection validation
- [x] 7.4 Add connection health checks (ping on creation)

## 8. Domain Error Types

- [x] 8.1 Create `shared/database/errors.go` with domain error types (NotFound, Database, Constraint)
- [x] 8.2 Implement error wrapping that preserves pgx errors via `errors.Unwrap()`
- [x] 8.3 Add helper functions for error type checking (e.g., `IsNotFound()`)

## 9. Migration Script

- [x] 9.1 Create `shared/database/migrate.sh` script to auto-discover and apply migrations
- [x] 9.2 Implement logic to detect pending migrations in `shared/database/migrations/`
- [x] 9.3 Add rollback capability (optional, defer if needed)
- [x] 9.4 Make script executable and test with sample migrations

## 10. Docker Compose (Development)

- [x] 10.1 Create `docker-compose.yml` with `postgres:16-alpine`, port 5432, persistent volume
- [x] 10.2 Add `.env` file with default database credentials (fleet/fleet/fleet_dev)
- [x] 10.3 Verify `docker compose up -d` starts PostgreSQL and accepts connections

## 11. Integration Test Infrastructure

- [x] 11.1 Add `testcontainers-go` and `testcontainers-go/modules/postgres` to `shared/database/go.mod`
- [x] 11.2 Create `shared/database/testutil/testdb.go` with `StartTestDB(ctx)` helper
- [x] 11.3 Create `shared/database/testutil/schema.go` with `ApplySchema(dsn)` helper (reads schema.sql, executes in transaction)
- [x] 11.4 Create `shared/database/database_test.go` with `TestMain` that spins up container, applies schema, runs tests, tears down
- [x] 11.5 Verify container is destroyed after tests complete

## 12. Integration Tests

- [x] 12.1 Test `NewPoolFactory()` creates valid connection pool with ping
- [x] 12.2 Test `NewConnFactory()` creates valid single connection with ping
- [x] 12.3 Test error wrapping for `pgx.ErrNoRows` → `ErrNotFound`
- [x] 12.4 Test error wrapping for constraint violations → `ErrConstraint`
- [x] 12.5 Test error wrapping for generic database errors → `ErrDatabase`

## 13. Build Integration

- [x] 13.1 Create Makefile target `make generate` to run `sqlc generate`
- [x] 13.2 Add .gitignore entries for generated files (or commit them based on preference)

## 14. Documentation

- [x] 14.1 Create `shared/database/README.md` documenting usage patterns
- [x] 14.2 Add examples for using pool vs conn factories
- [x] 14.3 Document how to add new queries and regenerate
- [x] 14.4 Document migration workflow (writing schema, generating migrations, applying)
- [x] 14.5 Document `docker compose up -d` as the standard way to start local PostgreSQL
- [x] 14.6 Document how to run integration tests and Docker prerequisites

## 15. cmd/api Integration

- [x] 15.1 Update `cmd/api/main.go` to initialize database pool from config
- [x] 15.2 Create database adapters for service interfaces (VehicleQuerier, etc.)
- [x] 15.3 Inject adapted querier into service constructors
- [x] 15.4 Add error handling middleware to map domain errors to HTTP status codes

## 15. Verification

- [x] 15.1 Run `sqlc generate` and verify no errors
- [x] 15.2 Run `docker compose up -d` and verify PostgreSQL starts
- [x] 15.3 Run integration tests (`go test ./shared/database/...`) and verify all pass
- [x] 15.4 Verify cmd/api compiles with database layer integrated
- [x] 15.5 Test connection factory with live PostgreSQL instance (via docker-compose)
