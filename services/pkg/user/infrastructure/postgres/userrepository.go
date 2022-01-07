package postgres

import (
	"arch-homework/pkg/common/infrastructure/postgres"
	"arch-homework/pkg/user/app"

	"database/sql"

	"github.com/pkg/errors"
)

func NewUserRepository(client postgres.Client) app.UserRepository {
	return &userRepository{client: client}
}

type userRepository struct {
	client postgres.Client
}

func (repo *userRepository) Store(user *app.User) error {
	const query = `
			INSERT INTO users (id, first_name, last_name, email, phone)
			VALUES (:id, :first_name, :last_name, :email, :phone)
			ON CONFLICT (id) DO UPDATE SET
				first_name = excluded.first_name,
				last_name = excluded.last_name,
				email = excluded.email,
				phone = excluded.phone;
		`

	userx := sqlxUser{
		ID:        string(user.UserID),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     string(user.Email),
		Phone:     string(user.Phone),
	}

	_, err := repo.client.NamedExec(query, &userx)
	return errors.WithStack(err)
}

func (repo *userRepository) Remove(id app.UserID) error {
	const query = `DELETE FROM users WHERE id = $1`
	_, err := repo.client.Exec(query, string(id))
	return errors.WithStack(err)
}

func (repo *userRepository) FindByID(id app.UserID) (*app.User, error) {
	const query = `SELECT id, first_name, last_name, email, phone FROM users WHERE id = $1`

	var user sqlxUser
	err := repo.client.Get(&user, query, string(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app.ErrUserNotFound
		}
		return nil, errors.WithStack(err)
	}
	res := sqlxUserToUser(&user)
	return &res, nil
}

func (repo *userRepository) FindAll() ([]app.User, error) {
	const query = `SELECT id, first_name, last_name, email, phone FROM users`

	var users []*sqlxUser
	err := repo.client.Select(&users, query)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res := make([]app.User, 0, len(users))
	for _, user := range users {
		res = append(res, sqlxUserToUser(user))
	}
	return res, nil
}

func sqlxUserToUser(user *sqlxUser) app.User {
	return app.User{
		UserID:    app.UserID(user.ID),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     app.Email(user.Email),
		Phone:     app.Phone(user.Phone),
	}
}

type sqlxUser struct {
	ID        string `db:"id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
	Phone     string `db:"phone"`
}
