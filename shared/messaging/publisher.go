package messaging

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, msg *message.Message) error
	Close() error
}

type WatermillPublisher struct {
	pub *message.Publisher
}

func NewPublisher(pub *message.Publisher) *WatermillPublisher {
	return &WatermillPublisher{pub: pub}
}

func (p *WatermillPublisher) Publish(ctx context.Context, topic string, msg *message.Message) error {
	return p.pub.Publish(topic, msg)
}

func (p *WatermillPublisher) Close() error {
	return p.pub.Close()
}

type PublisherFactory interface {
	CreatePublisher(config interface{}) (Publisher, error)
}

type RabbitMQPublisherFactory struct{}

func NewRabbitMQPublisherFactory() *RabbitMQPublisherFactory {
	return &RabbitMQPublisherFactory{}
}

func (f *RabbitMQPublisherFactory) CreatePublisher(url, exchange string) (Publisher, error) {
	pub, err := message.NewPublisher(
		message.RabbitMQPublisherConfig{
			URL:        url,
			Exchange:   exchange,
			RoutingKey: "",
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ publisher: %w", err)
	}

	return NewPublisher(pub), nil
}

func PublishTelemetry(ctx context.Context, pub Publisher, tm TelemetryMessage) error {
	payload, err := tm.Payload()
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set("vehicle_id", tm.VehicleID)
	msg.Metadata.Set("type", string(MessageTypeTelemetry))

	return pub.Publish(ctx, "raw", msg)
}

func PublishAlert(ctx context.Context, pub Publisher, am AlertMessage) error {
	payload, err := am.Payload()
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set("vehicle_id", am.VehicleID)
	msg.Metadata.Set("type", string(MessageTypeAlert))

	return pub.Publish(ctx, "alerts", msg)
}