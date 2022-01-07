package app

import "github.com/pkg/errors"

var ErrNotEnoughFunds = errors.New("not enough funds for payment")

func NewBillingService(repo UserAccountRepository) BillingService {
	return &billingService{
		repo: repo,
	}
}

type BillingService interface {
	CreateAccount(userID UserID) error
	AccountBalance(userID UserID) (Amount, error)
	TopUpAccount(userID UserID, amount Amount) error
	ProcessPayment(userID UserID, amount Amount) error
}

type billingService struct {
	repo UserAccountRepository
}

func (s *billingService) CreateAccount(userID UserID) error {
	_, err := s.repo.FindByID(userID)
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
	return s.repo.Store(&userAccount)
}

func (s *billingService) AccountBalance(userID UserID) (Amount, error) {
	account, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	return account.Amount, nil
}

func (s *billingService) TopUpAccount(userID UserID, amount Amount) error {
	account, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	account.Amount = AmountFromRawValue(account.Amount.RawValue() + amount.RawValue())
	return s.repo.Store(account)
}

func (s *billingService) ProcessPayment(userID UserID, amount Amount) error {
	account, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if amount.RawValue() > account.Amount.RawValue() {
		return ErrNotEnoughFunds
	}
	account.Amount = AmountFromRawValue(account.Amount.RawValue() - amount.RawValue())
	return s.repo.Store(account)
}
