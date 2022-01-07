package app

type RepositoryProvider interface {
	NotificationRepository() NotificationRepository
	ProcessedEventRepo() ProcessedEventRepo
}

type TransactionalUnit interface {
	RepositoryProvider
	Complete(err error) error
}

type TransactionalUnitFactory interface {
	NewTransactionalUnit() (TransactionalUnit, error)
}
