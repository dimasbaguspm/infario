package deployment

import (
	"context"
	"fmt"

	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/dimasbaguspm/infario/pkgs/redis"
	"github.com/dimasbaguspm/infario/pkgs/validator"
	goredis "github.com/redis/go-redis/v9"
)

type Service struct {
	repo       DeploymentRepository
	redis      *goredis.Client
	fileEngine *engine.FileEngine
}

func NewService(repo DeploymentRepository, redisClient *goredis.Client, fileEngine *engine.FileEngine) *Service {
	return &Service{
		repo:       repo,
		redis:      redisClient,
		fileEngine: fileEngine,
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

	ID, err := s.repo.Upload(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("Failed to create deployment record: %w", err)
	}

	// Extract uploaded archive asynchronously
	go func() {
		destPath := "deployments/" + d.ProjectID + "/" + ID

		// Open and extract the file
		file, err := d.File.Open()
		if err != nil {
			fmt.Printf("failed to open upload file for extraction: %v\n", err)
			return
		}
		defer file.Close()

		err = s.fileEngine.Extract(context.Background(), destPath, file, d.File.Filename)
		if err != nil {
			fmt.Printf("failed to extract deployment file: %v\n", err)
			return
		}

		// After extraction, emit task to Redis for validation and Traefik config
		task := DeploymentTask{
			Deployment: &Deployment{
				ID: ID,
			},
			OriginalName: d.File.Filename,
		}
		if err := redis.Emit(context.Background(), s.redis, QueueKey, task); err != nil {
			fmt.Printf("failed to emit deployment task: %v\n", err)
		}
	}()

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
