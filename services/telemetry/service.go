package telemetry

import "context"

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
	GetVehicleTelemetry(ctx context.Context, vehicleID, from, to string) ([]*TelemetryEvent, error)
	GetLastLocation(ctx context.Context, vehicleID string) (*TelemetryEvent, error)
}