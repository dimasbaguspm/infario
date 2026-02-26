package resources

import (
	"context"
	"log/slog"

	"github.com/dimasbaguspm/infario/internal/gateway"
	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/dimasbaguspm/infario/internal/resources/deployment/workers"
	"github.com/dimasbaguspm/infario/pkgs/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// InitWorkers initializes and starts all background workers.
func InitWorkers(
	ctx context.Context,
	db *pgxpool.Pool,
	fileEngine *engine.FileEngine,
	cfg *config.Config,
	redisClient *redis.Client,
) {
	logger := slog.Default()
	tg := gateway.NewTraefikGateway("./traefik/dynamic", cfg.TraefikDomain)

	workers.StartDeploymentConsumer(ctx, db, tg, redisClient, logger)
	workers.StartExpiryCleanup(ctx, db, fileEngine, tg, logger)
}
