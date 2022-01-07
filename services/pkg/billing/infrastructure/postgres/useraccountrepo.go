package postgres

import (
	"arch-homework/pkg/billing/app"
	"arch-homework/pkg/common/infrastructure/postgres"

	"database/sql"

	"github.com/pkg/errors"
)

func NewUserAccountRepository(client postgres.Client) app.UserAccountRepository {
	return &userAccountRepository{client: client}
}

type userAccountRepository struct {
	client postgres.Client
}

func (repo *userAccountRepository) Store(userAccount *app.UserAccount) error {
	const query = `
			INSERT INTO user_account (user_id, amount)
			VALUES (:user_id, :amount)
			ON CONFLICT (user_id) DO UPDATE SET amount = excluded.amount
		`

	userAccountx := sqlxUserAccount{
		UserID: string(userAccount.UserID),
		Amount: userAccount.Amount.RawValue(),
	}

	_, err := repo.client.NamedExec(query, &userAccountx)
	return errors.WithStack(err)
}

func (repo *userAccountRepository) FindByID(id app.UserID) (*app.UserAccount, error) {
	const query = `SELECT user_id, amount FROM user_account WHERE user_id = $1`

	var user sqlxUserAccount
	err := repo.client.Get(&user, query, string(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app.ErrUserAccountNotFound
		}
		return nil, errors.WithStack(err)
	}
	res := app.UserAccount{
		UserID: app.UserID(user.UserID),
		Amount: app.AmountFromRawValue(user.Amount),
	}
	return &res, nil
}

type sqlxUserAccount struct {
	UserID string `db:"user_id"`
	Amount uint64 `db:"amount"`
}
