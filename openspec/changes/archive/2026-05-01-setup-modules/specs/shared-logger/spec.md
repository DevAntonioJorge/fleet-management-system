## ADDED Requirements

### Requirement: Structured logging with slog

The system SHALL provide structured logging based on Go's slog package.

#### Scenario: Logger created with default config
- **WHEN** shared/logger.New() is called without options
- **THEN** a logger instance MUST be returned with JSON output handler

### Requirement: JSON output format

The logger SHALL output logs in JSON format suitable for log aggregation systems.

#### Scenario: Log output is valid JSON
- **WHEN** a log message is written
- **THEN** the output MUST be valid JSON with at least time, level, and msg fields

### Requirement: Debug level with sampling

The logger SHALL support debug-level logging with sampling to reduce volume.

#### Scenario: Debug logging enabled
- **WHEN** logger level is set to debug
- **THEN** debug messages MUST be output

#### Scenario: Debug sampling
- **WHEN** debug sampling is configured with a rate (e.g., 0.1)
- **THEN** approximately 10% of debug messages MUST be output

### Requirement: Contextual logging

The logger SHALL support adding contextual attributes to log entries.

#### Scenario: Adding context to logger
- **WHEN** logger.With(key, value) is called
- **THEN** subsequent log calls MUST include that key-value pair

### Requirement: Log levels

The system SHALL support standard log levels: debug, info, warn, error.

#### Scenario: Info level logging
- **WHEN** logger.Info() is called
- **THEN** the message MUST be logged at info level

#### Scenario: Error level logging
- **WHEN** logger.Error() is called
- **THEN** the message MUST be logged at error level

#### Scenario: Warn level logging
- **WHEN** logger.Warn() is called
- **THEN** the message MUST be logged at warn level