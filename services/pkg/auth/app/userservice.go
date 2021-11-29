package app

import (
	"github.com/pkg/errors"

	"arch-homework5/pkg/common/uuid"
)

const maxLoginLen = 255

func NewUserService(userRepo UserRepository, passwordEncoder PasswordEncoder) *UserService {
	return &UserService{repo: userRepo, passwordEncoder: passwordEncoder}
}

type UserService struct {
	repo            UserRepository
	passwordEncoder PasswordEncoder
}

func (s *UserService) Add(login, password string) (UserID, error) {
	if err := s.checkLogin(Login(login)); err != nil {
		return "", err
	}

	id := UserID(uuid.GenerateNew())
	user := User{
		UserID:   id,
		Login:    Login(login),
		Password: s.passwordEncoder.Encode(password, id),
	}

	return id, s.repo.Store(&user)
}

func (s *UserService) Remove(id UserID) error {
	return s.repo.Remove(id)
}

func (s *UserService) FindUserByID(id UserID) (*User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) FindUserByLoginAndPassword(login, password string) (*User, error) {
	user, err := s.repo.FindByLogin(Login(login))
	if err != nil {
		return nil, err
	}
	encodedPass := s.passwordEncoder.Encode(password, user.UserID)
	if encodedPass != user.Password {
		return nil, ErrInvalidPassword
	}
	return user, nil
}

func (s *UserService) checkLogin(login Login) error {
	if len(login) > maxLoginLen {
		return errors.Wrapf(ErrLoginTooLong, "max login length (%d symbols) exceeded", maxLoginLen)
	}
	if user, err := s.repo.FindByLogin(login); err != ErrUserNotFound || user != nil {
		if user != nil {
			return ErrLoginAlreadyExists
		}
		return err
	}
	return nil
}
