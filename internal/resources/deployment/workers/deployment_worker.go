package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dimasbaguspm/infario/internal/gateway"
	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/dimasbaguspm/infario/internal/resources/deployment"
	"github.com/dimasbaguspm/infario/pkgs/request"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	MaxConcurrentWorkers = 50
)

// StartDeploymentConsumer drains deployment tasks from Redis and processes them with controlled concurrency.
// Uses a semaphore pattern to limit concurrent workers and ensures tasks are removed only after completion.
func StartDeploymentConsumer(
	ctx context.Context,
	db *pgxpool.Pool,
	tg *gateway.NginxGateway,
	redisClient *redis.Client,
	fileEngine *engine.FileEngine,
	logger *slog.Logger,
) {
	go func() {
		repo := deployment.NewPostgresRepository(db)

		if logger != nil {
			logger.InfoContext(ctx, "deployment consumer started", "max_workers", MaxConcurrentWorkers)
		}

		// Semaphore to limit concurrent workers
		sem := make(chan struct{}, MaxConcurrentWorkers)
		var wg sync.WaitGroup

		for {
			select {
			case <-ctx.Done():
				if logger != nil {
					logger.InfoContext(ctx, "deployment consumer shutting down, waiting for pending jobs")
				}
				// Wait for all pending jobs to complete
				wg.Wait()
				return
			default:
			}

			result, err := redisClient.BLPop(ctx, 1*time.Second, deployment.QueueKey).Result()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				if logger != nil {
					logger.ErrorContext(ctx, "failed to pop from queue", "error", err)
				}
				continue
			}

			if len(result) < 2 {
				continue
			}

			taskData := result[1]

			// Unmarshal task into DeploymentTask (embeds Deployment)
			var task deployment.DeploymentTask
			err = json.Unmarshal([]byte(taskData), &task)
			if err != nil {
				if logger != nil {
					logger.ErrorContext(ctx, "failed to unmarshal task", "error", err)
				}
				continue
			}

			// Acquire semaphore slot (blocking if all workers are busy)
			sem <- struct{}{}
			wg.Add(1)

			// Process task in goroutine
			go processDeploymentTask(ctx, &task, repo, tg, fileEngine, sem, &wg, logger)
		}
	}()
}

// processDeploymentTask validates extracted deployment files and regenerates Traefik config.
func processDeploymentTask(
	ctx context.Context,
	task *deployment.DeploymentTask,
	repo deployment.DeploymentRepository,
	tg *gateway.NginxGateway,
	fileEngine *engine.FileEngine,
	sem chan struct{},
	wg *sync.WaitGroup,
	logger *slog.Logger,
) {
	defer wg.Done()
	defer func() { <-sem }() // Release semaphore slot

	// Fetch the full Deployment record from DB
	dep, err := repo.GetByID(ctx, deployment.GetSingleDeployment{ID: task.Deployment.ID})
	if err != nil {
		if logger != nil {
			logger.ErrorContext(ctx, "failed to fetch deployment", "id", task.Deployment.ID, "error", err)
		}
		return
	}

	if logger != nil {
		logger.InfoContext(ctx, "processing deployment task", "deployment_id", dep.ID, "entry_path", dep.EntryPath)
	}

	// Define deployment directory where files were extracted
	deploymentDir := "deployments/" + dep.ProjectID + "/" + dep.ID

	// Validate that entry_path was provided at upload time
	if dep.EntryPath == "" {
		if logger != nil {
			logger.ErrorContext(ctx, "entry_path is required but was not provided at upload", "id", dep.ID)
		}
		if err := repo.UpdateStatus(ctx, deployment.UpdateDeploymentStatus{
			ID:     dep.ID,
			Status: deployment.StatusError,
		}); err != nil && logger != nil {
			logger.ErrorContext(ctx, "failed to update deployment status to error", "id", dep.ID, "error", err)
		}
		return
	}

	// Verify that entry_path exists within the extracted files
	// entry_path can be a file (e.g., "/index.html") or directory (e.g., "/app")
	entryPathFull := deploymentDir + dep.EntryPath
	fmt.Printf("Validating entry_path for deployment %s: checking existence of %s\n", dep.ID, entryPathFull)
	if !fileEngine.Exists(ctx, entryPathFull) {
		if logger != nil {
			logger.ErrorContext(ctx, "entry_path not found in extracted archive", "id", dep.ID, "entry_path", dep.EntryPath, "full_path", entryPathFull)
		}
		if err := repo.UpdateStatus(ctx, deployment.UpdateDeploymentStatus{
			ID:     dep.ID,
			Status: deployment.StatusError,
		}); err != nil && logger != nil {
			logger.ErrorContext(ctx, "failed to update deployment status to error", "id", dep.ID, "error", err)
		}
		return
	}

	// Update status to "ready"
	if err := repo.UpdateStatus(ctx, deployment.UpdateDeploymentStatus{
		ID:     dep.ID,
		Status: deployment.StatusReady,
	}); err != nil {
		if logger != nil {
			logger.ErrorContext(ctx, "failed to update deployment status", "id", dep.ID, "error", err)
		}
		return
	}

	// Regenerate Traefik config
	if tg != nil {
		status := deployment.StatusReady
		readyDeps, err := repo.GetPaged(ctx, deployment.GetPagedDeployment{
			PagingParams: request.PagingParams{PageNumber: 1, PageSize: 100},
			ProjectID:    &dep.ProjectID,
			Status:       &status,
		})
		if err == nil && len(readyDeps.Items) > 0 {
			deps := make([]gateway.GatewayDeployment, len(readyDeps.Items))
			for i, d := range readyDeps.Items {
				projectName := ""
				if d.ProjectName != nil {
					projectName = *d.ProjectName
				}
				entryPath := d.EntryPath
				deps[i] = gateway.GatewayDeployment{
					ID:          d.ID,
					Hash:        d.Hash,
					ProjectID:   d.ProjectID,
					ProjectName: projectName,
					EntryPath:   &entryPath,
				}
			}
			if readyDeps.Items[0].ProjectName != nil {
				if err := tg.WriteProjectConfig(dep.ProjectID, *readyDeps.Items[0].ProjectName, deps); err != nil && logger != nil {
					logger.ErrorContext(ctx, "failed to write traefik config", "error", err)
				}
			}
		}
	}

	if logger != nil {
		logger.InfoContext(ctx, "deployment task completed", "deployment_id", dep.ID)
	}
}
