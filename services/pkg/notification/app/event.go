package app

import (
	"arch-homework/pkg/common/app/integrationevent"
	"arch-homework/pkg/common/app/uuid"
)

type ProcessedEventRepo interface {
	SetProcessed(uid integrationevent.EventUID) (alreadyProcessed bool, err error)
}

type UserEvent interface {
	UserID() UserID
}

func NewOrderConfirmedEvent(userID UserID, orderID uuid.UUID) UserEvent {
	return orderConfirmedEvent{userID: userID, orderID: orderID}
}

func NewOrderRejectedEvent(userID UserID, orderID uuid.UUID) UserEvent {
	return orderRejectedEvent{userID: userID, orderID: orderID}
}

type orderConfirmedEvent struct {
	userID  UserID
	orderID uuid.UUID
}

func (e orderConfirmedEvent) UserID() UserID {
	return e.userID
}

type orderRejectedEvent struct {
	userID  UserID
	orderID uuid.UUID
}

func (e orderRejectedEvent) UserID() UserID {
	return e.userID
}
