package app

import "arch-homework/pkg/common/app/storedevent"

type RepositoryProvider interface {
	UserRepository() UserRepository
	EventStore() storedevent.EventStore
}

type ReadRepositoryProvider interface {
	UserRepositoryRead() UserRepositoryRead
}

type TransactionalUnit interface {
	RepositoryProvider
	Complete(err error) error
}

type TransactionalUnitFactory interface {
	NewTransactionalUnit() (TransactionalUnit, error)
}

type DBDependency interface {
	TransactionalUnitFactory
	ReadRepositoryProvider
}
