package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dimasbaguspm/infario/pkgs/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func NewPostgres(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.DSN)

	if err != nil {
		slog.Error("Unable to parse database URL", "error", err)
		return nil, err
	}

	config.MaxConns = int32(cfg.MaxOpenConns)
	config.MaxConnLifetime = cfg.ConnMaxLifetime

	db, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	slog.Info("Connecting to the database...")

	if err := db.Ping(ctx); err != nil {
		slog.Error("Failed to connect to the database", "error", err)
		return nil, err
	}

	slog.Info("Successfully connected to the database")

	return db, nil
}

func RunMigrations(cfg config.Config) error {
	migrator, err := migrate.New(fmt.Sprintf("file://%s", filepath.Clean("migrations")), cfg.DBDSN)

	if err != nil {
		slog.Error("Something wrong with migrator", "error", err)
		os.Exit(1)
		return err
	}

	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("Migration is done, no changed schemas are detected")
			return nil
		}
		slog.Error("Unable to run migration", "error", err)
		return err
	}
	return nil
}
