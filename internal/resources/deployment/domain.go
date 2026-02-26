package deployment

import (
	"context"
	"time"

	"github.com/dimasbaguspm/infario/pkgs/request"
	"github.com/dimasbaguspm/infario/pkgs/response"
)

const (
	StatusPending = "pending"
	StatusReady   = "ready"
	StatusError   = "error"
	StatusExpired = "expired"

	// Redis queue key for deployment tasks
	QueueKey = "deployments"
)

// Deployment represents a single immutable build artifact.
// Storage assets are managed at: /storage/{projectId}/{hash}/*
// @Description Deployment entity representing a built artifact with content-addressable identifier
// @Name Deployment
type Deployment struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	Hash          string     `json:"hash"` // The content-addressable identifier
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiredAt     *time.Time `json:"expired_at,omitempty"` // Nullable: some builds may never expire
	ProjectName   *string    `json:"project_name,omitempty"`
}

// GetSingleDeployment represents the payload for retrieving a deployment by ID.
// @Description Payload for fetching a deployment by its ID
// @Name GetSingleDeployment
type GetSingleDeployment struct {
	ID string `json:"id" validate:"required,uuid4"`
}

// GetPagedDeployment represents pagination parameters for listing deployments.
// @Description Pagination parameters for listing deployments with optional filters
// @Name GetPagedDeployment
type GetPagedDeployment struct {
	request.PagingParams
	ProjectID *string `json:"project_id" validate:"omitempty,uuid4"`
	Status    *string `json:"status" validate:"omitempty,oneof=pending ready error expired"`
}

// DeploymentPaged represents a paginated response of deployments.
// @Description Paginated deployment response with metadata
// @Name DeploymentPaged
type DeploymentPaged response.Collection[*Deployment]

// UploadDeployment represents the payload for uploading a deployment artifact.
// @Description Deployment upload DTO (zip or tar.gz with default 30-day TTL)
// @Name UploadDeployment
type UploadDeployment struct {
	ProjectID string `json:"project_id" validate:"required,uuid4"`
	Hash      string `json:"hash" validate:"required"` // Content-addressable identifier
	request.FileUpload
}

// UpdateDeploymentStatus represents the payload for updating deployment status.
// @Description Deployment status update DTO
// @Name UpdateDeploymentStatus
type UpdateDeploymentStatus struct {
	ID     string `json:"id" validate:"required"`
	Status string `json:"status" validate:"required,oneof=pending ready error expired"`
}

type DeploymentRepository interface {
	GetByID(ctx context.Context, d GetSingleDeployment) (*Deployment, error)
	GetPaged(ctx context.Context, params GetPagedDeployment) (*DeploymentPaged, error)
	Upload(ctx context.Context, d UploadDeployment) (string, error)
	UpdateStatus(ctx context.Context, d UpdateDeploymentStatus) error
	GetExpired(ctx context.Context) ([]Deployment, error)
}

type DeploymentService interface {
	GetDeploymentByID(ctx context.Context, d GetSingleDeployment) (*Deployment, error)
	GetPagedDeployments(ctx context.Context, params GetPagedDeployment) (*DeploymentPaged, error)
	Upload(ctx context.Context, d UploadDeployment) (*Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, d UpdateDeploymentStatus) (*Deployment, error)
}
