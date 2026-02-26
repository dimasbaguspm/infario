package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/dimasbaguspm/infario/docs" // Import generated docs
	"github.com/dimasbaguspm/infario/internal/gateway"
	"github.com/dimasbaguspm/infario/internal/platform/engine"
	"github.com/dimasbaguspm/infario/internal/resources"
	"github.com/dimasbaguspm/infario/pkgs/config"
	"github.com/dimasbaguspm/infario/pkgs/database"
	"github.com/dimasbaguspm/infario/pkgs/redis"
	"github.com/dimasbaguspm/infario/pkgs/response"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Infario API
// @version         1.0
// @description     The core API for the Infario Cloud Platform.
// @host            localhost:8080
// @BasePath        /
// @schemes         http https
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	db, err := database.NewPostgres(ctx, database.Config{
		DSN:             cfg.DBDSN,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		ConnMaxLifetime: cfg.DBConnLifetime,
	})

	if err != nil {
		slog.Error("Could not connect to database", "Error", err)
		os.Exit(1)
	}

	defer db.Close()

	if err := database.RunMigrations(*cfg); err != nil {
		slog.Error("Migration failed", "Error", err)
		os.Exit(1)
	}

	redisClient, err := redis.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("Could not connect to Redis", "Error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	fileEngine := engine.NewFileEngine("./storage")
	ng := gateway.NewNginxGateway("./nginx/conf.d", cfg.NginxDomain, "./storage")

	mux := http.NewServeMux()

	// Initialize background workers (consumers that drain from Redis)
	resources.InitWorkers(ctx, db, redisClient, fileEngine, ng)

	// Initialize HTTP routes (service emits directly to Redis)
	resources.InitHttps(mux, db, redisClient, fileEngine)

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusNotFound, "The requested resource was not found")
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("HTTP server started")

	<-ctx.Done()

	slog.Info("Shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed, forcing exit", "err", err)
	} else {
		slog.Info("Server stopped")
	}

	slog.Info("Background workers shut down via context cancellation")
}
