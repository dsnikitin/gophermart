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

type Users struct {
	db *pgxpool.Pool
}

func NewUsers(db *pgxpool.Pool) *Users {
	return &Users{db: db}
}

const userCreateSQL = `
	INSERT INTO gophermart.users(login, password)
	VALUES (@login, @password)
`

func (r *Users) Create(ctx context.Context, login, password string) error {
	_, err := r.db.Exec(ctx, userCreateSQL, pgx.NamedArgs{"login": login, "password": password})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return errx.ErrAlreadyExists
		}
	}

	return errors.Wrap(err, "exec")
}

const userGetSQL = `
	SELECT login, password
	FROM gophermart.users
	WHERE login = @login
`

func (r *Users) Get(ctx context.Context, login string) (models.User, error) {
	row := r.db.QueryRow(ctx, userGetSQL, pgx.NamedArgs{"login": login})

	var user models.User
	if err := row.Scan(&user.Login, &user.Password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, errx.ErrNotFound
		}

		return models.User{}, errors.Wrap(err, "scan user row")
	}

	return user, nil
}
