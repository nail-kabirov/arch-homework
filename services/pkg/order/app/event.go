package app

import (
	"arch-homework/pkg/common/app/integrationevent"
	"arch-homework/pkg/common/app/uuid"
	"encoding/json"
)

const typeOrderConfirmed = "order.order_confirmed"
const typeOrderRejected = "order.order_rejected"

func NewOrderConfirmedEvent(orderID OrderID, userID UserID) integrationevent.EventData {
	body, _ := json.Marshal(orderEventBody{
		OrderID: string(orderID),
		UserID:  string(userID),
	})

	return integrationevent.EventData{
		UID:  newUID(),
		Type: typeOrderConfirmed,
		Body: string(body),
	}
}

func NewOrderRejectedEvent(orderID OrderID, userID UserID) integrationevent.EventData {
	body, _ := json.Marshal(orderEventBody{
		OrderID: string(orderID),
		UserID:  string(userID),
	})

	return integrationevent.EventData{
		UID:  newUID(),
		Type: typeOrderRejected,
		Body: string(body),
	}
}

func newUID() integrationevent.EventUID {
	return integrationevent.EventUID(uuid.GenerateNew())
}

type orderEventBody struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}
