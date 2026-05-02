## ADDED Requirements

### Requirement: Domain error types wrap pgx errors
The system SHALL define domain-specific error types in `shared/database/errors.go` that wrap pgx errors for clean error propagation. Services SHALL NOT expose pgx error types directly.

#### Scenario: NotFound error wraps pgx.ErrNoRows
- **WHEN** a query returns no results
- **THEN** a domain error (e.g., `ErrVehicleNotFound`) is returned instead of pgx.ErrNoRows

#### Scenario: Database errors are wrapped
- **WHEN** a pgx error occurs (connection, constraint, etc.)
- **THEN** it is wrapped in a domain error type (e.g., `ErrDatabase`) with the original error accessible via `errors.Unwrap()`

### Requirement: Error handling consistency
All services (vehicle, telemetry, alert) SHALL use the same domain error types from `shared/database/errors`. This enables consistent error handling in `cmd/api` request handlers.

#### Scenario: Service returns domain error
- **WHEN** `vehicle.Service.RegisterVehicle()` fails due to database constraint
- **THEN** it returns a domain error, not a pgx error

#### Scenario: cmd/api maps domain errors to HTTP status codes
- **WHEN** a handler receives a domain error from a service
- **THEN** it maps the error to an appropriate HTTP status code (404 for NotFound, 400 for validation, 500 for database)
