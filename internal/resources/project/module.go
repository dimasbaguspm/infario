package project

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(mux *http.ServeMux, pgx *pgxpool.Pool) {
	repo := NewPostgresRepository(pgx)
	service := NewService(repo)

	RegisterRoutes(mux, *service)
}
