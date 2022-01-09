package app

import (
	"arch-homework/pkg/common/app/integrationevent"
	"arch-homework/pkg/common/app/storedevent"
	"arch-homework/pkg/common/app/uuid"

	"github.com/pkg/errors"

	"time"
)

var ErrPaymentFailed = errors.New("order payment failed")

func NewOrderService(dbDependency DBDependency, eventSender storedevent.Sender, billingClient BillingClient) *OrderService {
	return &OrderService{
		readRepo:      dbDependency.OrderRepositoryRead(),
		trUnitFactory: dbDependency,
		eventSender:   eventSender,
		billingClient: billingClient,
	}
}

type OrderService struct {
	readRepo      OrderRepositoryRead
	trUnitFactory TransactionalUnitFactory
	eventSender   storedevent.Sender
	billingClient BillingClient
}

func (s *OrderService) Create(userID UserID, priceFloat float64) (OrderID, error) {
	price, err := PriceFromFloat(priceFloat)
	if err != nil {
		return "", err
	}
	id := OrderID(uuid.GenerateNew())

	order := Order{
		ID:           id,
		UserID:       userID,
		Price:        price,
		CreationDate: time.Now(),
	}

	paymentSucceeded, err := s.billingClient.ProcessOrderPayment(userID, price)
	if err != nil {
		return "", err
	}

	err = s.executeInTransaction(func(provider RepositoryProvider) error {
		var event integrationevent.EventData
		if paymentSucceeded {
			err2 := provider.OrderRepository().Store(&order)
			if err2 != nil {
				return err2
			}
			event = NewOrderConfirmedEvent(id, userID)
		} else {
			event = NewOrderRejectedEvent(id, userID)
		}

		err2 := provider.EventStore().Add(event)
		if err2 != nil {
			return err2
		}
		s.eventSender.EventStored(event.UID)

		return nil
	})
	if err != nil {
		return "", err
	}

	s.eventSender.SendStoredEvents()

	if !paymentSucceeded {
		return id, errors.WithStack(ErrPaymentFailed)
	}

	return id, err
}

func (s *OrderService) Get(userID UserID, id OrderID) (*Order, error) {
	order, err := s.readRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s *OrderService) FindAll(userID UserID) ([]Order, error) {
	return s.readRepo.FindAllByUserID(userID)
}

func (s *OrderService) executeInTransaction(f func(RepositoryProvider) error) (err error) {
	var trUnit TransactionalUnit
	trUnit, err = s.trUnitFactory.NewTransactionalUnit()
	if err != nil {
		return err
	}
	defer func() {
		err = trUnit.Complete(err)
	}()
	err = f(trUnit)
	return err
}
