package resources

import (
	"net/http"

	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/dimasbaguspm/infario/internal/resources/deployment"
	"github.com/dimasbaguspm/infario/internal/resources/project"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func InitHttps(mux *http.ServeMux, db *pgxpool.Pool, redisClient *redis.Client, fileEngine *engine.FileEngine) {
	project.Init(mux, db)
	deployment.InitHttp(mux, db, redisClient, fileEngine)
}
