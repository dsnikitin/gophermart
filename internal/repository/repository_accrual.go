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
	var order models.Order
	row := r.queryRow(ctx, getOldersOrderWithLock, pgx.NamedArgs{"status": status})
	if err := row.Scan(order.ScanFields()...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, errx.ErrNotFound
		}
	}

	return order, nil
}

const getStaleOrdersSQL = `
	SELECT number, status, uploaded_at, accrual, user_login
	FROM gophermart.orders
	WHERE status = ANY(@statuses) AND uploaded_at <= @threshold
`

func (r *AccrualRepository) GetStaleOrders(
	ctx context.Context, threshold time.Time, statuses ...order.Status,
) ([]models.Order, error) {
	rows, err := r.query(ctx, getStaleOrdersSQL, pgx.NamedArgs{"statuses": statuses, "threshold": threshold})
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(order.ScanFields()...); err != nil {
			return nil, errors.Wrap(err, "scan order")
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iteration error")
	}

	return orders, nil
}

func (r *AccrualRepository) DoTx(ctx context.Context, fn func(*AccrualRepository) error) error {
	return r.baseRepo.doTx(ctx, func(brTx *baseRepo) error {
		balanceTx := &AccrualRepository{baseRepo: brTx}
		return fn(balanceTx)
	})
}
