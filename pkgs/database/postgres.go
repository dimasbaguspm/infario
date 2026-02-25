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
	// Get all resource-specific migration directories
	migrationsDir := filepath.Clean("migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		slog.Error("Failed to read migrations directory", "error", err)
		return err
	}

	// Run migrations from each resource directory in order
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		resourceDir := filepath.Join(migrationsDir, entry.Name())
		slog.Info("Running migrations", "resource", entry.Name())

		migrator, err := migrate.New(fmt.Sprintf("file://%s", filepath.Clean(resourceDir)), cfg.DBDSN)
		if err != nil {
			slog.Error("Failed to initialize migrator", "resource", entry.Name(), "error", err)
			return err
		}

		if err := migrator.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				slog.Info("No changes detected for resource", "resource", entry.Name())
				migrator.Close()
				continue
			}
			slog.Error("Migration failed for resource", "resource", entry.Name(), "error", err)
			migrator.Close()
			return err
		}

		migrator.Close()
		slog.Info("Migration completed for resource", "resource", entry.Name())
	}

	return nil
}
