## Why

The Fleet Management System requires asynchronous event processing via a message broker. The PRD specifies Watermill as the abstraction layer with RabbitMQ for MVP and Kafka as the target. Currently, only `InMemoryPublisher` and `InMemorySubscriber` exist as placeholders—RabbitMQ and Kafka implementations are stubbed out with "not implemented" errors.

## What Changes

- Implement Watermill-based `Publisher` and `Subscriber` using RabbitMQ as the initial broker
- Add `github.com/ThreeDotsLabs/watermill` and `github.com/ThreeDotsLabs/watermill-amqp/v2` (RabbitMQ) dependencies
- Replace `RabbitMQPublisherFactory` and `RabbitMQSubscriberFactory` stubs with working implementations
- Update `shared/messaging` package to use Watermill's `message.Message` abstraction
- Add configuration support for broker URL and exchange/queue settings
- Prepare architecture for future Kafka migration (Watermill makes this a drop-in replacement)

## Capabilities

### New Capabilities
- `watermill-rabbitmq`: RabbitMQ publisher and subscriber implementation using Watermill, including connection management, topic routing, and message serialization

### Modified Capabilities
<!-- No existing specs to modify -->

## Impact

- **Code**: `shared/messaging/publisher.go`, `shared/messaging/subscriber.go`
- **Dependencies**: New Go modules (`watermill`, `watermill-amqp`)
- **Configuration**: New broker settings in config (URL, exchange, queue names)
- **Services**: `services/telemetry`, `services/alert` (consumers will use new subscriber)
- **Documentation**: Update PRD alignment, add broker setup instructions
