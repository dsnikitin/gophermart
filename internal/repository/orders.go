package repository

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Orders struct {
	db *pgxpool.Pool
}

func NewOrders(db *pgxpool.Pool) *Orders {
	return &Orders{db: db}
}

const ordersUploadOrderSQL = `
	INSERT INTO gophermart.orders(number, status, accrual, uploaded_at, user_login)
	VALUES(@number, @status, @accrual, @uploadedAt, @login)
	ON CONFLICT(number) DO UPDATE
	SET number = @number
	RETURNING number, status, accrual, uploaded_at, user_login
`

func (r *Orders) UploadOrder(ctx context.Context, newOrder models.Order) (models.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.Order{}, errors.Wrap(err, "begin tx")
	}
	defer tx.Rollback(ctx)

	row := r.db.QueryRow(ctx, ordersUploadOrderSQL, pgx.NamedArgs{
		"number":     newOrder.Number,
		"status":     newOrder.Status,
		"accrual":    newOrder.Accrual,
		"uploadedAt": newOrder.UploadedAt,
		"login":      newOrder.UserLogin,
	})

	var order models.Order
	if err := row.Scan(order.ScanFields()...); err != nil {
		return models.Order{}, errors.Wrap(err, "scan order")
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Order{}, errors.Wrap(err, "commit tx")
	}

	return order, nil
}

// const ordersGetOrderSQL = `
// 	SELECT number, status, accrual, uploaded_at, user_login
// 	FROM gophermart.orders
// 	WHERE number = @number
// `

// func (r *Orders) GetOrder(ctx context.Context, number string) (models.Order, error) {
// 	var order models.Order
// 	err := r.db.QueryRow(ctx, ordersGetOrderSQL, pgx.NamedArgs{"number": number}).Scan(order.ScanFields()...)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return models.Order{}, errx.ErrNotFound
// 		}

// 		return models.Order{}, errors.Wrap(err, "scan order")
// 	}

// 	return order, nil
// }

const ordersGetUserOrdersSQL = `
	SELECT number, status, accrual, uploaded_at, user_login
	FROM gophermart.orders
	WHERE user_login = @login
	ORDER BY uploaded_at DESC
`

func (r *Orders) GetUserOrders(ctx context.Context, login string) ([]models.Order, error) {
	rows, err := r.db.Query(ctx, ordersGetUserOrdersSQL, pgx.NamedArgs{"login": login})
	if err != nil {
		return nil, errors.Wrap(err, "do query")
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
