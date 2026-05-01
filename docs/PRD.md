### PRD — Fleet Management and Telemetry System

Version: 1.0.0
Last Update: 04/30/26

#### 1. Overview
The Fleet Management and Telemetry System aims to monitor vehicles in real-time, collect telemetry data, and provide operational insights through REST APIs.
The system will be designed with a focus on:
*   High data ingestion
*   Asynchronous processing
*   Observability and scalability
*   Architecture aligned with real-market systems

#### 2. Objectives
**Main Objectives**
*   Monitor vehicle location in real-time
*   Store historical telemetry data
*   Detect relevant events (e.g., speeding)
*   Provide data via REST API

**Technical Objectives**
*   Demonstrate the use of messaging (Kafka)
*   Handle high write volumes
*   Apply architecture best practices (Modular Monolith)
*   Utilize efficient data access (SQLc)

#### 3. Scope
**Included (MVP)**
*   Vehicle registration
*   Telemetry event ingestion
*   Asynchronous processing via message broker
*   Persistence in a relational database
*   Historical data queries
*   Last position retrieval
*   Basic alert system

**Out of Scope (Future)**
*   Full frontend interface
*   Advanced authentication (OAuth, RBAC)
*   Machine learning for predictive analysis
*   Load testing tool (decision deferred)

#### 4. Architecture
**Style**
*   Modular Monolith with go.work (Multi-module workspace)
*   Direct import between services (same process)

**Structure**
```
FMS/
├── go.work                    # Workspace root
├── cmd/                       # Entry points
│   ├── api/                   # REST API server
│   └── load-test/            # Load testing tool
├── services/                  # Domain services (separate modules)
│   ├── vehicle/
│   ├── telemetry/
│   └── alert/
├── shared/                    # Shared libraries
│   ├── logger/                # slog wrapper
│   ├── metrics/               # Prometheus metrics
│   └── config/                # Configuration
└── internal/                  # Critical code only (JWT, crypto)
```

**Components**
*   REST API (Ingestion + Queries)
*   Event Producer (Telemetry)
*   Event Consumer
*   PostgreSQL Database
*   Message Broker (RabbitMQ → Kafka via Watermill abstraction)

**Data Flow**
1. Client sends telemetry event via HTTP.
2. API publishes the event to the broker (via Watermill).
3. Consumer processes the event.
4. Data is persisted in the database.
5. Business rules generate alerts (in same consumer pipeline).

#### 5. Domain
**Entities**
*   **Vehicle**
    *   id (UUID)
    *   plate
    *   model
    *   created_at
*   **TelemetryEvent**
    *   id
    *   vehicle_id
    *   timestamp
    *   latitude
    *   longitude
    *   speed
    *   fuel_level
*   **Alert**
    *   id
    *   vehicle_id
    *   type
    *   description
    *   timestamp

#### 6. Functional Requirements
*   **RF01 — Vehicle Registration:** The system must allow vehicle registration.
*   **RF02 — Telemetry Ingestion:** The system must receive telemetry data via REST API.
*   **RF03 — Asynchronous Processing:** Events must be processed via a message broker.
*   **RF04 — Persistence:** Data must be stored in PostgreSQL.
*   **RF05 — Historical Query:** Users must be able to query telemetry by time period.
*   **RF06 — Last Location:** Users must be able to retrieve the vehicle's last known position.
*   **RF07 — Alerts:** The system must detect:
    *   Speeding
    *   (Optional) Geofence exit

#### 7. Non-Functional Requirements
*   **RNF01 — Performance:** Support high event rates.
*   **RNF02 — Scalability:** Use database partitioning and partition by vehicle_id.
*   **RNF03 — Consistency:** Ensure idempotency in the consumer.
*   **RNF04 — Observability:**
    *   Logger: slog with JSON output, debug level with sampling
    *   Metrics: Prometheus, endpoint at `/metrics`
    *   Granularity: both per-handler and per-service

#### 8. API (Summary)
*   **Vehicles**
    *   `POST /vehicles`
    *   `GET /vehicles`
*   **Telemetry**
    *   `POST /telemetry`
    *   `GET /vehicles/{id}/telemetry`
*   **Tracking**
    *   `GET /vehicles/{id}/location`
*   **Alerts**
    *   `GET /alerts`

#### 9. Data Modeling
**Strategy**
*   Telemetry table partitioned by time.
*   Indexes on `vehicle_id` and `timestamp`.

#### 10. Messaging
**Library**
*   Watermill (message broker abstraction)
*   Initial: RabbitMQ
*   Target: Kafka (migration path planned)

**Topics**
*   Phase 1 (by layer): `raw`, `processed`, `alerts`
*   Phase 2 (by capability): `vehicle.created`, `telemetry.ingested`, `alert.speeding`

**Strategies**
*   **Partition key:** `vehicle_id` (ensures ordering per vehicle)
*   Consumer group for parallelism.
*   Idempotent processing.

#### 11. Load Testing (Portfolio Focus)
*Load testing approach to be decided. Will replace simulator to demonstrate performance and metrics in the portfolio.*

#### 12. Success Criteria
*   The system successfully processes telemetry events asynchronously.
*   APIs respond correctly to queries.
*   Alerts are generated according to rules.
*   The project demonstrates proficiency in:
    *   Concurrency in Go
    *   Messaging
    *   Data modeling
    *   Architecture

#### 13. Tech Stack
*   **Go** (Backend)
*   **go.work** (Multi-module workspace)
*   **Watermill** (Messaging abstraction)
    *   Initial: RabbitMQ
    *   Target: Kafka
*   **PostgreSQL** (Database)
*   **SQLc** (Data Access)
*   **slog** (Structured logging)
*   **Prometheus + Grafana** (Observability)
*   **Docker** (Local Environment)

#### 15. Risks
*   Unnecessary complexity with Kafka without sufficient volume.
*   Lack of data to demonstrate system value.
*   Overengineering the modular monolith.
