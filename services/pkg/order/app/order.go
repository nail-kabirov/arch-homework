package app

import (
	"arch-homework/pkg/common/app/uuid"
	"errors"
	"time"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderID uuid.UUID
type UserID uuid.UUID

type Order struct {
	ID           OrderID
	UserID       UserID
	Price        Price
	CreationDate time.Time
}

type OrderRepositoryRead interface {
	FindByID(id OrderID) (*Order, error)
	FindAllByUserID(userID UserID) ([]Order, error)
}

type OrderRepository interface {
	OrderRepositoryRead
	Store(order *Order) error
}
