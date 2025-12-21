package accrualer

import (
	"context"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/models"
)

type Accrualer struct {
	cfg     *Config
	accrual *http.Client
}

type Config struct {
	Addr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func New(cfg *Config) *Accrualer {
	return &Accrualer{
		accrual: http.DefaultClient,
		cfg:     cfg,
	}
}

func (a *Accrualer) GetAccrual(ctx context.Context, orderNumber string) (models.Accrual, error) {
	return models.Accrual{}, nil
}
