package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

func TestWatermillSubscriber_Subscribe(t *testing.T) {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NewStdLogger(false, false))
	sub := NewWatermillSubscriber(pubSub)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var receivedPayload []byte
	handler := func(payload []byte, metadata map[string]string) error {
		receivedPayload = payload
		return nil
	}

	err := sub.Subscribe(ctx, "test-topic", handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Publish a test message using the same gochannel
	msg := message.NewMessage(watermill.NewUUID(), []byte("test-payload"))
	err = pubSub.Publish("test-topic", msg)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for message to be processed
	time.Sleep(100 * time.Millisecond)

	if string(receivedPayload) != "test-payload" {
		t.Fatalf("Expected payload 'test-payload', got '%s'", receivedPayload)
	}
}

func TestWatermillSubscriber_Subscribe_ErrorHandler(t *testing.T) {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NewStdLogger(false, false))
	sub := NewWatermillSubscriber(pubSub)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	handler := func(payload []byte, metadata map[string]string) error {
		return NewError("handler error")
	}

	err := sub.Subscribe(ctx, "test-topic", handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Publish a message that will cause error
	msg := message.NewMessage(watermill.NewUUID(), []byte("test-payload"))
	// Note: gochannel doesn't support Nack, so this test is limited
	_ = pubSub.Publish("test-topic", msg)
}

func NewError(msg string) error {
	return &customError{msg: msg}
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
