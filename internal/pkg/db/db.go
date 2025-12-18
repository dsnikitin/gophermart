package db

import (
	"context"
	"strings"
	"time"

	"github.com/dsnikitin/gophermart/internal/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Config struct {
	URI            string `env:"DATABASE_URI"`
	MigrationsPath string `env:"DATABASE_MIGRATIONS_PATH"`
}

func Connect(cfg *Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.URI)
	if err != nil {
		return nil, errors.Wrap(err, "new pgxpool")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.Wrap(err, "ping db")
	}

	logger.Log.Infow(
		"Successfuly connected to PostgresDB",
		"name", pool.Config().ConnConfig.Database,
		"host", pool.Config().ConnConfig.Host,
		"port", pool.Config().ConnConfig.Port,
	)

	return pool, nil
}

func ApplyMigrations(cfg *Config) error {
	logger.Log.Info("Applying migrations...")

	replacer := strings.NewReplacer("postgres://", "pgx://", "postgresql://", "pgx://")

	m, err := migrate.New("file://"+cfg.MigrationsPath, replacer.Replace(cfg.URI))
	if err != nil {
		return errors.Wrap(err, "create migrate instance")
	}
	defer m.Close()

	err = m.Up()
	switch {
	case err == nil:
		logger.Log.Info("Migrations successfully applied")
	case err == migrate.ErrNoChange:
		logger.Log.Info("No new migrations for applying")
	default:
		return errors.Wrap(err, "up migrations")
	}

	return nil
}
