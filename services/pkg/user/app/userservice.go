package app

import (
	"github.com/pkg/errors"
)

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

type UserService struct {
	repo UserRepository
}

func (s *UserService) Update(id UserID, firstName, lastName *string, email *Email, phone *Phone) error {
	user, err := s.Get(id)
	if err != nil {
		return err
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

func (s *UserService) Get(id UserID) (*User, error) {
	user, err := s.repo.FindByID(id)
	if errors.Cause(err) == ErrUserNotFound {
		u := User{UserID: id}
		return &u, nil
	}
	return user, err
}

func (s *UserService) FindAll() ([]User, error) {
	return s.repo.FindAll()
}
