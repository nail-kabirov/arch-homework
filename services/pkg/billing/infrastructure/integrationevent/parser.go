package integrationevent

import (
	"arch-homework/pkg/billing/app"
	"arch-homework/pkg/common/app/integrationevent"
	"arch-homework/pkg/common/app/uuid"

	"encoding/json"

	"github.com/pkg/errors"
)

const typeUserRegistered = "auth.user_registered"

func NewEventParser() app.IntegrationEventParser {
	return eventParser{}
}

type eventParser struct {
}

func (e eventParser) ParseIntegrationEvent(event integrationevent.EventData) (app.UserEvent, error) {
	switch event.Type {
	case typeUserRegistered:
		return parseUserRegisteredEvent(event.Body)
	default:
		return nil, nil
	}
}

func parseUserRegisteredEvent(strBody string) (app.UserEvent, error) {
	var body userRegisteredEventBody
	err := json.Unmarshal([]byte(strBody), &body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = uuid.ValidateUUID(body.UserID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return app.NewUserRegisteredEvent(app.UserID(body.UserID), body.Login), nil
}

type userRegisteredEventBody struct {
	UserID string `json:"user_id"`
	Login  string `json:"login"`
}
