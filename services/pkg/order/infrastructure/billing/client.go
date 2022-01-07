package billing

import (
	"arch-homework/pkg/common/infrastructure/httpclient"
	"arch-homework/pkg/order/app"

	"github.com/pkg/errors"

	"net/http"
)

const processPaymentURL = "/internal/api/v1/payment"

func NewClient(client http.Client, serviceHost string) app.BillingClient {
	return &billingClient{httpClient: httpclient.NewClient(client, serviceHost)}
}

type billingClient struct {
	httpClient httpclient.Client
}

func (c *billingClient) ProcessOrderPayment(userID app.UserID, price app.Price) (succeeded bool, err error) {
	request := processPaymentRequest{
		UserID: string(userID),
		Amount: price.Value(),
	}
	err = c.httpClient.MakeJSONRequest(request, nil, http.MethodPost, processPaymentURL)
	if err == nil {
		return true, nil
	}
	if e, ok := errors.Cause(err).(*httpclient.HTTPError); ok {
		if e.StatusCode == http.StatusBadRequest {
			return false, nil
		}
	}

	return false, err
}

type processPaymentRequest struct {
	UserID string  `json:"userID"`
	Amount float64 `json:"amount"`
}
