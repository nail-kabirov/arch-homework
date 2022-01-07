package app

type RepositoryProvider interface {
	UserAccountRepository() UserAccountRepository
	ProcessedEventRepo() ProcessedEventRepo
}

type TransactionalUnit interface {
	RepositoryProvider
	Complete(err error) error
}

type TransactionalUnitFactory interface {
	NewTransactionalUnit() (TransactionalUnit, error)
}
