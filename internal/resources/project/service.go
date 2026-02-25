package project

import (
	"context"
	"fmt"
)

type Service struct {
	repo ProjectRepository
}

func NewService(repo ProjectRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetProjectByID(ctx context.Context, p GetSingleProject) (*Project, error) {
	resp, err := s.repo.GetByID(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("Failed to get project by id: %w", err)
	}
	return resp, nil
}

func (s *Service) CreateNewProject(ctx context.Context, p CreateProject) (*Project, error) {
	ID, err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("Failed to create project: %w", err)
	}
	return s.GetProjectByID(ctx, GetSingleProject{ID: ID})
}

func (s *Service) UpdateProject(ctx context.Context, p UpdateProject) (*Project, error) {
	err := s.repo.Update(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("Failed to update project: %w", err)
	}
	return s.GetProjectByID(ctx, GetSingleProject{ID: p.ID})
}

func (s *Service) DeleteProject(ctx context.Context, p DeleteProject) error {
	err := s.repo.Delete(ctx, p)
	if err != nil {
		return fmt.Errorf("Failed to delete project: %w", err)
	}
	return nil
}
