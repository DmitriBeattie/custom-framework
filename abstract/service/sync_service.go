package service

import (
	"fmt"
	"delivery-report/framework/interfaces/app"
)

type SyncService struct {
	repoDef map[string]app.EventRepository
	consDef map[string]app.EventConsumer
	log app.Logger
}

func NewSyncService(
	_repoDef map[string]app.EventRepository,
	_consDef map[string]app.EventConsumer,
	_log app.Logger,
) *SyncService {
	return &SyncService{
		repoDef: _repoDef,
		consDef: _consDef,
		log:     _log,
	}
}

func (srv *SyncService) Exec(eventName string, adapter app.EventAdapter, needLogNotProcessedEvents bool) error {
	repo, isRepoExists := srv.repoDef[eventName]
	if !isRepoExists {
		return fmt.Errorf("Not found event repo with name: %s", eventName)
	}

	cons, isConsumerExists := srv.consDef[eventName]
	if !isConsumerExists {
		return fmt.Errorf("Not found consumer with name: %s", eventName)
	}

	return app.NewEventProcessor(repo, cons, eventName, adapter, srv.log).Process(needLogNotProcessedEvents)
}
