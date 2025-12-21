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

type OrderRepository struct {
	*baseRepo
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{&baseRepo{db: db}}
}

const uploadOrderSQL = `
	INSERT INTO gophermart.orders(number, user_login)
	VALUES(@number, @login)
`

func (r *OrderRepository) UploadOrder(ctx context.Context, login, number string) error {
	_, err := r.exec(ctx, uploadOrderSQL, pgx.NamedArgs{"number": number, "login": login})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return errx.ErrAlreadyExists
		}
	}

	return nil
}

const getOrderSQL = `
	SELECT number, status, uploaded_at, accrual, user_login
	FROM gophermart.orders
	WHERE number = @number
`

func (r *OrderRepository) GetOrder(ctx context.Context, number string) (models.Order, error) {
	row := r.queryRow(ctx, getOrderSQL, pgx.NamedArgs{"number": number})

	var order models.Order
	if err := row.Scan(order.ScanFields()...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, errx.ErrNotFound
		}

		return models.Order{}, errors.Wrap(err, "scan order")
	}

	return order, nil
}

const getOrdersSQL = `
	SELECT number, status, uploaded_at, accrual, user_login
	FROM gophermart.orders
	WHERE user_login = @login
	ORDER BY uploaded_at DESC
`

func (r *OrderRepository) GetOrders(ctx context.Context, login string) ([]models.Order, error) {
	rows, err := r.query(ctx, getOrdersSQL, pgx.NamedArgs{"login": login})
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
