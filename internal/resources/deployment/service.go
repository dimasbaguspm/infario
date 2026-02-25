package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/dimasbaguspm/infario/internal/platform/worker"
	"github.com/dimasbaguspm/infario/pkgs/validator"
)

type Service struct {
	repo       DeploymentRepository
	workerPool *worker.DeploymentWorkerPool
}

func NewService(repo DeploymentRepository, workerPool *worker.DeploymentWorkerPool) *Service {
	return &Service{
		repo:       repo,
		workerPool: workerPool,
	}
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

func (s *Service) Upload(ctx context.Context, d UploadDeployment) (*Deployment, error) {
	if err := validator.Validate.Struct(d); err != nil {
		return nil, fmt.Errorf("Validation failed: %w", err)
	}

	// 1. Open the uploaded file
	file, err := d.File.Open()
	if err != nil {
		return nil, fmt.Errorf("Failed to open uploaded file: %w", err)
	}

	// 2. Create deployment record in database with "pending" status
	ID, err := s.repo.Upload(ctx, d)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("Failed to create deployment record: %w", err)
	}

	// 3. Enqueue async worker task to extract files and update status
	task := worker.DeploymentTask{
		ID:         ID,
		ProjectID:  d.ProjectID,
		Hash:       d.Hash,
		FileReader: file,
		OnComplete: func(status string, taskErr error) {
			// Update deployment status asynchronously with timeout
			updateCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			updateErr := s.repo.UpdateStatus(updateCtx, UpdateDeploymentStatus{
				ID:     ID,
				Status: status,
			})
			if updateErr != nil {
				// Log but don't fail - deployment record exists with status info
				fmt.Printf("failed to update deployment status: %v\n", updateErr)
			}
		},
	}
	s.workerPool.Enqueue(task)

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
