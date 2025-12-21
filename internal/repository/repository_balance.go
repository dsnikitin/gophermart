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

func NewBalance(db *pgxpool.Pool) *BalanceRepository {
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
	row := r.queryRow(ctx, getBalanceSQL, pgx.NamedArgs{"login": login})

	var balance models.Balance
	if err := row.Scan(balance.ScanFields()...); err != nil {
		return models.Balance{}, errors.Wrap(err, "scan balance")
	}

	return balance, nil
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
	rows, err := r.query(ctx, getWithdrawalsSQL, pgx.NamedArgs{"login": login})
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(w.ScanFields()...); err != nil {
			return nil, errors.Wrap(err, "scan withdrawal")
		}

		withdrawals = append(withdrawals, w)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iteration error")
	}

	return withdrawals, nil
}

// const lockUserSQL = `
// 	SELECT 1
// 	FROM gophermart.users
// 	WHERE user_login = @login
// 	FOR UPDATE
// `

// func (r *BalanceRepository) LockUser(ctx context.Context, login string) error {
// 	row := r.queryRow(ctx, lockUserSQL, pgx.NamedArgs{"login": login})

// 	var locked int
// 	if err := row.Scan(&locked); err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return errx.ErrNotFound
// 		}

// 		return errors.Wrap(err, "scan user locked")
// 	}

// 	return nil
// }

const advisoryLockSQL = `
    SELECT pg_advisory_xact_lock(hashtext(@login))
`

func (r *BalanceRepository) LockBalance(ctx context.Context, login string) error {
	// var _ int8
	_, err := r.exec(ctx, advisoryLockSQL, pgx.NamedArgs{"login": login})
	return errors.Wrap(err, "exec")
}

func (r *BalanceRepository) DoTx(ctx context.Context, fn func(*BalanceRepository) error) error {
	return r.baseRepo.doTx(ctx, func(brTx *baseRepo) error {
		balanceTx := &BalanceRepository{baseRepo: brTx}
		return fn(balanceTx)
	})
}
