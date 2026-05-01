package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/fms/fms/shared/logger"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, payload []byte, metadata map[string]string) error
	Close() error
}

type InMemoryPublisher struct {
	mu    sync.RWMutex
	subs  map[string][]chan []byte
	closed bool
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		subs: make(map[string][]chan []byte),
	}
}

func (p *InMemoryPublisher) Publish(ctx context.Context, topic string, payload []byte, metadata map[string]string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("publisher is closed")
	}

	chans, ok := p.subs[topic]
	if !ok {
		return nil
	}

	for _, ch := range chans {
		select {
		case ch <- payload:
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

func (p *InMemoryPublisher) Subscribe(topic string) chan []byte {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan []byte, 100)
	p.subs[topic] = append(p.subs[topic], ch)
	return ch
}

func (p *InMemoryPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	for _, chans := range p.subs {
		for _, ch := range chans {
			close(ch)
		}
	}
	p.subs = nil
	return nil
}

// WatermillPublisher wraps message.Publisher to implement messaging.Publisher
type WatermillPublisher struct {
	pub message.Publisher
}

func (p *WatermillPublisher) Publish(ctx context.Context, topic string, payload []byte, metadata map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	return p.pub.Publish(topic, msg)
}

func (p *WatermillPublisher) Close() error {
	return p.pub.Close()
}

// NewWatermillPublisher creates a WatermillPublisher from a message.Publisher
func NewWatermillPublisher(pub message.Publisher) *WatermillPublisher {
	return &WatermillPublisher{pub: pub}
}

type PublisherFactory interface {
	CreatePublisher(url, exchange string) (Publisher, error)
}

type RabbitMQPublisherFactory struct{}

func NewRabbitMQPublisherFactory() *RabbitMQPublisherFactory {
	return &RabbitMQPublisherFactory{}
}

func (f *RabbitMQPublisherFactory) CreatePublisher(url, exchange string) (Publisher, error) {
	config := amqp.Config{
		Connection: amqp.ConnectionConfig{
			AmqpURI: url,
		},
		Exchange: amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return exchange
			},
			Type:    "topic",
			Durable: true,
		},
	}

	logger := NewWatermillLoggerAdapter(logger.Default())
	pub, err := amqp.NewPublisher(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create AMQP publisher: %w", err)
	}

	return NewWatermillPublisher(pub), nil
}

func PublishTelemetry(ctx context.Context, pub Publisher, tm TelemetryMessage) error {
	payload, err := json.Marshal(tm)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry: %w", err)
	}

	return pub.Publish(ctx, "raw", payload, map[string]string{
		"vehicle_id": tm.VehicleID,
		"type":       string(MessageTypeTelemetry),
	})
}

func PublishAlert(ctx context.Context, pub Publisher, am AlertMessage) error {
	payload, err := json.Marshal(am)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	return pub.Publish(ctx, "alerts", payload, map[string]string{
		"vehicle_id": am.VehicleID,
		"type":       string(MessageTypeAlert),
	})
}