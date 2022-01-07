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

type UserAccountRepository interface {
	Store(userAccount *UserAccount) error
	FindByID(id UserID) (*UserAccount, error)
}
