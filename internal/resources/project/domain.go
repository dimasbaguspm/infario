package project

import (
	"context"
	"time"

	"github.com/dimasbaguspm/infario/pkgs/request"
	"github.com/dimasbaguspm/infario/pkgs/response"
)

// Project represents a project entity in the system.
// @Description Project entity representing a project with its metadata
// @Name Project
type Project struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// GetSingleProject represents the payload for retrieving a project by ID.
// @Description Payload for fetching a project by its ID
// @Name GetSingleProject
type GetSingleProject struct {
	ID string `json:"id" validate:"required"`
}

// GetPagedProject represents pagination parameters for listing projects.
// @Description Pagination parameters for listing projects
// @Name GetPagedProject
type GetPagedProject struct {
	request.PagingParams
}

// ProjectPaged represents an offset-based paginated response of projects.
// @Description Offset-based paginated project response with metadata
// @Name ProjectPaged
type ProjectPaged response.Collection[*Project]

// CreateProject represents the payload for new projects.
// @Description Project creation DTO
// @Name CreateProject
type CreateProject struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}

// UpdateProject represents the payload for updating existing projects.
// @Description Project update DTO
// @Name UpdateProject
type UpdateProject struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
}

// DeleteProject represents the payload for deleting a project.
// @Description Project deletion DTO
// @Name DeleteProject
type DeleteProject struct {
	ID string `json:"id" validate:"required"`
}

type ProjectRepository interface {
	GetPaged(ctx context.Context, params GetPagedProject) (*ProjectPaged, error)
	GetByID(ctx context.Context, p GetSingleProject) (*Project, error)
	Create(ctx context.Context, p CreateProject) (string, error)
	Update(ctx context.Context, p UpdateProject) error
	Delete(ctx context.Context, p DeleteProject) error
}

type ProjectService interface {
	GetPagedProjects(ctx context.Context, params GetPagedProject) (*ProjectPaged, error)
	GetProjectByID(ctx context.Context, p GetSingleProject) (*Project, error)
	CreateNewProject(ctx context.Context, p CreateProject) (*Project, error)
	UpdateProject(ctx context.Context, p UpdateProject) (*Project, error)
	DeleteProject(ctx context.Context, p DeleteProject) error
}
