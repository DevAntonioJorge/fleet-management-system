## Why

The Fleet Management System needs a modular Go project structure using go.work to enable independent development and testing of domain services while maintaining a cohesive architecture aligned with industry standards.

## What Changes

- Create go.work workspace with multiple modules
- Establish cmd/, services/, shared/, internal/ directory structure
- Add shared libraries for logger, metrics, and config
- Create initial service modules (vehicle, telemetry, alert)
- Set up Watermill messaging abstraction with RabbitMQ

## Capabilities

### New Capabilities

- **module-structure**: Define the multi-module workspace structure with go.work, establishing clear boundaries between entry points (cmd/), domain services (services/), shared libraries (shared/), and critical code (internal/).
- **shared-logger**: Implement slog-based structured logging wrapper with JSON output and debug-level sampling support.
- **shared-metrics**: Create Prometheus metrics infrastructure with per-handler and per-service granularity, exposed at /metrics endpoint.
- **shared-config**: Build configuration management for environment-based settings.
- **messaging-abstraction**: Implement Watermill-based message broker abstraction supporting RabbitMQ initially with Kafka migration path.

## Impact

- New directories: cmd/, services/, shared/, internal/
- New files: go.work, multiple go.mod files per service
- Dependencies: Watermill, Prometheus client, slog-compatible logger