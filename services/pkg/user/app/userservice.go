package app

import (
	"github.com/pkg/errors"
)

var ErrAlreadyProcessed = errors.New("request with this id already processed")

func NewUserService(dbDependency DBDependency) *UserService {
	return &UserService{dbDependency: dbDependency}
}

type UserService struct {
	dbDependency DBDependency
}

func (s *UserService) Update(requestID RequestID, id UserID, firstName, lastName *string, email *Email, phone *Phone) error {
	return s.executeInTransaction(func(provider RepositoryProvider) error {
		eventRepo := provider.ProcessedRequestRepository()
		alreadyProcessed, err := eventRepo.SetRequestProcessed(requestID)
		if err != nil {
			return err
		}
		if alreadyProcessed {
			return ErrAlreadyProcessed
		}

		userRepo := provider.UserRepository()
		user, err := s.dbDependency.UserRepositoryRead().FindByID(id)
		if err != nil {
			if errors.Cause(err) == ErrUserNotFound {
				user = &User{UserID: id}
			} else {
				return err
			}
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
		return userRepo.Store(user)
	})
}

func (s *UserService) Remove(id UserID) error {
	return s.executeInTransaction(func(provider RepositoryProvider) error {
		return provider.UserRepository().Remove(id)
	})
}

func (s *UserService) Get(id UserID) (*User, error) {
	user, err := s.dbDependency.UserRepositoryRead().FindByID(id)
	if errors.Cause(err) == ErrUserNotFound {
		u := User{UserID: id}
		return &u, nil
	}
	return user, err
}

func (s *UserService) FindAll() ([]User, error) {
	return s.dbDependency.UserRepositoryRead().FindAll()
}

func (s *UserService) executeInTransaction(f func(RepositoryProvider) error) (err error) {
	var trUnit TransactionalUnit
	trUnit, err = s.dbDependency.NewTransactionalUnit()
	if err != nil {
		return err
	}
	defer func() {
		err = trUnit.Complete(err)
	}()
	err = f(trUnit)
	return err
}
