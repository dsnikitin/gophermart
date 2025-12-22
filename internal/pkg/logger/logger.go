package logger

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Lvl          string `env:"LOG_LEVEL"`
	IsProduction bool   `env:"LOG_IS_PROD"`
}

var Log = zap.Must(zap.NewDevelopment()).Sugar()

func Setup(cfg *Config) error {
	lvl, err := zap.ParseAtomicLevel(cfg.Lvl)
	if err != nil {
		return errors.Wrap(err, "parse level")
	}

	var zapCfg zap.Config
	if cfg.IsProduction {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	zapCfg.Level = lvl
	log, err := zapCfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return errors.Wrap(err, "build logger")
	}

	Log = log.Sugar()

	return nil
}
