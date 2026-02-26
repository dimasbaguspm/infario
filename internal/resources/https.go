package resources

import (
	"net/http"

	"github.com/dimasbaguspm/infario/internal/platform/worker"
	"github.com/dimasbaguspm/infario/internal/resources/deployment"
	"github.com/dimasbaguspm/infario/internal/resources/project"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitHttps(mux *http.ServeMux, db *pgxpool.Pool, workerPool *worker.DeploymentWorkerPool) {
	project.Init(mux, db)
	deployment.InitHttp(mux, db, workerPool)
}
