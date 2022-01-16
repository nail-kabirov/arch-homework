package app

import "github.com/pkg/errors"

var ErrNotEnoughFunds = errors.New("not enough funds for payment")
var ErrAlreadyProcessed = errors.New("request with this id already processed")

func NewBillingService(trUnitFactory TransactionalUnitFactory) BillingService {
	return &billingService{trUnitFactory: trUnitFactory}
}

type BillingService interface {
	CreateAccount(userID UserID) error
	TopUpAccount(requestID RequestID, userID UserID, amount Amount) error
	ProcessPayment(userID UserID, amount Amount) error
}

type billingService struct {
	trUnitFactory TransactionalUnitFactory
}

func (s *billingService) CreateAccount(userID UserID) error {
	return s.executeInTransaction(func(repoProvider RepositoryProvider) error {
		repo := repoProvider.UserAccountRepository()
		_, err := repo.FindByID(userID)
		if err == nil {
			return ErrUserAccountAlreadyExists
		}
		if errors.Cause(err) != ErrUserAccountNotFound {
			return err
		}
		userAccount := UserAccount{
			UserID: userID,
			Amount: AmountFromRawValue(0),
		}
		return repo.Store(&userAccount)
	})
}

func (s *billingService) TopUpAccount(requestID RequestID, userID UserID, amount Amount) error {
	return s.executeInTransaction(func(provider RepositoryProvider) error {
		eventRepo := provider.ProcessedRequestRepository()
		alreadyProcessed, err := eventRepo.SetRequestProcessed(requestID)
		if err != nil {
			return err
		}
		if alreadyProcessed {
			return ErrAlreadyProcessed
		}

		accountRepo := provider.UserAccountRepository()
		account, err := accountRepo.FindByID(userID)
		if err != nil {
			return err
		}
		account.Amount = AmountFromRawValue(account.Amount.RawValue() + amount.RawValue())
		return accountRepo.Store(account)
	})
}

func (s *billingService) ProcessPayment(userID UserID, amount Amount) error {
	return s.executeInTransaction(func(provider RepositoryProvider) error {
		repo := provider.UserAccountRepository()
		account, err := repo.FindByID(userID)
		if err != nil {
			return err
		}
		if amount.RawValue() > account.Amount.RawValue() {
			return ErrNotEnoughFunds
		}
		account.Amount = AmountFromRawValue(account.Amount.RawValue() - amount.RawValue())
		return repo.Store(account)
	})
}

func (s *billingService) executeInTransaction(f func(RepositoryProvider) error) (err error) {
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
