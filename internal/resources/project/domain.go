package project

import (
	"context"
	"time"
)

// Project represents a project entity in the system.
// @Description Project entity representing a code repository and its metadata
// @Name Project
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

// GetSingleProject represents the payload for retrieving a project by ID.
// @Description Payload for fetching a project by its ID
// @Name GetSingleProject
type GetSingleProject struct {
	ID string `json:"id" validate:"required"`
}

// CreateProject represents the payload for new projects.
// @Description Project creation DTO
// @Name CreateProject
type CreateProject struct {
	Name          string `json:"name" validate:"required,min=3,max=100"`
	GitURL        string `json:"git_url" validate:"required,url"`
	GitProvider   string `json:"git_provider,omitempty" validate:"required,oneof=github gitlab bitbucket"`
	PrimaryBranch string `json:"primary_branch,omitempty" validate:"omitempty"`
}

// UpdateProject represents the payload for updating existing projects.
// @Description Project update DTO
// @Name UpdateProject
type UpdateProject struct {
	ID            string `json:"id" validate:"required"`
	Name          string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	GitURL        string `json:"git_url,omitempty" validate:"omitempty,url"`
	GitProvider   string `json:"git_provider,omitempty" validate:"omitempty,oneof=github gitlab bitbucket"`
	PrimaryBranch string `json:"primary_branch,omitempty" validate:"omitempty"`
}

// DeleteProject represents the payload for deleting a project.
// @Description Project deletion DTO
// @Name DeleteProject
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
