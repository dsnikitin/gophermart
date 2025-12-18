package repository

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Balance struct {
	*baseRepo
}

func NewBalance(db *pgxpool.Pool) *Balance {
	return &Balance{&baseRepo{db: db}}
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

func (r *Balance) GetBalance(ctx context.Context, login string) (*models.BalanceDB, error) {
	row := r.queryRow(ctx, getBalanceSQL, pgx.NamedArgs{"login": login})

	var balance models.BalanceDB
	if err := row.Scan(balance.ScanFields()...); err != nil {
		return nil, errors.Wrap(err, "scan balance")
	}

	return &balance, nil
}

const createWithdrawalSQL = `
	INSERT INTO gophermart.withdrawals(order_number, user_login, amount, processed_at)
	VALUES(@orderNumber, @userLogin, @amount, @processedAt)
`

func (r *Balance) CreatWithdrawal(ctx context.Context, login string, withdrawal models.WithdrawalDB) error {
	_, err := r.exec(ctx, createWithdrawalSQL, pgx.NamedArgs{
		"orderNumber": withdrawal.OrderNumber,
		"userLogin":   login,
		"amount":      withdrawal.Amount,
		"processedAt": withdrawal.ProcessedAt,
	})

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return errx.ErrAlreadyExists
	}

	return errors.Wrap(err, "exec")
}

const getWithdrawalsSQL = `
	SELECT order_number, amount, processed_at
	FROM gophermart.withdrawals
	WHERE user_login = @login
`

func (r *Balance) GetWithdrawals(ctx context.Context, login string) ([]*models.WithdrawalDB, error) {
	rows, err := r.query(ctx, getWithdrawalsSQL, pgx.NamedArgs{"login": login})
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	var withdrawals []*models.WithdrawalDB
	for rows.Next() {
		var w models.WithdrawalDB
		if err := rows.Scan(w.ScanFields()...); err != nil {
			return nil, errors.Wrap(err, "scan withdrawal")
		}

		withdrawals = append(withdrawals, &w)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iteration error")
	}

	return withdrawals, nil
}

const lockUserSQL = `
	SELECT 1
	FROM gophermart.users
	WHERE user_login = @login
	FOR UPDATE
`

func (r *Balance) LockUser(ctx context.Context, login string) error {
	row := r.queryRow(ctx, lockUserSQL, pgx.NamedArgs{"login": login})

	var locked int
	if err := row.Scan(&locked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errx.ErrNotFound
		}

		return errors.Wrap(err, "scan user locked")
	}

	return nil
}

func (r *Balance) DoTx(ctx context.Context, fn func(*Balance) error) error {
	return r.baseRepo.doTx(ctx, func(brTx *baseRepo) error {
		balanceTx := &Balance{baseRepo: brTx}
		return fn(balanceTx)
	})
}
