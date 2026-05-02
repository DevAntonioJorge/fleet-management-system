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

## 10. Integration Tests

- [ ] 10.1 Create `shared/database/database_test.go` with connection factory tests
- [ ] 10.2 Add test database setup (Docker or test container)
- [ ] 10.3 Test pool factory creates valid connection pool
- [ ] 10.4 Test conn factory creates valid single connection
- [ ] 10.5 Test error wrapping for common pgx errors

## 11. Build Integration

- [x] 11.1 Create Makefile target `make generate` to run `sqlc generate`
- [ ] 11.2 Add pre-build step to CI to ensure generated code is current
- [x] 11.3 Add .gitignore entries for generated files (or commit them based on preference)

## 12. Documentation

- [x] 12.1 Create `shared/database/README.md` documenting usage patterns
- [x] 12.2 Add examples for using pool vs conn factories
- [x] 12.3 Document how to add new queries and regenerate
- [x] 12.4 Document migration workflow (writing schema, generating migrations, applying)

## 13. cmd/api Integration

- [ ] 13.1 Update `cmd/api/main.go` to initialize database pool from config
- [ ] 13.2 Create database adapters for service interfaces (VehicleQuerier, etc.)
- [ ] 13.3 Inject adapted querier into service constructors
- [ ] 13.4 Add error handling middleware to map domain errors to HTTP status codes

## 14. Verification

- [x] 14.1 Run `sqlc generate` and verify no errors
- [ ] 14.2 Run integration tests and verify all pass
- [x] 14.3 Verify cmd/api compiles with database layer integrated
- [ ] 14.4 Test connection factory with live PostgreSQL instance
