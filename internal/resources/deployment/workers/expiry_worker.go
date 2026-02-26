package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/dimasbaguspm/infario/internal/gateway"
	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/dimasbaguspm/infario/internal/platform/scheduler"
	"github.com/dimasbaguspm/infario/internal/resources/deployment"
	"github.com/dimasbaguspm/infario/pkgs/request"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// ExpiryCheckInterval defines how often to check for expired deployments
	ExpiryCheckInterval = 1 * time.Hour
	// ExpiryCleanupConcurrency limits concurrent cleanup workers
	ExpiryCleanupConcurrency = 5
)

// StartExpiryCleanup initializes and starts the expiry cleanup worker.
// This periodically looks up expired deployments, removes files, and marks them as inactive.
func StartExpiryCleanup(
	ctx context.Context,
	db *pgxpool.Pool,
	fileEngine *engine.FileEngine,
	tg *gateway.TraefikGateway,
	logger *slog.Logger,
) {
	repo := deployment.NewPostgresRepository(db)

	retriever := func(ctx context.Context) ([]*deployment.Deployment, error) {
		deployments, err := repo.GetExpired(ctx)
		if err != nil {
			return nil, err
		}
		result := make([]*deployment.Deployment, len(deployments))
		for i := range deployments {
			result[i] = &deployments[i]
		}
		return result, nil
	}

	// Define executor: clean up physical files and mark as expired
	executor := func(ctx context.Context, d *deployment.Deployment) error {
		// Physical cleanup via FileEngine
		path := "deployments/" + d.ProjectID + "/" + d.Hash
		if err := fileEngine.Remove(ctx, path); err != nil {
			if logger != nil {
				logger.Error("failed to remove deployment files", "id", d.ID, "err", err)
			}
			return err
		}

		// Logical cleanup via Repository
		if err := repo.UpdateStatus(ctx, deployment.UpdateDeploymentStatus{
			ID:     d.ID,
			Status: deployment.StatusExpired,
		}); err != nil {
			return err
		}

		// Regenerate Traefik config after marking deployment as expired
		if tg != nil {
			status := deployment.StatusReady
			readyDeps, err := repo.GetPaged(ctx, deployment.GetPagedDeployment{
				PagingParams: request.PagingParams{PageNumber: 1, PageSize: 100},
				ProjectID:    &d.ProjectID,
				Status:       &status,
			})
			if err == nil && len(readyDeps.Items) > 0 {
				// Convert to GatewayDeployment format
				deps := make([]gateway.GatewayDeployment, len(readyDeps.Items))
				for i, rd := range readyDeps.Items {
					projectName := ""
					if rd.ProjectName != nil {
						projectName = *rd.ProjectName
					}
					deps[i] = gateway.GatewayDeployment{
						ID:          rd.ID,
						Hash:        rd.Hash,
						ProjectID:   rd.ProjectID,
						ProjectName: projectName,
					}
				}
				if readyDeps.Items[0].ProjectName != nil {
					tg.WriteProjectConfig(d.ProjectID, *readyDeps.Items[0].ProjectName, deps)
				}
			} else if err == nil {
				// No ready deployments left, remove config
				tg.RemoveProjectConfig(d.ProjectID)
			}
		}

		return nil
	}

	// Define error handler for executor failures
	onError := func(d *deployment.Deployment, err error) {
		if logger != nil {
			logger.Error("deployment cleanup failed", "id", d.ID, "err", err)
		}
	}

	// Create and start the maintenance runner
	runner := scheduler.NewMaintenanceRunner(
		ExpiryCheckInterval,
		ExpiryCleanupConcurrency,
		retriever,
		executor,
		onError,
		logger,
	)

	if logger != nil {
		logger.InfoContext(ctx, "expiry cleanup worker started", "interval", ExpiryCheckInterval, "concurrency", ExpiryCleanupConcurrency)
	}
	go runner.Start(ctx)
}
