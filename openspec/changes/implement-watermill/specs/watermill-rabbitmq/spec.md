## ADDED Requirements

### Requirement: Watermill RabbitMQ publisher
The system SHALL use Watermill with RabbitMQ to publish messages to topics.

#### Scenario: Publish message to RabbitMQ
- **WHEN** a publisher publishes a message to a topic
- **THEN** the message MUST be delivered to RabbitMQ with the correct topic routing key

#### Scenario: Publish with metadata
- **WHEN** a publisher publishes a message with metadata
- **THEN** the metadata MUST be stored as Watermill message metadata

#### Scenario: Publisher handles connection failure
- **WHEN** RabbitMQ is unavailable during publish
- **THEN** the publisher MUST return an error without panicking

### Requirement: Watermill RabbitMQ subscriber
The system SHALL use Watermill with RabbitMQ to subscribe to topics and process messages.

#### Scenario: Subscribe to topic
- **WHEN** a subscriber subscribes to a topic with a handler
- **THEN** the handler MUST be called for each message received on that topic

#### Scenario: Message acknowledgment
- **WHEN** a handler successfully processes a message
- **THEN** the message MUST be acknowledged (Ack) to RabbitMQ

#### Scenario: Message negative acknowledgment
- **WHEN** a handler returns an error during message processing
- **THEN** the message MUST be negatively acknowledged (Nack) for retry

#### Scenario: Subscriber handles connection failure
- **WHEN** RabbitMQ connection is lost
- **THEN** the subscriber MUST attempt to reconnect or return an error

### Requirement: Watermill message abstraction
The system SHALL use Watermill's `message.Message` type for message representation.

#### Scenario: Convert to Watermill message
- **WHEN** publishing a payload with metadata
- **THEN** the system MUST create a `message.Message` with payload as `Payload` and metadata as `Metadata`

#### Scenario: Convert from Watermill message
- **WHEN** processing a received Watermill message
- **THEN** the system MUST extract `Payload` and `Metadata` for the handler

### Requirement: Factory pattern for broker selection
The system SHALL use factory patterns to create broker-specific publishers and subscribers.

#### Scenario: Create RabbitMQ publisher
- **WHEN** `RabbitMQPublisherFactory.CreatePublisher` is called with URL and exchange
- **THEN** a Watermill RabbitMQ publisher MUST be returned

#### Scenario: Create RabbitMQ subscriber
- **WHEN** `RabbitMQSubscriberFactory.CreateSubscriber` is called with URL and queue
- **THEN** a Watermill RabbitMQ subscriber MUST be returned

### Requirement: Configuration support
The system SHALL support configuration for RabbitMQ connection settings.

#### Scenario: Configure broker URL
- **WHEN** the system starts with broker configuration
- **THEN** the publisher and subscriber MUST connect using the configured URL

#### Scenario: Configure exchange and queue
- **WHEN** the system starts with exchange/queue configuration
- **THEN** the publisher MUST use the configured exchange and subscriber MUST use the configured queue
