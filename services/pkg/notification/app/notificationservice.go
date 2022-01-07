package app

import (
	"arch-homework/pkg/common/app/uuid"

	"fmt"

	"github.com/pkg/errors"
)

func NewNotificationService(repo NotificationRepository) NotificationService {
	return &notificationService{
		repo: repo,
	}
}

type NotificationService interface {
	AddNotification(notificationType NotificationType, userID UserID, orderID uuid.UUID) error
}

type notificationService struct {
	repo NotificationRepository
}

func (n *notificationService) AddNotification(notificationType NotificationType, userID UserID, orderID uuid.UUID) error {
	msg, err := messageForOrder(notificationType, orderID)
	if err != nil {
		return errors.WithStack(err)
	}
	notification := Notification{
		Type:    notificationType,
		UserID:  userID,
		Message: msg,
	}
	return n.repo.Store(&notification)
}

func messageForOrder(notificationType NotificationType, orderID uuid.UUID) (string, error) {
	switch notificationType {
	case TypeOrderConfirmed:
		return fmt.Sprintf("Order %s confirmed", string(orderID)), nil
	case TypeOrderRejected:
		return fmt.Sprintf("Order %s rejected", string(orderID)), nil
	default:
		return "", errors.New("unknown notification type")
	}
}
