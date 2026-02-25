package deployment

import (
	"net/http"

	"github.com/dimasbaguspm/infario/internal/platform/worker"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(mux *http.ServeMux, pgx *pgxpool.Pool, workerPool *worker.DeploymentWorkerPool) {
	repo := NewPostgresRepository(pgx)
	service := NewService(repo, workerPool)

	RegisterRoutes(mux, *service)
}
