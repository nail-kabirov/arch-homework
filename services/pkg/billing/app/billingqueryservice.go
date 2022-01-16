package app

func NewBillingQueryService(repoRead UserAccountRepositoryRead) BillingQueryService {
	return &billingQueryService{
		repoRead: repoRead,
	}
}

type BillingQueryService interface {
	AccountBalance(userID UserID) (Amount, error)
}

type billingQueryService struct {
	repoRead UserAccountRepositoryRead
}

func (s *billingQueryService) AccountBalance(userID UserID) (Amount, error) {
	account, err := s.repoRead.FindByID(userID)
	if err != nil {
		return nil, err
	}
	return account.Amount, nil
}
