package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dimasbaguspm/infario/pkgs/config"
	"github.com/dimasbaguspm/infario/pkgs/database"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	fmt.Printf("Loaded config: %+v\n", cfg)
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

	svr := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: svr,
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
}
