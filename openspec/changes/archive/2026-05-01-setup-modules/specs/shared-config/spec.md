## ADDED Requirements

### Requirement: Configuration from environment

The system SHALL load configuration from environment variables with sensible defaults.

#### Scenario: Required config loaded
- **WHEN** required environment variables are set
- **THEN** the config MUST be available in the application

#### Scenario: Default values applied
- **WHEN** optional environment variables are not set
- **THEN** default values MUST be used

#### Scenario: Config validation
- **WHEN** required config is missing
- **THEN** the application MUST fail to start with clear error message

### Requirement: Database configuration

The system SHALL support database connection configuration.

#### Scenario: PostgreSQL connection string
- **WHEN** DATABASE_URL is provided
- **THEN** the connection string MUST be parsed and available

#### Scenario: Database connection pool
- **WHEN** config includes pool settings
- **THEN** the database pool MUST be configured accordingly

### Requirement: Message broker configuration

The system SHALL support message broker connection settings.

#### Scenario: RabbitMQ connection config
- **WHEN** RABBITMQ_URL is provided
- **THEN** the connection parameters MUST be available

#### Scenario: Kafka bootstrap servers
- **WHEN** KAFKA_BROKERS is provided
- **THEN** the bootstrap servers list MUST be available for future migration

### Requirement: Application configuration

The system SHALL support application-level settings.

#### Scenario: Server port configuration
- **WHEN** APP_PORT is provided
- **THEN** the server MUST listen on that port

#### Scenario: Log level configuration
- **WHEN** LOG_LEVEL is provided
- **THEN** the logger MUST be configured with that level

### Requirement: Configuration struct

The system SHALL provide a typed config struct for all settings.

#### Scenario: Config struct available
- **WHEN** config is loaded
- **THEN** a typed Config struct MUST be accessible to the application

#### Scenario: Config passed to services
- **WHEN** services are initialized
- **THEN** they MUST receive the config struct