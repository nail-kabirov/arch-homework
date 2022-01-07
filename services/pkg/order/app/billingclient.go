package app

type BillingClient interface {
	ProcessOrderPayment(userID UserID, price Price) (succeeded bool, err error)
}
