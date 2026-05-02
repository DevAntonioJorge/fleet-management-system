## ADDED Requirements

### Requirement: SQLc configuration
The system SHALL use SQLc to generate type-safe database code from SQL queries. SQLc configuration must be centralized at the project root in `sqlc.yaml` and point to queries in `shared/database/sql/`.

#### Scenario: SQLc generates type-safe code from SQL
- **WHEN** `sqlc generate` is executed
- **THEN** type-safe Go code is created in `shared/database/sqlc/` with query functions and models

### Requirement: pgx driver integration
The system SHALL use pgx as the database driver for PostgreSQL. Both `pgxpool.Pool` (for load testing and production) and `pgx.Conn` (for testing and debugging) SHALL be supported through factory functions.

#### Scenario: Connection pool factory
- **WHEN** `shared/database.NewPoolFactory()` is called with a DSN
- **THEN** a `pgxpool.Pool` is created and implements the SQLc Querier interface

#### Scenario: Single connection factory
- **WHEN** `shared/database.NewConnFactory()` is called with a DSN
- **THEN** a `pgx.Conn` is created and implements the SQLc Querier interface

### Requirement: Schema versioning with Atlas
The system SHALL use Atlas to manage PostgreSQL schema versions and automatically generate migrations from a hand-written `schema.sql` file. Migrations SHALL be stored in `shared/database/migrations/` with timestamp-based naming.

#### Scenario: Atlas generates migration from schema changes
- **WHEN** schema.sql is updated and `atlas migrate diff` is run
- **THEN** a new migration file is created in `shared/database/migrations/` with format `YYYYMMDD_NNN_description.sql`

#### Scenario: Migrations are auto-discovered and applied
- **WHEN** the migration script is executed with no arguments
- **THEN** all pending migrations are discovered and applied in order to the database

### Requirement: Three core entities
The system SHALL define schema for Vehicle, TelemetryEvent, and Alert entities as specified in the PRD. Entity definitions SHALL be in `shared/database/sql/schema.sql`.

#### Scenario: Vehicle entity persists
- **WHEN** a vehicle is registered via the API
- **THEN** it is stored in the vehicles table with id, plate, model, and created_at

#### Scenario: TelemetryEvent entity persists
- **WHEN** telemetry is ingested via the API
- **THEN** it is stored in the telemetry_events table with id, vehicle_id, timestamp, latitude, longitude, speed, and fuel_level

#### Scenario: Alert entity persists
- **WHEN** an alert is generated
- **THEN** it is stored in the alerts table with id, vehicle_id, type, description, and timestamp

### Requirement: Query organization
Queries SHALL be organized by domain in `shared/database/sql/<domain>/` (vehicle, telemetry, alert). Each domain directory contains `.sql` files with named queries that SQLc compiles to methods on the Querier interface.

#### Scenario: Queries are organized by domain
- **WHEN** `shared/database/sql/vehicle/vehicle.sql` is inspected
- **THEN** it contains named queries for vehicle operations (create, get, list)

### Requirement: Services depend on generated code through domain adapters
Services (vehicle, telemetry, alert) SHALL NOT depend directly on the SQLc Querier interface. Instead, services SHALL define minimal domain-specific interfaces (e.g., `VehicleQuerier`), and `cmd/api` SHALL adapt the full SQLc Querier to these interfaces during initialization.

#### Scenario: Service uses domain interface
- **WHEN** `vehicle.Service` needs to query the database
- **THEN** it calls methods on its injected `VehicleQuerier` interface, not on the SQLc Querier directly

#### Scenario: cmd/api adapts SQLc Querier
- **WHEN** `cmd/api` initializes services
- **THEN** it creates a single SQLc Querier from the pgx pool and adapts it to service-specific interfaces

### Requirement: Build integration
SQLc code generation SHALL be integrated into the build process. Before building any service, generated code in `shared/database/sqlc/` must be up-to-date.

#### Scenario: SQLc generation is part of build
- **WHEN** the project is built
- **THEN** `sqlc generate` is executed automatically if schema or queries have changed
