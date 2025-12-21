package repository

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Repository struct {
	User    *UserRepository
	Order   *OrderRepository
	Balance *BalanceRepository
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		User:    NewUser(db),
		Order:   NewOrder(db),
		Balance: NewBalance(db),
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

func (r *baseRepo) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if r.tx != nil {
		return r.tx.Query(ctx, sql, args...)
	}
	return r.db.Query(ctx, sql, args...)
}

func (r *baseRepo) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if r.tx != nil {
		return r.tx.QueryRow(ctx, sql, args...)
	}
	return r.db.QueryRow(ctx, sql, args...)
}

func (r *baseRepo) exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if r.tx != nil {
		return r.tx.Exec(ctx, sql, args...)
	}
	return r.db.Exec(ctx, sql, args...)
}
