package alert

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

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

type service struct {
	repo Repository
}

func NewService(repo Repository) (Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	return &service{repo: repo}, nil
}

func (s *service) CreateAlert(ctx context.Context, vehicleID, alertType, description string) (*Alert, error) {
	id := uuid.New().String()
	ts := time.Now().UTC().Format(time.RFC3339)

	a := &Alert{
		ID:          id,
		VehicleID:   vehicleID,
		Type:        alertType,
		Description: description,
		Timestamp:   ts,
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *service) GetVehicleAlerts(ctx context.Context, vehicleID string) ([]*Alert, error) {
	return s.repo.GetByVehicleID(ctx, vehicleID)
}

func (s *service) GetAllAlerts(ctx context.Context) ([]*Alert, error) {
	return s.repo.GetAll(ctx)
}