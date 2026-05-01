## 1. Workspace Setup

- [x] 1.1 Create root go.mod with module name github.com/fms/fms
- [x] 1.2 Create go.work with use directives for all modules
- [x] 1.3 Create cmd/ directory structure (cmd/api, cmd/load-test)
- [x] 1.4 Create services/ directory structure (services/vehicle, services/telemetry, services/alert)
- [x] 1.5 Create shared/ directory structure (shared/logger, shared/metrics, shared/config)
- [x] 1.6 Create internal/ directory structure (internal/auth only - using golang-jwt)

## 2. Service Modules

- [x] 2.1 Create services/vehicle/go.mod with module path
- [x] 2.2 Create services/vehicle/service.go with minimal stub (interfaces only)
- [x] 2.3 Create services/telemetry/go.mod with module path
- [x] 2.4 Create services/telemetry/service.go with minimal stub (interfaces only)
- [x] 2.5 Create services/alert/go.mod with module path
- [x] 2.6 Create services/alert/service.go with minimal stub (interfaces only)

## 3. Shared Libraries

- [x] 3.1 Create shared/logger/go.mod
- [x] 3.2 Implement shared/logger/logger.go with New(), With(), Debug(), Info(), Warn(), Error()
- [x] 3.3 Create shared/metrics/go.mod
- [x] 3.4 Implement shared/metrics/metrics.go with NewMetrics(), Handler(), service metrics
- [x] 3.5 Create shared/config/go.mod
- [x] 3.6 Implement shared/config/config.go with Config struct and Load() function
- [x] 3.7 Add tests for shared libraries

## 4. Message Broker Abstraction

- [x] 4.1 Add Watermill dependency to shared/messaging/go.mod (create module)
- [x] 4.2 Implement shared/messaging/publisher.go with Publisher interface
- [x] 4.3 Implement shared/messaging/subscriber.go with Subscriber interface
- [x] 4.4 Add RabbitMQ driver configuration
- [x] 4.5 Create message types for telemetry and alerts

## 5. API Entry Point

- [x] 5.1 Create cmd/api/main.go with basic HTTP server
- [x] 5.2 Add router with /metrics endpoint
- [x] 5.3 Integrate shared/logger in main
- [x] 5.4 Integrate shared/metrics middleware
- [x] 5.5 Integrate shared/config loading
- [x] 5.6 Add basic health check endpoint (GET /health)

## 6. Validation

- [x] 6.1 Run go work sync to verify all modules
- [x] 6.2 Run go build for cmd/api to verify compilation
- [x] 6.3 Run go test for shared libraries
- [x] 6.4 Verify /metrics endpoint responds
- [x] 6.5 Verify logger outputs JSON format