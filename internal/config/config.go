package config

import (
	"flag"

	"github.com/dsnikitin/gophermart/internal/pkg/auth"
	"github.com/dsnikitin/gophermart/internal/pkg/db"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"

	"github.com/caarlos0/env"
	"github.com/pkg/errors"
)

type Config struct {
	ServerAddr        string `env:"RUN_ADDRESS"`
	AccrualSystemAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DB                *db.Config
	Auth              *auth.Config
	Log               *logger.Config
}

func New() (*Config, error) {
	cfg := &Config{
		DB:   &db.Config{},
		Auth: &auth.Config{},
		Log:  &logger.Config{},
	}

	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "server host:port")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", "localhost:8090", "accrual system address")
	flag.StringVar(&cfg.DB.URI, "d", "", "database URI")
	flag.StringVar(&cfg.DB.MigrationsPath, "m", "migrations", "migrations path")
	flag.StringVar(&cfg.Log.Lvl, "l", "info", "log level")
	flag.StringVar(&cfg.Log.EnvType, "e", logger.DevEnv, "environment type")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return nil, errors.Wrap(err, "parse env")
	}

	if err := env.Parse(cfg.DB); err != nil {
		return nil, errors.Wrap(err, "parse auth env")
	}

	if err := env.Parse(cfg.Auth); err != nil {
		return nil, errors.Wrap(err, "parse auth env")
	}

	if err := env.Parse(cfg.Log); err != nil {
		return nil, errors.Wrap(err, "parse log env")
	}

	return cfg, nil
}
