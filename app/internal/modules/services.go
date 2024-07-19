package modules

import (
	tgservice "app/internal/modules/telegram/service"
)

type Services struct {
	Tg tgservice.TgServicer
}

func NewServices(tgService tgservice.TgServicer) *Services {
	return &Services{
		Tg: tgService,
	}
}
