package deployment

import (
	"net/http"

	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func InitHttp(mux *http.ServeMux, pgx *pgxpool.Pool, redisClient *redis.Client, fileEngine *engine.FileEngine) {
	repo := NewPostgresRepository(pgx)
	service := NewService(repo, redisClient, fileEngine)
	RegisterRoutes(mux, *service)
}
