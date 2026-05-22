package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/fms/fms/shared/messaging"
)

type TelemetryEvent struct {
	ID         string
	VehicleID  string
	Timestamp  string
	Latitude   float64
	Longitude  float64
	Speed      float64
	FuelLevel  float64
}

type Repository interface {
	Create(ctx context.Context, event *TelemetryEvent) error
	GetByVehicleID(ctx context.Context, vehicleID string, from, to string) ([]*TelemetryEvent, error)
	GetLastByVehicleID(ctx context.Context, vehicleID string) (*TelemetryEvent, error)
}

type Service interface {
	Ingest(ctx context.Context, event *TelemetryEvent) error
	Persist(ctx context.Context, event *TelemetryEvent) error
	GetVehicleTelemetry(ctx context.Context, vehicleID, from, to string) ([]*TelemetryEvent, error)
	GetLastLocation(ctx context.Context, vehicleID string) (*TelemetryEvent, error)
}

type service struct {
	repo      Repository
	publisher messaging.Publisher
}

func NewService(repo Repository, pub messaging.Publisher) (Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	if pub == nil {
		return nil, fmt.Errorf("publisher is required")
	}

	return &service{repo: repo, publisher: pub}, nil
}

func (s *service) Ingest(ctx context.Context, event *TelemetryEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	ts, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	// generate a stable event id for idempotency
	eventID := event.ID
	if eventID == "" {
		eventID = uuid.New().String()
	}

	tm := messaging.TelemetryMessage{
		EventID:   eventID,
		VehicleID: event.VehicleID,
		Timestamp: ts,
		Latitude:  event.Latitude,
		Longitude: event.Longitude,
		Speed:     event.Speed,
		FuelLevel: event.FuelLevel,
	}

	return messaging.PublishTelemetry(ctx, s.publisher, tm)
}

func (s *service) Persist(ctx context.Context, event *TelemetryEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	return s.repo.Create(ctx, event)
}

func (s *service) GetVehicleTelemetry(ctx context.Context, vehicleID, from, to string) ([]*TelemetryEvent, error) {
	return s.repo.GetByVehicleID(ctx, vehicleID, from, to)
}

func (s *service) GetLastLocation(ctx context.Context, vehicleID string) (*TelemetryEvent, error) {
	return s.repo.GetLastByVehicleID(ctx, vehicleID)
}