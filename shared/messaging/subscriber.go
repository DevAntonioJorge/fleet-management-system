package messaging

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type MessageHandler func(msg *message.Message) error

type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Close() error
}

type WatermillSubscriber struct {
	sub *message.Subscriber
}

func NewSubscriber(sub *message.Subscriber) *WatermillSubscriber {
	return &WatermillSubscriber{sub: sub}
}

func (s *WatermillSubscriber) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	return s.sub.Subscribe(ctx, topic, handler)
}

func (s *WatermillSubscriber) Close() error {
	return s.sub.Close()
}

type SubscriberFactory interface {
	CreateSubscriber(config interface{}) (Subscriber, error)
}

type RabbitMQSubscriberFactory struct{}

func NewRabbitMQSubscriberFactory() *RabbitMQSubscriberFactory {
	return &RabbitMQSubscriberFactory{}
}

func (f *RabbitMQSubscriberFactory) CreateSubscriber(url, queue string) (Subscriber, error) {
	sub, err := message.NewSubscriber(
		message.RabbitMQSubscriberConfig{
			QueueName:      queue,
			URL:            url,
			RoutingKey:    "#",
			DeliveryType:  message.DeliverAtLeastOnce,
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ subscriber: %w", err)
	}

	return NewSubscriber(sub), nil
}

func CreateTelemetryHandler(fn func(ctx context.Context, tm TelemetryMessage) error) MessageHandler {
	return func(msg *message.Message) error {
		var tm TelemetryMessage
		if err := tm.UnmarshalJSON(msg.Payload()); err != nil {
			return fmt.Errorf("failed to unmarshal telemetry: %w", err)
		}

		ctx := context.Background()
		return fn(ctx, tm)
	}
}

func CreateAlertHandler(fn func(ctx context.Context, am AlertMessage) error) MessageHandler {
	return func(msg *message.Message) error {
		var am AlertMessage
		if err := am.UnmarshalJSON(msg.Payload()); err != nil {
			return fmt.Errorf("failed to unmarshal alert: %w", err)
		}

		ctx := context.Background()
		return fn(ctx, am)
	}
}