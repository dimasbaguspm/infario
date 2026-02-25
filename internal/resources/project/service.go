package project

import (
	"context"
	"fmt"

	"github.com/dimasbaguspm/infario/pkgs/validator"
)

type Service struct {
	repo ProjectRepository
}

func NewService(repo ProjectRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetPagedProjects(ctx context.Context, params GetPagedProject) (*ProjectPaged, error) {
	if err := validator.Validate.Struct(params); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}
	page, err := s.repo.GetPaged(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("Failed to list projects: %w", err)
	}
	return page, nil
}

func (s *Service) GetProjectByID(ctx context.Context, p GetSingleProject) (*Project, error) {
	if err := validator.Validate.Struct(p); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}
	resp, err := s.repo.GetByID(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("Failed to get project by id: %w", err)
	}
	return resp, nil
}

func (s *Service) CreateNewProject(ctx context.Context, p CreateProject) (*Project, error) {
	if err := validator.Validate.Struct(p); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}
	ID, err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("Failed to create project: %w", err)
	}
	return s.GetProjectByID(ctx, GetSingleProject{ID: ID})
}

func (s *Service) UpdateProject(ctx context.Context, p UpdateProject) (*Project, error) {
	if err := validator.Validate.Struct(p); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}
	err := s.repo.Update(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("Failed to update project: %w", err)
	}
	return s.GetProjectByID(ctx, GetSingleProject{ID: p.ID})
}

func (s *Service) DeleteProject(ctx context.Context, p DeleteProject) error {
	if err := validator.Validate.Struct(p); err != nil {
		return fmt.Errorf("Validation failed: %w", err)
	}
	err := s.repo.Delete(ctx, p)
	if err != nil {
		return fmt.Errorf("Failed to delete project: %w", err)
	}
	return nil
}
