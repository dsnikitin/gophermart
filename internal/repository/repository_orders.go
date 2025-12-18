package repository

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Order struct {
	baseRepo
}

func NewOrder(db *pgxpool.Pool) *Order {
	return &Order{baseRepo{db: db}}
}

const uploadOrderSQL = `
	INSERT INTO gophermart.orders(number, status, uploaded_at, accrual, user_login)
	VALUES(@number, @status, @uploadedAt, @accrual, @login)
	ON CONFLICT(number) DO UPDATE
	SET number = @number
	RETURNING number, status, uploaded_at, accrual, user_login
`

func (r *Order) UploadOrder(ctx context.Context, newOrder models.OrderDB) (*models.OrderDB, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "begin tx")
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, uploadOrderSQL, pgx.NamedArgs{
		"number":     newOrder.Number,
		"status":     newOrder.Status,
		"uploadedAt": newOrder.UploadedAt,
		"accrual":    newOrder.Accrual,
		"login":      newOrder.UserLogin,
	})

	var order models.OrderDB
	if err := row.Scan(order.ScanFields()...); err != nil {
		return nil, errors.Wrap(err, "scan order")
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "commit tx")
	}

	return &order, nil
}

const getOrdersSQL = `
	SELECT number, status, uploaded_at, accrual, user_login
	FROM gophermart.orders
	WHERE user_login = @login
	ORDER BY uploaded_at DESC
`

func (r *Order) GetOrders(ctx context.Context, login string) ([]*models.OrderDB, error) {
	rows, err := r.query(ctx, getOrdersSQL, pgx.NamedArgs{"login": login})
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	var orders []*models.OrderDB
	for rows.Next() {
		var order models.OrderDB
		if err := rows.Scan(order.ScanFields()...); err != nil {
			return nil, errors.Wrap(err, "scan order")
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iteration error")
	}

	return orders, nil
}
