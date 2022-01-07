package app

import (
	"arch-homework/pkg/common/app/uuid"
	"time"
)

type UserID uuid.UUID

type NotificationType int

const (
	TypeOrderConfirmed NotificationType = 1
	TypeOrderRejected  NotificationType = 2
)

type Notification struct {
	Type         NotificationType
	UserID       UserID
	Message      string
	CreationDate time.Time
}

type NotificationRepository interface {
	Store(notification *Notification) error
	FindAllByUserID(userID UserID) ([]Notification, error)
}
