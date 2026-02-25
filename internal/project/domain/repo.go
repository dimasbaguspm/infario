package domain

import "context"

type ProjectRepository interface {
	Create(ctx context.Context, p CreateDeployment) error
	GetByID(ctx context.Context, p GetSingleDeployment) (*Project, error)
	Update(ctx context.Context, p UpdateDeployment) error
	Delete(ctx context.Context, p DeleteDeployment) error
}

type DeploymentRepository interface {
	Create(ctx context.Context, p CreateDeployment) error
	GetByID(ctx context.Context, p GetSingleDeployment) (*Deployment, error)
	Update(ctx context.Context, p UpdateDeployment) error
	Delete(ctx context.Context, p DeleteDeployment) error
}
