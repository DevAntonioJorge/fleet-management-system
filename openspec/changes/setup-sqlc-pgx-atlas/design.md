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
- Containerize Go services (all services run on host for now)

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

### Decision 7: Docker Compose for local development
**Choice:** Add `docker-compose.yml` running `postgres:16-alpine` with persistent volume for interactive development  
**Alternatives Considered:**
- Manual PostgreSQL install: OS-specific, harder to clean up, inconsistent environments
- Docker run command: Loses configuration, requires re-typing flags each time
- Testcontainers only: Great for tests but no persistent dev database

**Rationale:** Developers need a persistent PostgreSQL instance for running the API, testing migrations manually, and browsing data. Docker Compose provides a reproducible, one-command setup. `postgres:16-alpine` is chosen for compatibility with `gen_random_uuid()`, `TIMESTAMPTZ`, and small image footprint.

**Implementation:**
```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    ports: ["5432:5432"]
    environment:
      POSTGRES_DB: fleet_dev
      POSTGRES_USER: fleet
      POSTGRES_PASSWORD: fleet
    volumes:
      - pgdata:/var/lib/postgresql/data
volumes:
  pgdata:
```
- Default DSN: `postgres://fleet:fleet@localhost:5432/fleet_dev`
- `docker compose up -d` starts it, `docker compose down` stops it
- Data persists across restarts via named volume

### Decision 8: testcontainers-go for integration tests
**Choice:** Use `github.com/testcontainers/testcontainers-go` with one container per test package via `TestMain`, execute `schema.sql` directly (not Atlas CLI) for speed  
**Alternatives Considered:**
- Fresh container per test: Maximum isolation but ~4s overhead per test (too slow)
- Container reuse (testcontainers `WithReuse`): Fastest (~0.5s after first run) but risk of stale state between test runs
- Mocked interfaces only: Fast but doesn't verify SQL queries or schema correctness
- docker-compose shared DB for tests: No isolation between tests, state leakage

**Rationale:** One container per test package via `TestMain` is the best balance — ~3s startup for the entire package, not per test. Transaction rollback per test ensures isolation. Direct `schema.sql` execution avoids spawning the Atlas CLI, making schema setup faster and simpler.

**Implementation:**
```
shared/database/
├── database_test.go         ← TestMain + individual tests
└── testutil/
    └── testdb.go            ← Shared testcontainers helper
```

`testutil.StartTestDB(ctx)` → starts `postgres:16-alpine` container, returns container handle + DSN

`testutil.ApplySchema(dsn)` → reads `schema.sql`, executes all DDL in a single transaction

`database_test.go` TestMain flow:
1. `container, dsn := testutil.StartTestDB(ctx)`
2. `testutil.ApplySchema(dsn)`
3. `pool := pgxpool.New(ctx, dsn)`
4. `m.Run()`
5. Cleanup: `pool.Close()`, `container.Terminate(ctx)`

Individual test isolation via transaction rollback:
```go
func TestPoolFactory_CreatesValidPool(t *testing.T) {
    tx, _ := pool.Begin(ctx)
    defer tx.Rollback(ctx)  // all changes undone
    // test logic using tx
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

**[Risk] testcontainers requires Docker daemon running locally**  
→ Mitigation: Document Docker as a developer prerequisite; developers without Docker can fall back to docker-compose + manual testing

**[Risk] testcontainers startup time adds to test duration**  
→ Mitigation: One container per test package (~3s total), not per test; individual tests use tx rollback for fast isolation

**[Risk] Direct schema.sql execution in tests doesn't validate Atlas migration flow**  
→ Mitigation: Atlas migration is verified manually (task 3.x) and via migrate.sh; test schema execution is a separate concern focused on speed

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
5. Add `docker-compose.yml` for local development
6. Add `shared/database/testutil/` with testcontainers helper
7. Write integration tests for database layer (TestMain pattern, tx rollback)

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

- PostgreSQL version? (Resolved: `postgres:16-alpine` for dev and tests)
- Connection pool size for load testing? (Defer to load-test configuration)
- Do we use UUIDs or serial IDs for vehicle/alert? (Defer to first service implementation)
- Exact columns for TelemetryEvent (fuel_level units, format)? (Defer to telemetry service spec)
