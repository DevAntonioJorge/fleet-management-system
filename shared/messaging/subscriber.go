package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/fms/fms/shared/logger"
)

type MessageHandler func(payload []byte, metadata map[string]string) error

type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Close() error
}

type InMemorySubscriber struct {
	publisher *InMemoryPublisher
}

func NewInMemorySubscriber(pub *InMemoryPublisher) *InMemorySubscriber {
	return &InMemorySubscriber{publisher: pub}
}

func (s *InMemorySubscriber) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	ch := s.publisher.Subscribe(topic)

	go func() {
		for {
			select {
			case payload, ok := <-ch:
				if !ok {
					return
				}
				handler(payload, nil)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (s *InMemorySubscriber) Close() error {
	return nil
}

// WatermillSubscriber wraps message.Subscriber to implement messaging.Subscriber
type WatermillSubscriber struct {
	sub message.Subscriber
}

func (s *WatermillSubscriber) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	messages, err := s.sub.Subscribe(ctx, topic)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	go func() {
		for {
			select {
			case msg, ok := <-messages:
				if !ok {
					return
				}
				payload := msg.Payload
				metadata := make(map[string]string)
				for k, v := range msg.Metadata {
					metadata[k] = v
				}
				if err := handler(payload, metadata); err != nil {
					msg.Nack()
				} else {
					msg.Ack()
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (s *WatermillSubscriber) Close() error {
	return s.sub.Close()
}

// NewWatermillSubscriber creates a WatermillSubscriber from a message.Subscriber
func NewWatermillSubscriber(sub message.Subscriber) *WatermillSubscriber {
	return &WatermillSubscriber{sub: sub}
}

type SubscriberFactory interface {
	CreateSubscriber(url, queue string) (Subscriber, error)
}

type RabbitMQSubscriberFactory struct{}

func NewRabbitMQSubscriberFactory() *RabbitMQSubscriberFactory {
	return &RabbitMQSubscriberFactory{}
}

func (f *RabbitMQSubscriberFactory) CreateSubscriber(url, queue string) (Subscriber, error) {
	config := amqp.Config{
		Connection: amqp.ConnectionConfig{
			AmqpURI: url,
		},
		Queue: amqp.QueueConfig{
			GenerateName: amqp.GenerateQueueNameConstant(queue),
			Durable:      true,
		},
		QueueBind: amqp.QueueBindConfig{
			GenerateRoutingKey: func(topic string) string {
				return topic
			},
		},
		Exchange: amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return "fms"
			},
			Type:    "topic",
			Durable: true,
		},
	}

	logger := NewWatermillLoggerAdapter(logger.Default())
	sub, err := amqp.NewSubscriber(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create AMQP subscriber: %w", err)
	}

	return NewWatermillSubscriber(sub), nil
}

func CreateTelemetryHandler(fn func(ctx context.Context, tm TelemetryMessage) error) MessageHandler {
	return func(payload []byte, metadata map[string]string) error {
		var tm TelemetryMessage
		if err := json.Unmarshal(payload, &tm); err != nil {
			return fmt.Errorf("failed to unmarshal telemetry: %w", err)
		}

		ctx := context.Background()
		return fn(ctx, tm)
	}
}

func CreateAlertHandler(fn func(ctx context.Context, am AlertMessage) error) MessageHandler {
	return func(payload []byte, metadata map[string]string) error {
		var am AlertMessage
		if err := json.Unmarshal(payload, &am); err != nil {
			return fmt.Errorf("failed to unmarshal alert: %w", err)
		}

		ctx := context.Background()
		return fn(ctx, am)
	}
}