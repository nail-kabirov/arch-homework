package app

import (
	"arch-homework/pkg/common/app/storedevent"
	"arch-homework/pkg/common/app/uuid"

	"github.com/pkg/errors"
)

const maxLoginLen = 255

func NewUserService(dbDependency DBDependency, eventSender storedevent.Sender, passwordEncoder PasswordEncoder) *UserService {
	return &UserService{
		readRepo:        dbDependency.UserRepositoryRead(),
		trUnitFactory:   dbDependency,
		eventSender:     eventSender,
		passwordEncoder: passwordEncoder,
	}
}

type UserService struct {
	readRepo        UserRepositoryRead
	trUnitFactory   TransactionalUnitFactory
	eventSender     storedevent.Sender
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

	err := s.executeInTransaction(func(provider RepositoryProvider) error {
		err2 := provider.UserRepository().Store(&user)
		if err2 != nil {
			return err2
		}

		event := NewUserRegisteredEvent(id, login)
		err2 = provider.EventStore().Add(event)
		if err2 != nil {
			return err2
		}
		s.eventSender.EventStored(event.UID)

		return nil
	})
	if err != nil {
		return "", err
	}

	s.eventSender.SendStoredEvents()
	return id, err
}

func (s *UserService) Remove(id UserID) error {
	return s.executeInTransaction(func(provider RepositoryProvider) error {
		return provider.UserRepository().Remove(id)
	})
}

func (s *UserService) FindUserByID(id UserID) (*User, error) {
	return s.readRepo.FindByID(id)
}

func (s *UserService) FindUserByLoginAndPassword(login, password string) (*User, error) {
	user, err := s.readRepo.FindByLogin(Login(login))
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
	if user, err := s.readRepo.FindByLogin(login); err != ErrUserNotFound || user != nil {
		if user != nil {
			return ErrLoginAlreadyExists
		}
		return err
	}
	return nil
}

func (s *UserService) executeInTransaction(f func(RepositoryProvider) error) (err error) {
	var trUnit TransactionalUnit
	trUnit, err = s.trUnitFactory.NewTransactionalUnit()
	if err != nil {
		return err
	}
	defer func() {
		err = trUnit.Complete(err)
	}()
	err = f(trUnit)
	return err
}
