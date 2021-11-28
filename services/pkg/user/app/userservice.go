package app

import (
	"arch-homework5/pkg/common/uuid"

	"github.com/pkg/errors"
)

const maxUsernameLen = 255

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

type UserService struct {
	repo UserRepository
}

func (s *UserService) Add(username Username, firstName, lastName string, email Email, phone Phone) (UserID, error) {
	if err := s.checkUsername(username); err != nil {
		return "", err
	}

	id := UserID(uuid.GenerateNew())
	user := User{
		UserID:    id,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     phone,
	}

	return id, s.repo.Store(&user)
}

func (s *UserService) Update(id UserID, username *Username, firstName, lastName *string, email *Email, phone *Phone) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if username != nil && *username != user.Username {
		if err := s.checkUsername(*username); err != nil {
			return err
		}
		user.Username = *username
	}
	if firstName != nil {
		user.FirstName = *firstName
	}
	if lastName != nil {
		user.LastName = *lastName
	}
	if email != nil {
		user.Email = *email
	}
	if phone != nil {
		user.Phone = *phone
	}
	return s.repo.Store(user)
}

func (s *UserService) Remove(id UserID) error {
	return s.repo.Remove(id)
}

func (s *UserService) Find(id UserID) (*User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) FindAll() ([]User, error) {
	return s.repo.FindAll()
}

func (s *UserService) checkUsername(userName Username) error {
	if len(userName) > maxUsernameLen {
		return errors.Wrapf(ErrUsernameTooLong, "max username length (%d symbols) exceeded", maxUsernameLen)
	}
	if user, err := s.repo.FindByUserName(userName); err != ErrUserNotFound || user != nil {
		if user != nil {
			return ErrUsernameAlreadyExists
		}
		return err
	}
	return nil
}
