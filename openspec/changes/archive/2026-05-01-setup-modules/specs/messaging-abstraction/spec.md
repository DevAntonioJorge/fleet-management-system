## ADDED Requirements

### Requirement: Watermill-based message abstraction

The system SHALL use Watermill library as an abstraction layer for message publishing and subscribing.

#### Scenario: Publisher created
- **WHEN** a publisher is requested
- **THEN** a Watermill publisher instance MUST be returned

#### Scenario: Subscriber created
- **WHEN** a subscriber is requested
- **THEN** a Watermill subscriber instance MUST be returned

### Requirement: RabbitMQ as initial broker

The system SHALL use RabbitMQ as the message broker for the MVP.

#### Scenario: RabbitMQ publisher configured
- **WHEN** RabbitMQ config is provided
- **THEN** the publisher MUST connect to RabbitMQ

#### Scenario: RabbitMQ subscriber configured
- **WHEN** RabbitMQ config is provided
- **THEN** the subscriber MUST connect to RabbitMQ

### Requirement: Topic management

The system SHALL support publishing and subscribing to topics.

#### Scenario: Publish to topic
- **WHEN** Publish(topic, payload) is called
- **THEN** the message MUST be sent to the specified topic

#### Scenario: Subscribe to topic
- **WHEN** a subscription to a topic is created
- **THEN** messages from that topic MUST be received

### Requirement: Kafka migration path

The system SHALL support migration from RabbitMQ to Kafka via Watermill.

#### Scenario: Watermill handles broker switch
- **WHEN** broker driver is changed from RabbitMQ to Kafka
- **THEN** no code changes in publisher/subscriber usage MUST be required

#### Scenario: Kafka partition key
- **WHEN** publishing to Kafka
- **THEN** the vehicle_id MUST be used as the partition key

### Requirement: Message format

The system SHALL use a structured message format for telemetry and alerts.

#### Scenario: Telemetry message structure
- **WHEN** telemetry event is published
- **THEN** the message MUST contain vehicle_id, timestamp, lat, lon, speed, fuel

#### Scenario: Alert message structure
- **WHEN** alert event is published
- **THEN** the message MUST contain vehicle_id, alert_type, description, timestamp

### Requirement: Publisher interface

The system SHALL expose a simple publisher interface to services.

#### Scenario: Publish method available
- **WHEN** a service needs to publish
- **THEN** it MUST call publisher.Publish(topic, message)