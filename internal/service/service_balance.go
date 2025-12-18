package service

import (
	"context"
	"time"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/pkg/errors"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, login string) (*models.BalanceDB, error)
	CreatWithdrawal(ctx context.Context, login string, withdrawal models.WithdrawalDB) error
	GetWithdrawals(ctx context.Context, login string) ([]*models.WithdrawalDB, error)
	LockUser(ctx context.Context, login string) error
}

type TransactionProvider interface {
	Do(ctx context.Context, fn func(BalanceRepository) error) error
}

type BalanceService struct {
	repo BalanceRepository
	tx   TransactionProvider
}

func NewBalance(repo BalanceRepository, tx TransactionProvider) *BalanceService {
	return &BalanceService{repo: repo, tx: tx}
}

func (s *BalanceService) GetBalance(ctx context.Context, login string) (models.BalanceResponse, error) {
	balance, err := s.repo.GetBalance(ctx, login)
	if err != nil {
		return models.BalanceResponse{}, errors.Wrap(err, "get balance")
	}

	return balance.ToBalanceResponse(), nil
}

func (s *BalanceService) Withdraw(ctx context.Context, login string, req models.WithdrawRequest) error {
	if err := checkNumber(req.Order); err != nil {
		return errors.Wrap(err, "check number")
	}

	err := s.tx.Do(ctx, func(rtx BalanceRepository) error {
		if err := rtx.LockUser(ctx, login); err != nil {
			return errors.Wrap(err, "lock user")
		}

		balance, err := rtx.GetBalance(ctx, login)
		if err != nil {
			return errors.Wrap(err, "get balance")
		}

		withdrawalSum := int64(req.Sum * 100)

		if balance.Current-withdrawalSum < 0 {
			return errx.ErrInsufficientFunds
		}

		err = rtx.CreatWithdrawal(ctx, login, models.WithdrawalDB{
			OrderNumber: req.Order,
			Amount:      withdrawalSum,
			ProcessedAt: time.Now().UTC(),
		})
		return errors.Wrap(err, "create withdrawal")
	})

	return errors.Wrap(err, "do tx")
}

func (s *BalanceService) GetWithdrawals(ctx context.Context, login string) ([]models.WithdrawalResponse, error) {
	withdrawals, err := s.repo.GetWithdrawals(ctx, login)
	if err != nil {
		return nil, errors.Wrap(err, "get withdrawals")
	}

	res := make([]models.WithdrawalResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		res = append(res, w.ToWithdrawalResponse())
	}

	return res, nil
}
