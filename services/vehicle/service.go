package vehicle

import "context"

type Vehicle struct {
	ID        string
	Plate     string
	Model     string
	CreatedAt string
}

type Repository interface {
	Create(ctx context.Context, vehicle *Vehicle) error
	GetByID(ctx context.Context, id string) (*Vehicle, error)
	GetAll(ctx context.Context) ([]*Vehicle, error)
}

type Service interface {
	RegisterVehicle(ctx context.Context, plate, model string) (*Vehicle, error)
	GetVehicle(ctx context.Context, id string) (*Vehicle, error)
	ListVehicles(ctx context.Context) ([]*Vehicle, error)
}