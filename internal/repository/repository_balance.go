package repository

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type BalanceRepository struct {
	*baseRepo
}

func NewBalanceRepository(db *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{&baseRepo{db: db}}
}

const getBalanceSQL = `
	WITH accrued AS (
		SELECT COALESCE(SUM(accrual), 0) AS accrued_sum
		FROM gophermart.orders
		WHERE user_login = @login
	),
	withdrawn AS (
		SELECT COALESCE(SUM(amount), 0) AS withdrawn_sum
		FROM gophermart.withdrawals
		WHERE user_login = @login
	)
	SELECT accrued_sum - withdrawn_sum AS current, withdrawn_sum
	FROM accrued, withdrawn
`

func (r *BalanceRepository) GetBalance(ctx context.Context, login string) (models.Balance, error) {
	args := pgx.NamedArgs{"login": login}
	fieldsPointer := func(b *models.Balance) []any { return b.ScanFields() }

	balance, err := queryOne(ctx, r.baseRepo, getBalanceSQL, args, fieldsPointer)
	return balance, errors.Wrap(err, "query one")
}

const createWithdrawalSQL = `
	INSERT INTO gophermart.withdrawals(order_number, user_login, amount)
	VALUES(@orderNumber, @userLogin, @amount)
`

func (r *BalanceRepository) CreatWithdrawal(ctx context.Context, login, orderNumber string, sum float64) error {
	_, err := r.exec(ctx, createWithdrawalSQL, pgx.NamedArgs{
		"orderNumber": orderNumber,
		"userLogin":   login,
		"amount":      sum,
	})

	return errors.Wrap(err, "exec")
}

const getWithdrawalsSQL = `
	SELECT order_number, amount, processed_at
	FROM gophermart.withdrawals
	WHERE user_login = @login
`

func (r *BalanceRepository) GetWithdrawals(ctx context.Context, login string) ([]models.Withdrawal, error) {
	args := pgx.NamedArgs{"login": login}
	fieldsPointer := func(o *models.Withdrawal) []any { return o.ScanFields() }

	withdrawals, err := queryMany(ctx, r.baseRepo, getWithdrawalsSQL, args, fieldsPointer)
	return withdrawals, errors.Wrap(err, "query many")
}

const advisoryLockSQL = `
    SELECT pg_advisory_xact_lock(hashtext(@login))
`

func (r *BalanceRepository) LockBalance(ctx context.Context, login string) error {
	_, err := r.exec(ctx, advisoryLockSQL, pgx.NamedArgs{"login": login})
	return errors.Wrap(err, "exec")
}

func (r *BalanceRepository) DoTx(ctx context.Context, fn func(*BalanceRepository) error) error {
	return r.baseRepo.doTx(ctx, func(brTx *baseRepo) error {
		balanceTx := &BalanceRepository{baseRepo: brTx}
		return fn(balanceTx)
	})
}
