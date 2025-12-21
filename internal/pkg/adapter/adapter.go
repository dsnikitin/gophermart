package adapter

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/repository"
	"github.com/dsnikitin/gophermart/internal/service"
)

type BalanceTxAdapter struct {
	repo *repository.BalanceRepository
}

func NewBalanceTxAdapter(repo *repository.BalanceRepository) *BalanceTxAdapter {
	return &BalanceTxAdapter{repo: repo}
}

func (a *BalanceTxAdapter) Do(ctx context.Context, fn func(service.BalanceRepository) error) error {
	return a.repo.DoTx(ctx, func(txRepo *repository.BalanceRepository) error {
		return fn(txRepo)
	})
}
