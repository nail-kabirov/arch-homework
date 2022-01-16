package app

import "arch-homework/pkg/common/app/integrationevent"

type ProcessedEventRepository interface {
	SetEventProcessed(uid integrationevent.EventUID) (alreadyProcessed bool, err error)
}

type UserEvent interface {
	UserID() UserID
}

func NewUserRegisteredEvent(userID UserID, login string) UserEvent {
	return userRegisteredEvent{userID: userID, login: login}
}

type userRegisteredEvent struct {
	userID UserID
	login  string
}

func (e userRegisteredEvent) UserID() UserID {
	return e.userID
}

func (e userRegisteredEvent) Login() string {
	return e.login
}
