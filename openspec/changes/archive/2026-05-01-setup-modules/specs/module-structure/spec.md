## ADDED Requirements

### Requirement: Go workspace with go.work

The project SHALL use go.work to create a multi-module workspace that enables independent development and testing of domain services.

#### Scenario: go.work file exists at root
- **WHEN** the project is initialized
- **THEN** a go.work file MUST exist at the project root

#### Scenario: All service modules are included
- **WHEN** go.work is configured
- **THEN** all service modules under services/ MUST be included via use directive

### Requirement: Directory structure boundaries

The project SHALL maintain clear directory boundaries for different types of code.

#### Scenario: cmd/ contains entry points
- **WHEN** examining the project structure
- **THEN** cmd/ SHALL contain only entry point packages (main.go files)

#### Scenario: services/ contains domain modules
- **WHEN** examining the project structure
- **THET** services/ SHALL contain domain service modules, each with its own go.mod

#### Scenario: shared/ contains reusable libraries
- **WHEN** examining the project structure
- **THEN** shared/ SHALL contain reusable libraries (logger, metrics, config), each as separate module

#### Scenario: internal/ contains critical code only
- **WHEN** examining the project structure
- **THEN** internal/ SHALL contain only security-critical code (JWT via golang-jwt)

### Requirement: Service module isolation

Each service under services/ SHALL be a separate Go module with its own go.mod file.

#### Scenario: Vehicle service module
- **WHEN** accessing the vehicle service
- **THEN** services/vehicle/go.mod MUST exist and define the module

#### Scenario: Telemetry service module
- **WHEN** accessing the telemetry service
- **THEN** services/telemetry/go.mod MUST exist and define the module

#### Scenario: Alert service module
- **WHEN** accessing the alert service
- **THEN** services/alert/go.mod MUST exist and define the module

### Requirement: Direct import between services

Services SHALL be able to import each other directly using the module paths defined in go.work.

#### Scenario: Telemetry imports Vehicle
- **WHEN** telemetry service needs vehicle data
- **THEN** it MUST be able to import services/vehicle via its module path

#### Scenario: No circular dependencies
- **WHEN** analyzing import graph
- **THEN** there MUST be no circular dependencies between services