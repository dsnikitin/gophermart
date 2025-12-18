package logger

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DevEnv  string = "dev"
	ProdEnv string = "prod"
)

type Config struct {
	Lvl     string `env:"LOG_LEVEL"`
	EnvType string `env:"LOG_ENV_TYPE"`
}

var Log = zap.Must(zap.NewDevelopment()).Sugar()

func Setup(cfg *Config) error {
	lvl, err := zap.ParseAtomicLevel(cfg.Lvl)
	if err != nil {
		return errors.Wrap(err, "parse level")
	}

	var zapCfg zap.Config
	switch cfg.EnvType {
	case DevEnv:
		zapCfg = zap.NewDevelopmentConfig()
	case ProdEnv:
		zapCfg = zap.NewProductionConfig()
	default:
		return errors.Errorf("unknown environment %s", cfg.EnvType)
	}

	zapCfg.Level = lvl
	log, err := zapCfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return errors.Wrap(err, "build logger")
	}

	Log = log.Sugar()

	return nil
}
