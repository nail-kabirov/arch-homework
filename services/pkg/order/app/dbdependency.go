package app

import "arch-homework/pkg/common/app/storedevent"

type RepositoryProvider interface {
	OrderRepository() OrderRepository
	EventStore() storedevent.EventStore
}

type ReadRepositoryProvider interface {
	OrderRepositoryRead() OrderRepositoryRead
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
