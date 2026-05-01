package alert

import "context"

type Alert struct {
	ID          string
	VehicleID   string
	Type        string
	Description string
	Timestamp   string
}

type Repository interface {
	Create(ctx context.Context, alert *Alert) error
	GetByVehicleID(ctx context.Context, vehicleID string) ([]*Alert, error)
	GetAll(ctx context.Context) ([]*Alert, error)
}

type Service interface {
	CreateAlert(ctx context.Context, vehicleID, alertType, description string) (*Alert, error)
	GetVehicleAlerts(ctx context.Context, vehicleID string) ([]*Alert, error)
	GetAllAlerts(ctx context.Context) ([]*Alert, error)
}