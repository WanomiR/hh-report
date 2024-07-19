package modules

import tgcontroller "app/internal/modules/telegram/controller"

type Controllers struct {
	Tg tgcontroller.TgController
}

func NewControllers(tgController tgcontroller.TgController) *Controllers {
	return &Controllers{
		Tg: tgController,
	}
}
