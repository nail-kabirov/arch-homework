package app

import (
	"arch-homework5/pkg/common/uuid"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")
var ErrLoginTooLong = errors.New("login too long")
var ErrLoginAlreadyExists = errors.New("login already exists")
var ErrInvalidPassword = errors.New("invalid password")

type UserID uuid.UUID
type Login string
type Password string

type User struct {
	UserID   UserID
	Login    Login
	Password Password
}

type UserRepository interface {
	Store(user *User) error
	Remove(id UserID) error
	FindByID(id UserID) (*User, error)
	FindByLogin(login Login) (*User, error)
}
