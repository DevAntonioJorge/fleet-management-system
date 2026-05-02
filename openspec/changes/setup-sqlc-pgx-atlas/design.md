## Context

The FMS project is a modular monolith for fleet telemetry built on go.work. Services (vehicle, telemetry, alert) currently have domain interfaces but no database layer. The PRD mandates SQLc for type safety and efficient queries. Three core entities (Vehicle, TelemetryEvent, Alert) must persist in PostgreSQL with eventual partitioning of telemetry by time.

## Goals / Non-Goals

**Goals:**
- Establish SQLc + pgx + Atlas as the database foundation for all services
- Enable both pool-based (load testing/production) and connection-based (debug/testing) database access
- Automate schema versioning and migration generation with Atlas
- Implement clean domain error types that wrap pgx errors
- Create clear adapters between SQLc-generated code and service domain logic
- Unblock implementation of vehicle, telemetry, and alert services

**Non-Goals:**
- Implement telemetry table partitioning (deferred to Phase 2)
- Add advanced ORM features or query builders
- Optimize schema for specific query patterns (optimize incrementally based on load)
- Deploy or configure production PostgreSQL instance (assumes postgres is running)

## Decisions

### Decision 1: SQLc for type-safe code generation
**Choice:** Use SQLc to generate Go code from hand-written SQL queries  
**Alternatives Considered:**
- GORM: Higher abstraction, easier to write but less control over queries, larger overhead, less efficient for high-volume operations
- sqlc: Type-safe, minimal runtime overhead, explicit SQL, generates exactly what we write

**Rationale:** The PRD requires high-efficiency data access. SQLc gives us type safety without the abstraction penalty. SQL queries are explicit and reviewable. Load testing with high volumes means every millisecond matters.

**Implementation:**
- Central `sqlc.yaml` at project root
- Queries organized in `shared/database/sql/<domain>/` (vehicle, telemetry, alert)
- Generated code in `shared/database/sqlc/`
- Services import generated code only through domain interfaces

### Decision 2: Dual connection model (pool + single conn)
**Choice:** Support both `pgxpool.Pool` and `pgx.Conn` via factory functions  
**Alternatives Considered:**
- Pool only: Simpler but impossible to debug with tx rollback in tests
- Conn only: Doesn't scale for load testing

**Rationale:** Load testing requires connection pooling for throughput. Testing and debugging benefit from single connections with transaction rollback. Both satisfy the SQLc Querier interface.

**Implementation:**
- `shared/database.NewPoolFactory(dsn, poolSize)` → pgxpool.Pool
- `shared/database.NewConnFactory(dsn)` → pgx.Conn
- cmd/api uses pool, tests use conn, load-test can use either

### Decision 3: Atlas for schema versioning
**Choice:** Hand-write schema.sql, use Atlas to auto-generate migrations  
**Alternatives Considered:**
- write raw SQL migrations: More control but manual, error-prone
- Atlas HCL: Tight coupling to Atlas tool, less portable
- Liquibase: Java-based, adds dependency

**Rationale:** Hand-written schema is readable and portable. Atlas generates diffs automatically. Timestamp-based migrations auto-sort correctly.

**Implementation:**
- Write desired schema in `shared/database/sql/schema.sql`
- Run `atlas migrate diff --env local` to generate migration
- Auto-discovery script applies pending migrations in order

### Decision 4: Domain interfaces decouple services from SQLc
**Choice:** Services define minimal domain interfaces (`VehicleQuerier`), cmd/api adapts SQLc Querier  
**Alternatives Considered:**
- Services import SQLc Querier directly: Tightly coupled to generated code, harder to test
- Monolithic service interface: Too broad, loses type safety

**Rationale:** Isolates services from SQLc changes. Services are testable with mock implementations. cmd/api adapter is the only place that knows about SQLc.

**Implementation:**
```go
// vehicle/service.go
type VehicleQuerier interface {
    CreateVehicle(ctx, vehicle) error
    GetVehicle(ctx, id) (*Vehicle, error)
    GetAllVehicles(ctx) ([]*Vehicle, error)
}

// cmd/api/main.go
querier := sqlc.New(pool)
vehicleService := vehicle.NewService(&QueuerAdapter{querier})
```

### Decision 5: Domain error types wrap pgx errors
**Choice:** Define custom error types in `shared/database/errors.go`, wrap pgx errors  
**Alternatives Considered:**
- Return pgx errors directly: Couples services to pgx, harder to change database drivers
- Custom errors without wrapping: Loses error context, harder to debug

**Rationale:** Clean error boundaries. Services and cmd/api deal with domain errors. Original pgx error available via `errors.Unwrap()` for debugging.

**Implementation:**
```go
// shared/database/errors.go
type ErrVehicleNotFound struct{ Err error }
type ErrDatabase struct{ Err error }

// vehicle service
if err := repo.GetVehicle(...); err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, &ErrVehicleNotFound{Err: err}
    }
    return nil, &ErrDatabase{Err: err}
}
```

### Decision 6: Build integration for SQLc generation
**Choice:** Integrate `sqlc generate` into build process (Makefile/CI)  
**Alternatives Considered:**
- Manual generation before each commit: Error-prone, inconsistent
- Generated code in git: Bloats repo, merge conflicts

**Rationale:** Ensures generated code is always up-to-date. Prevents "works on my machine" issues.

**Implementation:**
- Makefile target: `make generate` runs `sqlc generate`
- CI checks that generated code matches (prevents committing stale code)

## Risks / Trade-offs

**[Risk] SQLc SQL syntax errors caught at generation time, not runtime**  
→ Mitigation: Review SQL carefully, integrate into CI, run tests locally before push

**[Risk] Schema changes require running Atlas + migration scripts**  
→ Mitigation: Automate with scripts, document in README, make it easy for developers

**[Risk] Connection pooling adds complexity to testing**  
→ Mitigation: NewConnFactory for unit tests with tx rollback, clear documentation on when to use which

**[Risk] Dual pool/conn model could lead to inconsistent behavior**  
→ Mitigation: Both implement same Querier interface, tests ensure consistency

**[Risk] Atlas may not handle all PostgreSQL features perfectly**  
→ Mitigation: Start with basic schema, escalate complex changes to manual SQL when needed

**[Trade-off] More modules/files to manage (sqlc, migrations, errors)**  
→ Benefit: Cleaner separation of concerns, easier to maintain and evolve

## Migration Plan

### Phase 1: Foundation (this change)
1. Set up `shared/database/` with sqlc config, schema, factories
2. Create initial schema for Vehicle, TelemetryEvent, Alert
3. Generate initial migration with Atlas
4. Implement domain error types
5. Write integration tests for database layer

### Phase 2: Service integration (next change)
1. Define SQLc queries for each domain
2. Implement domain adapters in cmd/api
3. Integrate services with database layer
4. Add service tests

### Future: Optimization
- Add indexes based on load testing results
- Implement telemetry table partitioning by time
- Consider read replicas if needed

## Open Questions

- PostgreSQL version? (Assuming 14+, needed for JSON/UUIDs)
- Connection pool size for load testing? (Defer to load-test configuration)
- Do we use UUIDs or serial IDs for vehicle/alert? (Defer to first service implementation)
- Exact columns for TelemetryEvent (fuel_level units, format)? (Defer to telemetry service spec)
