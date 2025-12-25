package repository

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/jackc/pgx/v5"
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
	_, err := r.exec(ctx, uploadOrderSQL, pgx.NamedArgs{
		"number": number,
		"login":  login,
	})

	return errors.Wrap(err, "exec")
}

const getOrderSQL = `
	SELECT number, status, uploaded_at, accrual, user_login
	FROM gophermart.orders
	WHERE number = @number
`

func (r *OrderRepository) GetOrder(ctx context.Context, number string) (models.Order, error) {
	fieldsPointer := func(o *models.Order) []any { return o.ScanFields() }

	order, err := queryOne(ctx, r.baseRepo, getOrderSQL, pgx.NamedArgs{"number": number}, fieldsPointer)
	return order, errors.Wrap(err, "query one")
}

const getOrdersSQL = `
	SELECT number, status, uploaded_at, accrual, user_login
	FROM gophermart.orders
	WHERE user_login = @login
	ORDER BY uploaded_at DESC
`

func (r *OrderRepository) GetOrders(ctx context.Context, login string) ([]models.Order, error) {
	args := pgx.NamedArgs{"login": login}
	fieldsPointer := func(o *models.Order) []any { return o.ScanFields() }

	orders, err := queryMany(ctx, r.baseRepo, getOrdersSQL, args, fieldsPointer)
	return orders, errors.Wrap(err, "query many")
}
