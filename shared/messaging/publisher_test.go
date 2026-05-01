package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

func TestWatermillPublisher_Publish(t *testing.T) {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NewStdLogger(false, false))
	pub := NewWatermillPublisher(pubSub)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pub.Publish(ctx, "test-topic", []byte("test-payload"), map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
}

func TestWatermillPublisher_Publish_ContextCancelled(t *testing.T) {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NewStdLogger(false, false))
	pub := NewWatermillPublisher(pubSub)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := pub.Publish(ctx, "test-topic", []byte("test-payload"), nil)
	if err == nil {
		t.Fatal("Expected context cancelled error, got nil")
	}
}

func TestWatermillPublisher_Close(t *testing.T) {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NewStdLogger(false, false))
	pub := NewWatermillPublisher(pubSub)

	err := pub.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}
