package service

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/pkg/errors"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, login string) (models.Balance, error)
	CreatWithdrawal(ctx context.Context, login, orderNumber string, sum float64) error
	GetWithdrawals(ctx context.Context, login string) ([]models.Withdrawal, error)
	// LockUser(ctx context.Context, login string) error
	LockBalance(ctx context.Context, login string) error
}

type TransactionProvider interface {
	Do(ctx context.Context, fn func(BalanceRepository) error) error
}

type BalanceService struct {
	r  BalanceRepository
	tx TransactionProvider
}

func NewBalance(r BalanceRepository, tx TransactionProvider) *BalanceService {
	return &BalanceService{r: r, tx: tx}
}

func (s *BalanceService) GetBalance(ctx context.Context, login string) (models.Balance, error) {
	return s.r.GetBalance(ctx, login)
}

func (s *BalanceService) GetWithdrawals(ctx context.Context, login string) ([]models.Withdrawal, error) {
	return s.r.GetWithdrawals(ctx, login)
}

func (s *BalanceService) Withdraw(ctx context.Context, login string, req models.WithdrawRequest) error {
	if err := checkNumber(req.Order); err != nil {
		return errors.Wrap(err, "check number")
	}

	err := s.tx.Do(ctx, func(rtx BalanceRepository) error {
		// if err := rtx.LockUser(ctx, login); err != nil {
		// 	return errors.Wrap(err, "lock user")
		// }

		if err := rtx.LockBalance(ctx, login); err != nil {
			return errors.Wrap(err, "lock balance")
		}

		if err := rtx.CreatWithdrawal(ctx, login, req.Order, req.Sum); err != nil {
			return errors.Wrap(err, "create withdrawal")
		}

		balance, err := rtx.GetBalance(ctx, login)
		if err != nil {
			return errors.Wrap(err, "get balance")
		}

		if balance.Current < 0 {
			return errx.ErrInsufficientFunds
		}

		return nil
	})

	return errors.Wrap(err, "do tx")
}
