package app

import (
	"arch-homework/pkg/common/app/uuid"
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

type UserRepositoryRead interface {
	FindByID(id UserID) (*User, error)
	FindAll() ([]User, error)
}

type UserRepository interface {
	UserRepositoryRead
	Store(user *User) error
	Remove(id UserID) error
}
