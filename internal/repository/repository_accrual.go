package repository

import (
	"context"
	"time"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/consts/order"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type AccrualRepository struct {
	*baseRepo
}

func NewAccrualRepository(db *pgxpool.Pool) *AccrualRepository {
	return &AccrualRepository{&baseRepo{db: db}}
}

const updateOrderSQL = `
	UPDATE gophermart.orders
	SET status = @status,
		accrual = COALESCE(@accrual, accrual)
	WHERE number = @number
`

func (r *AccrualRepository) UpdateOrder(ctx context.Context, number string, status order.Status, accrual *float64) error {
	res, err := r.exec(ctx, updateOrderSQL, pgx.NamedArgs{
		"number":  number,
		"status":  status,
		"accrual": accrual,
	})
	if err != nil {
		return errors.Wrap(err, "exec")
	}

	if res.RowsAffected() == 0 {
		return errx.ErrNotFound
	}

	return nil
}

const getOldersOrderWithLock = `
	SELECT number, status, uploaded_at, accrual, user_login
	FROM gophermart.orders
	WHERE status = @status
	ORDER BY uploaded_at
	LIMIT 1
	FOR UPDATE SKIP LOCKED
`

func (r *AccrualRepository) GetOldestOrderWithLock(ctx context.Context, status order.Status) (models.Order, error) {
	args := pgx.NamedArgs{"status": status}
	fieldsPointer := func(o *models.Order) []any { return o.ScanFields() }

	order, err := queryOne(ctx, r.baseRepo, getOldersOrderWithLock, args, fieldsPointer)
	return order, errors.Wrap(err, "query one")

}

const getStaleOrdersSQL = `
	SELECT number, status, uploaded_at, accrual, user_login
	FROM gophermart.orders
	WHERE status = @status AND uploaded_at <= @threshold
`

func (r *AccrualRepository) GetStaleOrders(
	ctx context.Context, threshold time.Time, status order.Status,
) ([]models.Order, error) {
	args := pgx.NamedArgs{"status": status, "threshold": threshold}
	fieldsPointer := func(o *models.Order) []any { return o.ScanFields() }

	orders, err := queryMany(ctx, r.baseRepo, getStaleOrdersSQL, args, fieldsPointer)
	return orders, errors.Wrap(err, "query many")
}

func (r *AccrualRepository) DoTx(ctx context.Context, fn func(*AccrualRepository) error) error {
	return r.baseRepo.doTx(ctx, func(brTx *baseRepo) error {
		balanceTx := &AccrualRepository{baseRepo: brTx}
		return fn(balanceTx)
	})
}
