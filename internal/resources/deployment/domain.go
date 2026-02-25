package deployment

import (
	"context"
	"time"

	"github.com/dimasbaguspm/infario/pkgs/request"
	"github.com/dimasbaguspm/infario/pkgs/response"
)

const (
	StatusQueued   = "queued"
	StatusBuilding = "building"
	StatusReady    = "ready"
	StatusFailed   = "failed"
)

// Deployment represents a deployment entity in the system.
// @Description Deployment entity representing a built and deployed version of a project
// @Name Deployment
type Deployment struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Status        string    `json:"status"`
	CommitHash    string    `json:"commit_hash"`
	CommitMessage string    `json:"commit_message,omitempty"`
	StoragePath   string    `json:"-"` // Hidden from JSON for security
	PublicURL     string    `json:"public_url,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetSingleDeployment represents the payload for retrieving a deployment by ID.
// @Description Payload for fetching a deployment by its ID
// @Name GetSingleDeployment
type GetSingleDeployment struct {
	ID string `json:"id" validate:"required"`
}

// GetPagedDeployment represents pagination parameters for listing deployments.
// @Description Pagination parameters for listing deployments
// @Name GetPagedDeployment
type GetPagedDeployment struct {
	ProjectID string `json:"project_id" validate:"required"`
	request.PagingParams
}

// DeploymentPaged represents a paginated response of deployments.
// @Description Paginated deployment response with metadata
// @Name DeploymentPaged
type DeploymentPaged response.Collection[*Deployment]

// CreateDeployment represents the payload for creating a new deployment.
// @Description Deployment creation DTO
// @Name CreateDeployment
type CreateDeployment struct {
	ProjectID     string `json:"project_id" validate:"required"`
	CommitHash    string `json:"commit_hash" validate:"required,len=40"`
	CommitMessage string `json:"commit_message" validate:"omitempty"`
	StoragePath   string `json:"storage_path" validate:"omitempty"`
}

// UpdateDeploymentStatus represents the payload for updating deployment status.
// @Description Deployment status update DTO
// @Name UpdateDeploymentStatus
type UpdateDeploymentStatus struct {
	ID        string `json:"id" validate:"required"`
	Status    string `json:"status" validate:"required,oneof=queued building ready failed"`
	PublicURL string `json:"public_url,omitempty" validate:"omitempty,url"`
}

// DeleteDeployment represents the payload for deleting a deployment.
// @Description Deployment deletion DTO
// @Name DeleteDeployment
type DeleteDeployment struct {
	ID string `json:"id" validate:"required"`
}

type DeploymentRepository interface {
	GetByID(ctx context.Context, d GetSingleDeployment) (*Deployment, error)
	GetPaged(ctx context.Context, params GetPagedDeployment) (*DeploymentPaged, error)
	Create(ctx context.Context, d CreateDeployment) (string, error)
	UpdateStatus(ctx context.Context, d UpdateDeploymentStatus) error
	Delete(ctx context.Context, d DeleteDeployment) error
}

type DeploymentService interface {
	GetDeploymentByID(ctx context.Context, d GetSingleDeployment) (*Deployment, error)
	GetPagedDeployments(ctx context.Context, params GetPagedDeployment) (*DeploymentPaged, error)
	CreateDeployment(ctx context.Context, d CreateDeployment) (*Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, d UpdateDeploymentStatus) (*Deployment, error)
	DeleteDeployment(ctx context.Context, d DeleteDeployment) error
}
