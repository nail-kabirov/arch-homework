package app

import (
	"arch-homework1/pkg/user/common/uuid"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUsernameAlreadyExists = errors.New("username already exists")
var ErrUsernameTooLong = errors.New("username too long")

type UserID uuid.UUID
type Username string
type Email string
type Phone string

type User struct {
	UserID    UserID
	Username  Username
	FirstName string
	LastName  string
	Email     Email
	Phone     Phone
}

type UserRepository interface {
	Store(user *User) error
	Remove(id UserID) error
	FindByID(id UserID) (*User, error)
	FindByUserName(userName Username) (*User, error)
	FindAll() ([]User, error)
}
