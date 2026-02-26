package resources

import (
	"context"
	"log/slog"

	"github.com/dimasbaguspm/infario/internal/gateway"
	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/dimasbaguspm/infario/internal/resources/deployment/workers"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// InitWorkers initializes and starts all background workers.
func InitWorkers(
	ctx context.Context,
	db *pgxpool.Pool,
	redisClient *redis.Client,
	fileEngine *engine.FileEngine,
	ng *gateway.NginxGateway,
) {
	logger := slog.Default()

	workers.StartDeploymentConsumer(ctx, db, ng, redisClient, fileEngine, logger)
	workers.StartExpiryCleanup(ctx, db, fileEngine, ng, logger)
}
