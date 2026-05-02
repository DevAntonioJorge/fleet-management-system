## Why

The PRD commits to SQLc for "efficient data access" but the project currently has no database layer. Without SQLc, pgx, and schema management, we cannot implement any of the three core services (vehicle, telemetry, alert). Setting this up now unblocks all business logic and establishes the data access pattern for the entire system.

## What Changes

- Add `sqlc.yaml` at project root to configure SQLc for code generation
- Create `shared/database/` module with connection factories (pool for load testing, conn for debugging/testing)
- Define schema for three domain entities: Vehicle, TelemetryEvent, Alert
- Set up Atlas for schema versioning and automatic migration generation
- Implement domain error types that wrap pgx errors
- Services will depend on generated SQLc querier interface through domain-specific adapters
- Create migration script for manual schema application

## Capabilities

### New Capabilities
- `database-layer`: Establish SQLc + pgx + Atlas foundation with schema, migrations, and connection factories
- `database-error-handling`: Domain error types wrapping pgx errors for clean error propagation

### Modified Capabilities

(None - this is foundational work with no requirement changes to existing specs)

## Impact

- **New module**: `shared/database/` becomes required dependency for all services
- **cmd/api**: Will initialize pgx pool and inject into services
- **Services (vehicle, telemetry, alert)**: Will import SQLc-generated code through domain interfaces
- **Build process**: SQLc generation must run before building (integrate into Makefile or CI)
- **Development workflow**: Schema changes require running Atlas and migration scripts
