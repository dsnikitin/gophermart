package repository

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Repository struct {
	User    *UserRepository
	Order   *OrderRepository
	Balance *BalanceRepository
	Accrual *AccrualRepository
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		User:    NewUserRepository(db),
		Order:   NewOrderRepository(db),
		Balance: NewBalanceRepository(db),
		Accrual: NewAccrualRepository(db),
	}
}

type baseRepo struct {
	db *pgxpool.Pool
	tx pgx.Tx
}

func (r *baseRepo) doTx(ctx context.Context, fn func(*baseRepo) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logger.Log.Errorw("Failed to rollback tx", "error", err.Error())
		}
	}()

	rtx := &baseRepo{db: r.db, tx: tx}

	if err = fn(rtx); err != nil {
		return errors.Wrap(err, "do job into tx")
	}

	err = tx.Commit(ctx)
	return errors.Wrap(err, "commit tx")
}

func (r *baseRepo) exec(ctx context.Context, sql string, args pgx.NamedArgs) (res pgconn.CommandTag, err error) {
	if r.tx != nil {
		res, err = r.tx.Exec(ctx, sql, args)
	} else {
		res, err = r.db.Exec(ctx, sql, args)
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return pgconn.CommandTag{}, errx.ErrAlreadyExists
		}

		return pgconn.CommandTag{}, errors.Wrap(err, "pool exec error")
	}

	return res, nil
}

func queryOne[T any](
	ctx context.Context, r *baseRepo, sql string, args pgx.NamedArgs, fieldsPointer func(*T) []any,
) (obj T, err error) {
	if r.tx != nil {
		err = r.tx.QueryRow(ctx, sql, args).Scan(fieldsPointer(&obj)...)
	} else {
		err = r.db.QueryRow(ctx, sql, args).Scan(fieldsPointer(&obj)...)
	}

	if err == pgx.ErrNoRows {
		return obj, errx.ErrNotFound
	}

	return obj, errors.Wrap(err, "scan db row")
}

func queryMany[T any](
	ctx context.Context, r *baseRepo, sql string, args pgx.NamedArgs, fieldsPointer func(*T) []any,
) (objs []T, err error) {
	var rows pgx.Rows
	if r.tx != nil {
		rows, err = r.tx.Query(ctx, sql, args)
	} else {
		rows, err = r.db.Query(ctx, sql, args)
	}

	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	for rows.Next() {
		var obj T
		if err := rows.Scan(fieldsPointer(&obj)...); err != nil {
			return nil, errors.Wrap(err, "scan db row")
		}

		objs = append(objs, obj)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iteration error")
	}

	return objs, nil
}
