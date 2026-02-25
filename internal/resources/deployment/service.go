package deployment

import (
	"context"
	"fmt"

	"github.com/dimasbaguspm/infario/pkgs/validator"
)

type Service struct {
	repo DeploymentRepository
}

func NewService(repo DeploymentRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDeploymentByID(ctx context.Context, d GetSingleDeployment) (*Deployment, error) {
	if err := validator.Validate.Struct(d); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}
	resp, err := s.repo.GetByID(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("Failed to get deployment by id: %w", err)
	}
	return resp, nil
}

func (s *Service) GetPagedDeployments(ctx context.Context, params GetPagedDeployment) (*DeploymentPaged, error) {
	if err := validator.Validate.Struct(params); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}
	page, err := s.repo.GetPaged(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("Failed to list deployments: %w", err)
	}
	return page, nil
}

func (s *Service) CreateDeployment(ctx context.Context, d CreateDeployment) (*Deployment, error) {
	if err := validator.Validate.Struct(d); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}
	ID, err := s.repo.Create(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("Failed to create deployment: %w", err)
	}
	return s.GetDeploymentByID(ctx, GetSingleDeployment{ID: ID})
}

func (s *Service) UpdateDeploymentStatus(ctx context.Context, d UpdateDeploymentStatus) (*Deployment, error) {
	if err := validator.Validate.Struct(d); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}
	err := s.repo.UpdateStatus(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("Failed to update deployment status: %w", err)
	}
	return s.GetDeploymentByID(ctx, GetSingleDeployment{ID: d.ID})
}

func (s *Service) DeleteDeployment(ctx context.Context, d DeleteDeployment) error {
	if err := validator.Validate.Struct(d); err != nil {
		return fmt.Errorf("Validation failed: %w", err)
	}
	err := s.repo.Delete(ctx, d)
	if err != nil {
		return fmt.Errorf("Failed to delete deployment: %w", err)
	}
	return nil
}
