package app

import (
	"arch-homework/pkg/common/app/integrationevent"
)

type IntegrationEventParser interface {
	ParseIntegrationEvent(event integrationevent.EventData) (UserEvent, error)
}

func NewEventHandler(trUnitFactory TransactionalUnitFactory, parser IntegrationEventParser) integrationevent.EventHandler {
	return &eventHandler{
		trUnitFactory: trUnitFactory,
		parser:        parser,
	}
}

type eventHandler struct {
	trUnitFactory TransactionalUnitFactory
	parser        IntegrationEventParser
}

func (handler *eventHandler) Handle(event integrationevent.EventData) error {
	parsedEvent, err := handler.parser.ParseIntegrationEvent(event)
	if err != nil || parsedEvent == nil {
		return err
	}

	return handler.executeInTransaction(func(trUnit TransactionalUnit) error {
		eventRepo := trUnit.ProcessedEventRepository()
		alreadyProcessed, err := eventRepo.SetEventProcessed(event.UID)
		if err != nil {
			return err
		}
		if alreadyProcessed {
			return nil
		}

		switch e := parsedEvent.(type) {
		case userRegisteredEvent:
			return handleUserRegisteredEvent(trUnit, e)
		default:
			return nil
		}
	})
}

func (handler *eventHandler) executeInTransaction(f func(TransactionalUnit) error) (err error) {
	var trUnit TransactionalUnit
	trUnit, err = handler.trUnitFactory.NewTransactionalUnit()
	if err != nil {
		return err
	}
	defer func() {
		err = trUnit.Complete(err)
	}()
	err = f(trUnit)
	return err
}

func handleUserRegisteredEvent(trUnit TransactionalUnit, e userRegisteredEvent) error {
	service := NewBillingService(trUnit)
	return service.CreateAccount(e.UserID())
}
