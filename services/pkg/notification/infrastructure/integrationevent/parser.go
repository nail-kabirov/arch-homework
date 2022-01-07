package integrationevent

import (
	"arch-homework/pkg/common/app/integrationevent"
	"arch-homework/pkg/common/app/uuid"
	"arch-homework/pkg/notification/app"

	"encoding/json"

	"github.com/pkg/errors"
)

const typeOrderConfirmed = "order.order_confirmed"
const typeOrderRejected = "order.order_rejected"

func NewEventParser() app.IntegrationEventParser {
	return eventParser{}
}

type eventParser struct {
}

func (e eventParser) ParseIntegrationEvent(event integrationevent.EventData) (app.UserEvent, error) {
	switch event.Type {
	case typeOrderConfirmed:
		return parseOrderConfirmedEvent(event.Body)
	case typeOrderRejected:
		return parseOrderRejectedEvent(event.Body)
	default:
		return nil, nil
	}
}

func parseOrderConfirmedEvent(strBody string) (app.UserEvent, error) {
	body, err := parseOrderEvent(strBody)
	if err != nil {
		return nil, err
	}
	return app.NewOrderConfirmedEvent(app.UserID(body.UserID), uuid.UUID(body.OrderID)), nil
}

func parseOrderRejectedEvent(strBody string) (app.UserEvent, error) {
	body, err := parseOrderEvent(strBody)
	if err != nil {
		return nil, err
	}
	return app.NewOrderRejectedEvent(app.UserID(body.UserID), uuid.UUID(body.OrderID)), nil
}

func parseOrderEvent(strBody string) (orderEventBody, error) {
	var body orderEventBody
	err := json.Unmarshal([]byte(strBody), &body)
	if err != nil {
		return body, errors.WithStack(err)
	}
	err = uuid.ValidateUUID(body.UserID)
	if err != nil {
		return body, errors.WithStack(err)
	}
	err = uuid.ValidateUUID(body.OrderID)
	return body, errors.WithStack(err)
}

type orderEventBody struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}
