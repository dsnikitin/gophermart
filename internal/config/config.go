package config

import (
	"flag"
	"time"

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
	Log               *logger.Config
	Auth              *auth.Config
}

func New() (*Config, error) {
	cfg := &Config{
		DB:   &db.Config{},
		Log:  &logger.Config{},
		Auth: &auth.Config{},
	}

	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "server host:port")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", "localhost:8090", "accrual system address")
	flag.StringVar(&cfg.DB.URI, "d", "", "database URI")
	flag.StringVar(&cfg.DB.MigrationsPath, "m", "migrations", "migrations path")
	flag.StringVar(&cfg.Log.Lvl, "l", "info", "log level")
	flag.BoolVar(&cfg.Log.IsProduction, "e", false, "is production flag")
	flag.StringVar(&cfg.Auth.CookieName, "c", "auth_token", "auth cookie name")
	flag.DurationVar(&cfg.Auth.TokenExp, "t", time.Hour*3, "auth token ttl")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return nil, errors.Wrap(err, "parse env")
	}

	if err := env.Parse(cfg.DB); err != nil {
		return nil, errors.Wrap(err, "parse db env")
	}

	if err := env.Parse(cfg.Log); err != nil {
		return nil, errors.Wrap(err, "parse log env")
	}

	if err := env.Parse(cfg.Auth); err != nil {
		return nil, errors.Wrap(err, "parse auth env")
	}

	return cfg, nil
}
