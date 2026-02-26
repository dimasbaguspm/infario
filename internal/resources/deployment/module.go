package deployment

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/dimasbaguspm/infario/internal/platform/scheduler"
	"github.com/dimasbaguspm/infario/internal/platform/worker"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeploymentConfig holds configuration for the deployment module's background maintenance.
type DeploymentConfig struct {
	MaintenanceInterval time.Duration
	CleanupConcurrency  int
	Logger              *slog.Logger
}

// InitHttp registers HTTP routes for deployment endpoints.
// This is called from resources.RegisterRoutes() to set up the HTTP API.
func InitHttp(mux *http.ServeMux, pgx *pgxpool.Pool, workerPool *worker.DeploymentWorkerPool) {
	repo := NewPostgresRepository(pgx)
	service := NewService(repo, workerPool)
	RegisterRoutes(mux, *service)
}

// InitWorker creates a maintenance runner for cleaning up expired deployments.
// This is called from resources.RegisterRoutes() to initialize background tasks.
// Returns a runner that should be started in main.go via go runner.Start(ctx).
func InitWorker(
	pgx *pgxpool.Pool,
	fileEngine *engine.FileEngine,
	config DeploymentConfig,
) *scheduler.MaintenanceRunner[*Deployment] {
	repo := NewPostgresRepository(pgx)

	// Define retriever: fetch expired deployments
	retriever := func(ctx context.Context) ([]*Deployment, error) {
		deployments, err := repo.GetExpired(ctx)
		if err != nil {
			return nil, err
		}
		// Convert slice of values to slice of pointers
		result := make([]*Deployment, len(deployments))
		for i := range deployments {
			result[i] = &deployments[i]
		}
		return result, nil
	}

	// Define executor: clean up physical files and mark as expired
	executor := func(ctx context.Context, deployment *Deployment) error {
		// Physical cleanup via FileEngine
		path := "deployments/" + deployment.ProjectID + "/" + deployment.Hash
		if err := fileEngine.Remove(ctx, path); err != nil {
			if config.Logger != nil {
				config.Logger.Error("failed to remove deployment files", "id", deployment.ID, "err", err)
			}
			return err
		}

		// Logical cleanup via Repository
		return repo.UpdateStatus(ctx, UpdateDeploymentStatus{
			ID:     deployment.ID,
			Status: StatusExpired,
		})
	}

	// Define error handler for executor failures
	onError := func(deployment *Deployment, err error) {
		if config.Logger != nil {
			config.Logger.Error("deployment cleanup failed", "id", deployment.ID, "err", err)
		}
	}

	// Instantiate the generic maintenance runner
	return scheduler.NewMaintenanceRunner(
		config.MaintenanceInterval,
		config.CleanupConcurrency,
		retriever,
		executor,
		onError,
		config.Logger,
	)
}

// Deprecated: Use InitHttp instead. Kept for backwards compatibility.
func Init(mux *http.ServeMux, pgx *pgxpool.Pool, workerPool *worker.DeploymentWorkerPool) {
	InitHttp(mux, pgx, workerPool)
}
