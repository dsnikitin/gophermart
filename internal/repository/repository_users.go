package repository

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type UserRepository struct {
	*baseRepo
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{&baseRepo{db: db}}
}

const createUserSQL = `
	INSERT INTO gophermart.users(login, password)
	VALUES (@login, @password)
`

func (r *UserRepository) CreateUser(ctx context.Context, user models.User) error {
	_, err := r.exec(ctx, createUserSQL, pgx.NamedArgs{
		"login":    user.Login,
		"password": user.Password,
	})

	return errors.Wrap(err, "exec")
}

const getUserSQL = `
	SELECT login, password
	FROM gophermart.users
	WHERE login = @login
`

func (r *UserRepository) GetUser(ctx context.Context, login string) (models.User, error) {
	args := pgx.NamedArgs{"login": login}
	fieldsPointer := func(u *models.User) []any { return u.ScanFields() }

	user, err := queryOne(ctx, r.baseRepo, getUserSQL, args, fieldsPointer)
	return user, errors.Wrap(err, "query one")
}
