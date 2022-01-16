package app

import (
	"errors"

	"arch-homework/pkg/common/app/uuid"
)

var ErrUserAccountAlreadyExists = errors.New("user account already exists")
var ErrUserAccountNotFound = errors.New("user account not found")

type UserID uuid.UUID

type UserAccount struct {
	UserID UserID
	Amount Amount
}

type UserAccountRepositoryRead interface {
	FindByID(id UserID) (*UserAccount, error)
}

type UserAccountRepository interface {
	UserAccountRepositoryRead
	Store(userAccount *UserAccount) error
}
