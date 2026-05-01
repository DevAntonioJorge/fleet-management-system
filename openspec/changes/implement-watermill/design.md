## Context

The Fleet Management System (FMS) requires asynchronous event processing via a message broker. The PRD specifies Watermill as the abstraction layer with RabbitMQ for MVP and Kafka as the future target.

Current state:
- `shared/messaging/publisher.go` has `Publisher` interface and `InMemoryPublisher` (working)
- `shared/messaging/subscriber.go` has `Subscriber` interface and `InMemorySubscriber` (working)
- `RabbitMQPublisherFactory` and `RabbitMQSubscriberFactory` are stubs returning "not implemented" errors
- The system uses topics: `raw`, `processed`, `alerts` (per PRD section 10)

Constraints:
- Must maintain backward compatibility with existing `Publisher` and `Subscriber` interfaces
- Must support the factory pattern already in place
- Watermill provides broker abstraction—migrating to Kafka later should be a config change

## Goals / Non-Goals

**Goals:**
- Implement working RabbitMQ publisher using Watermill
- Implement working RabbitMQ subscriber using Watermill
- Maintain existing `Publisher`/`Subscriber` interface contracts
- Support metadata propagation via Watermill message metadata
- Enable easy migration to Kafka (leverage Watermill abstraction)

**Non-Goals:**
- Kafka implementation (separate change)
- Message replay or dead letter queues (out of scope for MVP)
- Exactly-once delivery semantics (at-least-once is sufficient)

## Decisions

### Decision 1: Use Watermill AMQP package for RabbitMQ

**Choice**: `github.com/ThreeDotsLabs/watermill-amqp/v2`

**Rationale**:
- Official Watermill-supported RabbitMQ/AMQP implementation
- Handles connection pooling, reconnection, and channel management
- Provides `amqp.Publisher` and `amqp.Subscriber` that implement Watermill's `Publisher`/`Subscriber` interfaces

**Alternatives considered**:
- Direct `github.com/streadway/amqp` (RabbitMQ Go client): Would require manual Watermill adapter, defeats purpose of using Watermill
- `github.com/wagslane/go-rabbitmq`: Not Watermill-compatible, would break abstraction goal

### Decision 2: Wrap Watermill publisher/subscriber to match existing interfaces

**Choice**: Create `WatermillPublisher` and `WatermillSubscriber` structs that wrap Watermill's types and implement the existing `Publisher`/`Subscriber` interfaces.

**Rationale**:
- Existing code uses `messaging.Publisher` interface with signature `Publish(ctx, topic, payload, metadata)`
- Watermill's `message.Publisher` has signature `Publish(topic string, messages ...*message.Message) error`
- Wrapper adapts Watermill's API to our interface, keeping consumers unchanged

**Implementation approach**:
```go
type WatermillPublisher struct {
    pub *amqp.Publisher
}

func (p *WatermillPublisher) Publish(ctx context.Context, topic string, payload []byte, metadata map[string]string) error {
    msg := message.NewMessage(watermill.NewUUID(), payload)
    for k, v := range metadata {
        msg.Metadata.Set(k, v)
    }
    return p.pub.Publish(topic, msg)
}
```

### Decision 3: Connection configuration via factory

**Choice**: Pass AMQP URL and exchange/queue names to factory, create Watermill config with sensible defaults.

**Rationale**:
- Factories already accept `url` and `exchange`/`queue` parameters
- Watermill AMQP config supports customizing exchange type, queue durability, etc.
- Keep config simple for MVP; expose more options only if needed

**Config defaults**:
- Exchange type: `topic` (supports routing keys)
- Queue durable: `true`
- Message persistent: `true` (delivery mode 2)

### Decision 4: Metadata mapping strategy

**Choice**: Map metadata to Watermill `message.Metadata` (a `map[string]string`).

**Rationale**:
- Watermill metadata is `map[string]string`, our interface uses `map[string]string`
- Direct mapping without transformation needed
- Preserves all metadata through the broker

## Risks / Trade-offs

**[Risk] Connection loss during publish → [Mitigation]** Watermill AMQP handles reconnection automatically; return error to caller for retry logic

**[Risk] Message ordering not guaranteed → [Mitigation]** Use `vehicle_id` as routing key (per PRD); RabbitMQ topic exchanges route by routing key, partial ordering per vehicle

**[Risk] Increased complexity with wrapper layer → [Mitigation]** Wrapper is thin (~50 lines each); justified by interface stability and future Kafka migration

**[Risk] RabbitMQ dependency for local dev → [Mitigation]** Keep `InMemoryPublisher`/`InMemorySubscriber` for local dev; use env var or config to select broker
