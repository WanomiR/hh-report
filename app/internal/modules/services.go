package modules

import (
	tgclvservice "app/internal/modules/telegram/client/service"
	tgevservice "app/internal/modules/telegram/events/service"
)

type Services struct {
	TgClient tgclvservice.TgClientServicer
	TgEvents tgevservice.TgEventsServicer
}

func NewServices(tgClient tgclvservice.TgClientServicer, tgEvents tgevservice.TgEventsServicer) *Services {
	srv := &Services{
		TgClient: tgClient,
		TgEvents: tgEvents,
	}
	return srv
}
