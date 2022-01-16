package app

type RepositoryProvider interface {
	UserAccountRepository() UserAccountRepository
	ProcessedEventRepository() ProcessedEventRepository
	ProcessedRequestRepository() ProcessedRequestRepository
}

type TransactionalUnit interface {
	RepositoryProvider
	TransactionalUnitFactory
	Complete(err error) error
}

type TransactionalUnitFactory interface {
	NewTransactionalUnit() (TransactionalUnit, error)
}
