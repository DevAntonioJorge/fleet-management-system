## ADDED Requirements

### Requirement: Prometheus metrics endpoint

The system SHALL expose metrics at /metrics endpoint in Prometheus format.

#### Scenario: Metrics endpoint accessible
- **WHEN** a GET request is made to /metrics
- **THEN** the response MUST be valid Prometheus text format

#### Scenario: Metrics endpoint returns 200
- **WHEN** /metrics is queried
- **THEN** the response MUST have HTTP status 200

### Requirement: Per-handler metrics

The system SHALL track metrics at the HTTP handler level.

#### Scenario: Request count per handler
- **WHEN** an HTTP request is made to any endpoint
- **THEN** a counter for that handler MUST be incremented

#### Scenario: Request latency histogram
- **WHEN** an HTTP request completes
- **THEN** a histogram observation for latency MUST be recorded

#### Scenario: Response status codes
- **WHEN** an HTTP response is sent
- **THEN** the status code MUST be tracked (2xx, 4xx, 5xx)

### Requirement: Per-service metrics

The system SHALL track business metrics at the service level.

#### Scenario: Vehicle service metrics
- **WHEN** vehicle operations occur
- **THEN** metrics MUST be recorded (vehicles registered, queries, errors)

#### Scenario: Telemetry service metrics
- **WHEN** telemetry operations occur
- **THEN** metrics MUST be recorded (events processed, events failed)

#### Scenario: Alert service metrics
- **WHEN** alert operations occur
- **THEN** metrics MUST be recorded (alerts generated, alerts queried)

### Requirement: Standard metric names

The system SHALL use consistent naming conventions for metrics.

#### Scenario: HTTP handler metric naming
- **WHEN** metrics are exported
- **THEN** handler metrics MUST follow pattern: http_server_requests_<method>_<path>

#### Scenario: Service metric naming
- **WHEN** metrics are exported
- **THEN** service metrics MUST follow pattern: <service>_<operation>_total

### Requirement: Label support

Metrics SHALL support labels for dimensional data.

#### Scenario: Handler labels
- **WHEN** recording handler metrics
- **THEN** labels MUST include method, path, and status

#### Scenario: Service labels
- **WHEN** recording service metrics
- **THEN** labels MUST include operation and result (success/error)