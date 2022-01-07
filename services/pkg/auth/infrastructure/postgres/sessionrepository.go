package postgres

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"arch-homework/pkg/auth/app"
	"arch-homework/pkg/common/infrastructure/postgres"
)

func NewSessionRepository(client postgres.Client) app.SessionRepository {
	return &sessionRepository{client: client}
}

type sessionRepository struct {
	client postgres.Client
}

func (repo *sessionRepository) Store(session *app.Session) error {
	const query = `
			INSERT INTO session (id, user_id, valid_till)
			VALUES (:id, :user_id, :valid_till)
			ON CONFLICT (id) DO UPDATE SET
				user_id = excluded.user_id,
				valid_till = excluded.valid_till
		`

	sessionx := sqlxSession{
		ID:        string(session.ID),
		UserID:    string(session.UserID),
		ValidTill: session.ValidTill,
	}

	_, err := repo.client.NamedExec(query, &sessionx)
	return errors.WithStack(err)
}

func (repo *sessionRepository) Remove(id app.SessionID) error {
	const query = `DELETE FROM session WHERE id = $1`
	_, err := repo.client.Exec(query, string(id))
	return errors.WithStack(err)
}

func (repo *sessionRepository) FindByID(id app.SessionID) (*app.Session, error) {
	const query = `SELECT id, user_id, valid_till FROM session WHERE id = $1`

	var session sqlxSession
	err := repo.client.Get(&session, query, string(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app.ErrSessionNotFound
		}
		return nil, errors.WithStack(err)
	}
	res := sqlxSessionToSession(&session)
	return &res, nil
}

func sqlxSessionToSession(session *sqlxSession) app.Session {
	return app.Session{
		ID:        app.SessionID(session.ID),
		UserID:    app.UserID(session.UserID),
		ValidTill: session.ValidTill,
	}
}

type sqlxSession struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	ValidTill time.Time `db:"valid_till"`
}
