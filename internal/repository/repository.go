package repository

import "github.com/jackc/pgx/v5/pgxpool"

type Repository struct {
	*Users
	*Orders
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		Users:  NewUsers(db),
		Orders: NewOrders(db),
	}
}
