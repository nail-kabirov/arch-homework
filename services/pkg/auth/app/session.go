package app

import (
	"arch-homework/pkg/common/app/uuid"
	"errors"
	"time"
)

var ErrSessionNotFound = errors.New("user not found")

type SessionID uuid.UUID

type Session struct {
	ID        SessionID
	UserID    UserID
	ValidTill time.Time
}

type SessionRepository interface {
	Store(session *Session) error
	Remove(id SessionID) error
	FindByID(id SessionID) (*Session, error)
}
