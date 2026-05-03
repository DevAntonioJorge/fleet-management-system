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

type service struct {
	repo Repository
}

func NewService(repo Repository) (Service, error) {
	return &service{repo: repo}, nil
}

func (s *service) RegisterVehicle(ctx context.Context, plate, model string) (*Vehicle, error) {
	v := &Vehicle{
		Plate: plate,
		Model: model,
	}

	if err := s.repo.Create(ctx, v); err != nil {
		return nil, err
	}

	return v, nil
}

func (s *service) GetVehicle(ctx context.Context, id string) (*Vehicle, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListVehicles(ctx context.Context) ([]*Vehicle, error) {
	return s.repo.GetAll(ctx)
}