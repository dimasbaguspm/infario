package project

import (
	"context"
	"time"
)

type Project struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	GitURL        string     `json:"git_url"`
	GitProvider   string     `json:"git_provider"`
	PrimaryBranch string     `json:"primary_branch"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletdAt      *time.Time `json:"deleted_at,omitempty"`
}

type GetSingleProject struct {
	ID string `json:"id" validate:"required"`
}

type CreateProject struct {
	Name          string `json:"name" validate:"required"`
	GitURL        string `json:"git_url" validate:"required"`
	GitProvider   string `json:"git_provider,omitempty"`
	PrimaryBranch string `json:"primary_branch,omitempty"`
}

type UpdateProject struct {
	ID            string `json:"id" validate:"required"`
	Name          string `json:"name,omitempty"`
	GitURL        string `json:"git_url,omitempty"`
	GitProvider   string `json:"git_provider,omitempty"`
	PrimaryBranch string `json:"primary_branch,omitempty"`
}

type DeleteProject struct {
	ID string `json:"id" validate:"required"`
}

type ProjectRepository interface {
	Create(ctx context.Context, p CreateProject) (string, error)
	GetByID(ctx context.Context, p GetSingleProject) (*Project, error)
	Update(ctx context.Context, p UpdateProject) error
	Delete(ctx context.Context, p DeleteProject) error
}

type ProjectService interface {
	GetProjectByID(ctx context.Context, p GetSingleProject) (*Project, error)
	CreateNewProject(ctx context.Context, p CreateProject) (*Project, error)
	UpdateProject(ctx context.Context, p UpdateProject) (*Project, error)
	DeleteProject(ctx context.Context, p DeleteProject) error
}
