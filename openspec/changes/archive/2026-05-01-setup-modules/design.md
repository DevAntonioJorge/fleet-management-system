## Context

The Fleet Management System (FMS) is a greenfield Go project that requires a modular structure from the start. Based on PRD.md, the system will handle vehicle telemetry with async processing via message broker. The architecture needs to support high data ingestion, observability, and future scalability.

Current state: No code exists - empty project directory.
Target: Multi-module Go workspace using go.work.

## Goals / Non-Goals

**Goals:**
- Establish go.work workspace with separate modules for services
- Create clear directory boundaries: cmd/, services/, shared/, internal/ (auth only - using golang-jwt)
- Implement shared libraries for logging, metrics, and config
- Set up Watermill messaging abstraction with RabbitMQ driver
- Enable direct import between services (same process)

**Non-Goals:**
- Implement actual business logic (vehicle, telemetry, alert services)
- Database schema or SQLc generation
- Kafka migration (only prepare abstraction)
- Load testing implementation
- Custom crypto implementation (using golang-jwt for auth)

## Decisions

### D1: go.work over single module

**Decision**: Use go.work to create a multi-module workspace.

**Rationale**: Enables independent development/testing of services, cleaner dependency boundaries, and easier future extraction to separate repos.

**Alternatives considered**:
- Single module with packages: Simpler initially, but harder to enforce boundaries
- Monorepo with separate repos: Overkill for MVP, adds deployment complexity

### D2: Directory structure per capability

**Decision**: Separate services into their own modules under services/ directory.

```
services/
├── vehicle/
│   ├── go.mod
│   └── service.go
├── telemetry/
│   └── ...
```

**Rationale**: Clear ownership boundaries, each service can be tested independently, future migration to microservices path.

**Alternatives considered**:
- Single services/ package: Less isolation, harder to extract later

### D3: Shared libraries as separate modules

**Decision**: Each shared component (logger, metrics, config) is its own module under shared/.

**Rationale**: Avoids circular dependencies, allows services to import only what they need, explicit interface boundaries.

### D4: Watermill for messaging abstraction

**Decision**: Use Watermill library for message broker abstraction.

**Rationale**:
- Provides unified interface for RabbitMQ → Kafka migration
- Built-in support for pub/sub, request/response patterns
- Structured logging integration
- Active maintenance, well-documented

**Alternatives considered**:
- Direct RabbitMQ client: Simpler but harder to migrate
- Custom abstraction: More control but additional maintenance burden
- go-kafka: Only Kafka, no abstraction

### D5: slog as logging foundation

**Decision**: Wrap slog with additional utilities (JSON output, contextual logging).

**Rationale**:
- Built into Go stdlib (1.21+)
- Structured by design
- Easy Prometheus log scraping integration

**Alternatives considered**:
- zerolog/zap: More features but additional dependency
- Custom wrapper over stdlog: Reinventing functionality

### D6: Prometheus metrics with both per-handler and per-service granularity

**Decision**: Expose /metrics endpoint with counters/histograms at handler and service levels.

**Rationale**:
- Handler level: Request latency, status codes, request size
- Service level: Business metrics (vehicles registered, telemetry processed)
- Prometheus standard format for easy Grafana integration

## Risks / Trade-offs

- **[Risk] Over-engineering for MVP** → Start minimal, add complexity only when needed
- **[Risk] Module dependency conflicts** → Use explicit version constraints in go.work
- **[Risk] Watermill migration complexity** → Keep broker-specific code isolated, minimize direct usage
- **[Risk] go.work compatibility** → Test with different Go versions, use replace directives if needed
- **[Trade-off] More files/modules** → Offset by clearer boundaries and better testability
- **[Trade-off] Debug logging with sampling** → May miss important info during debugging → Adjust sampling rate as needed

## Migration Plan

1. Create go.work and root go.mod
2. Create shared/* modules with minimal implementations
3. Create service modules (stub implementations)
4. Create cmd/api with basic HTTP server
5. Add Watermill publisher/subscriber setup
6. Verify all imports resolve correctly
7. Run go work sync to validate

## Open Questions

- What specific logger levels/formats for production vs development?
- How to handle service health checks beyond /metrics?
- Configuration file format: env vars, YAML, or both?