//go:build integration

package messaging

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

func TestIntegration_PublishSubscribe(t *testing.T) {
	// This test requires a running RabbitMQ instance
	// Run with: go test -tags=integration ./shared/messaging/...
	url := "amqp://guest:guest@localhost:5672/"
	exchange := "test-exchange"
	queue := "test-queue"
	topic := "test-topic"

	// Create publisher
	pubFactory := NewRabbitMQPublisherFactory()
	publisher, err := pubFactory.CreatePublisher(url, exchange)
	if err != nil {
		t.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// Create subscriber
	subFactory := NewRabbitMQSubscriberFactory()
	subscriber, err := subFactory.CreateSubscriber(url, queue)
	if err != nil {
		t.Fatalf("Failed to create subscriber: %v", err)
	}
	defer subscriber.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Subscribe to topic
	var receivedPayload []byte
	err = subscriber.Subscribe(ctx, topic, func(payload []byte, metadata map[string]string) error {
		receivedPayload = payload
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Publish message
	testPayload := []byte("integration-test-payload")
	err = publisher.Publish(ctx, topic, testPayload, nil)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for message
	time.Sleep(1 * time.Second)

	if string(receivedPayload) != string(testPayload) {
		t.Fatalf("Expected payload '%s', got '%s'", testPayload, receivedPayload)
	}
}
