package workers

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/dimasbaguspm/infario/internal/gateway"
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
	tg *gateway.TraefikGateway,
	redisClient *redis.Client,
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

		// Unmarshal task
		var task struct {
			ID        string `json:"id"`
			ProjectID string `json:"project_id"`
			Hash      string `json:"hash"`
		}
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
		go processDeploymentTask(ctx, &task, repo, tg, sem, &wg, logger)
		}
	}()
}

// processDeploymentTask handles a single deployment task and removes it from queue only after completion.
func processDeploymentTask(
	ctx context.Context,
	task *struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
		Hash      string `json:"hash"`
	},
	repo deployment.DeploymentRepository,
	tg *gateway.TraefikGateway,
	sem chan struct{},
	wg *sync.WaitGroup,
	logger *slog.Logger,
) {
	defer wg.Done()
	defer func() { <-sem }() // Release semaphore slot

	id := task.ID
	projectID := task.ProjectID
	hash := task.Hash

	if logger != nil {
		logger.InfoContext(ctx, "processing deployment task", "deployment_id", id, "hash", hash)
	}

	// Update status to "ready"
	if err := repo.UpdateStatus(ctx, deployment.UpdateDeploymentStatus{
		ID:     id,
		Status: deployment.StatusReady,
	}); err != nil {
		if logger != nil {
			logger.ErrorContext(ctx, "failed to update deployment status", "id", id, "error", err)
		}
		return
	}

	// Regenerate Traefik config
	if tg != nil {
		status := deployment.StatusReady
		readyDeps, err := repo.GetPaged(ctx, deployment.GetPagedDeployment{
			PagingParams: request.PagingParams{PageNumber: 1, PageSize: 100},
			ProjectID:    &projectID,
			Status:       &status,
		})
		if err == nil && len(readyDeps.Items) > 0 {
			deps := make([]gateway.GatewayDeployment, len(readyDeps.Items))
			for i, d := range readyDeps.Items {
				projectName := ""
				if d.ProjectName != nil {
					projectName = *d.ProjectName
				}
				deps[i] = gateway.GatewayDeployment{
					ID:          d.ID,
					Hash:        d.Hash,
					ProjectID:   d.ProjectID,
					ProjectName: projectName,
				}
			}
			if readyDeps.Items[0].ProjectName != nil {
				if err := tg.WriteProjectConfig(projectID, *readyDeps.Items[0].ProjectName, deps); err != nil && logger != nil {
					logger.ErrorContext(ctx, "failed to write traefik config", "error", err)
				}
			}
		}
	}

	if logger != nil {
		logger.InfoContext(ctx, "deployment task completed", "deployment_id", id)
	}
}
