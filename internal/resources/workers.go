package resources

import (
	"context"
	"log/slog"
	"time"

	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/dimasbaguspm/infario/internal/resources/deployment"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InitWorkers initializes and starts all background maintenance tasks.
// This should be called from main.go after RegisterHttps to set up cleanup workers.
func InitWorkers(
	ctx context.Context,
	db *pgxpool.Pool,
	fileEngine *engine.FileEngine,
) {
	deploymentConfig := deployment.DeploymentConfig{
		MaintenanceInterval: 1 * time.Hour,
		CleanupConcurrency:  5,
		Logger:              slog.Default(),
	}

	deploymentRunner := deployment.InitWorker(db, fileEngine, deploymentConfig)
	if deploymentRunner != nil {
		go deploymentRunner.Start(ctx)
	}
}
