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

type UserRepository struct {
	baseRepo
}

func NewUser(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{baseRepo{db: db}}
}

const createUserSQL = `
	INSERT INTO gophermart.users(login, password)
	VALUES (@login, @password)
`

func (r *UserRepository) CreateUser(ctx context.Context, user models.User) error {
	_, err := r.exec(ctx, createUserSQL, pgx.NamedArgs{"login": user.Login, "password": user.Password})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return errx.ErrAlreadyExists
		}
	}

	return nil
}

const getUserSQL = `
	SELECT login, password
	FROM gophermart.users
	WHERE login = @login
`

func (r *UserRepository) GetUser(ctx context.Context, login string) (models.User, error) {
	row := r.queryRow(ctx, getUserSQL, pgx.NamedArgs{"login": login})

	var user models.User
	if err := row.Scan(user.ScanFields()...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, errx.ErrNotFound
		}

		return models.User{}, errors.Wrap(err, "scan user")
	}

	return user, nil
}
