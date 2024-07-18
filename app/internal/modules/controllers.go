package modules

import (
	tgclcontroller "app/internal/modules/telegram/client/controller"
	tgevcontroller "app/internal/modules/telegram/events/controller"
)

type Controllers struct {
	TgClient tgclcontroller.TgClientController
	TgEvents tgevcontroller.TgEventsController
}

func NewControllers(tgClient tgclcontroller.TgClientController, tgEvents tgevcontroller.TgEventsController) *Controllers {
	ctrl := &Controllers{
		TgClient: tgClient,
		TgEvents: tgEvents,
	}
	return ctrl
}
