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

	return handler.executeInTransaction(func(provider RepositoryProvider) error {
		eventRepo := provider.ProcessedEventRepo()
		alreadyProcessed, err := eventRepo.SetProcessed(event.UID)
		if err != nil {
			return err
		}
		if alreadyProcessed {
			return nil
		}

		switch e := parsedEvent.(type) {
		case userRegisteredEvent:
			return handleUserRegisteredEvent(provider, e)
		default:
			return nil
		}
	})
}

func (handler *eventHandler) executeInTransaction(f func(RepositoryProvider) error) (err error) {
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

func handleUserRegisteredEvent(provider RepositoryProvider, e userRegisteredEvent) error {
	service := NewBillingService(provider.UserAccountRepository())
	return service.CreateAccount(e.UserID())
}
