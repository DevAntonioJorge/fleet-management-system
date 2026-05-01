## 1. Dependencies

- [x] 1.1 Add `github.com/ThreeDotsLabs/watermill` to shared module
- [x] 1.2 Add `github.com/ThreeDotsLabs/watermill-amqp/v2` to shared module
- [x] 1.3 Run `go mod tidy` in shared/ and services/

## 2. Watermill Publisher Implementation

- [x] 2.1 Create `WatermillPublisher` struct wrapping `amqp.Publisher`
- [x] 2.2 Implement `Publish(ctx, topic, payload, metadata)` method
- [x] 2.3 Convert metadata to Watermill `message.Metadata`
- [x] 2.4 Handle context cancellation in Publish
- [x] 2.5 Implement `Close()` method

## 3. Watermill Subscriber Implementation

- [x] 3.1 Create `WatermillSubscriber` struct wrapping `amqp.Subscriber`
- [x] 3.2 Implement `Subscribe(ctx, topic, handler)` method
- [x] 3.3 Convert Watermill messages to payload/metadata
- [x] 3.4 Call handler and Ack/Nack based on result
- [x] 3.5 Implement `Close()` method
- [x] 3.6 Handle context cancellation in Subscribe

## 4. Factory Updates

- [x] 4.1 Update `RabbitMQPublisherFactory.CreatePublisher` to return `WatermillPublisher`
- [x] 4.2 Configure AMQP publisher with exchange settings (topic type, durable)
- [x] 4.3 Update `RabbitMQSubscriberFactory.CreateSubscriber` to return `WatermillSubscriber`
- [x] 4.4 Configure AMQP subscriber with queue settings (durable, binding to topic)

## 5. Configuration

- [x] 5.1 Add RabbitMQ URL to config struct
- [x] 5.2 Add exchange and queue name configuration
- [x] 5.3 Update factory creation to use config values

## 6. Integration

- [x] 6.1 Update `PublishTelemetry` to work with new publisher
- [x] 6.2 Update `PublishAlert` to work with new publisher
- [x] 6.3 Update `CreateTelemetryHandler` to work with new subscriber
- [x] 6.4 Update `CreateAlertHandler` to work with new subscriber

## 7. Testing

- [x] 7.1 Write unit tests for `WatermillPublisher` (uses gochannel for unit tests)
- [x] 7.2 Write unit tests for `WatermillSubscriber` (uses gochannel for unit tests)
- [x] 7.3 Write integration test with local RabbitMQ (Docker) (run with `-tags=integration`)
- [x] 7.4 Verify end-to-end: publish → consume → process (covered in integration test)
