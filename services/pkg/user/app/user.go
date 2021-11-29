package app

import (
	"arch-homework5/pkg/common/uuid"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

type UserID uuid.UUID
type Email string
type Phone string

type User struct {
	UserID    UserID
	FirstName string
	LastName  string
	Email     Email
	Phone     Phone
}

type UserRepository interface {
	Store(user *User) error
	Remove(id UserID) error
	FindByID(id UserID) (*User, error)
	FindAll() ([]User, error)
}
