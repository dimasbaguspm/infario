package resources

import (
	"net/http"

	"github.com/dimasbaguspm/infario/internal/resources/deployment"
	"github.com/dimasbaguspm/infario/internal/resources/project"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(mux *http.ServeMux, db *pgxpool.Pool) {
	project.Init(mux, db)
	deployment.Init(mux, db)
}
