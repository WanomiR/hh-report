package controller

import (
	"app/internal/lib/e"
	"app/internal/modules/telegram/entities"
	"app/internal/modules/telegram/events/service"
	"context"
)

type TgEventsControl struct {
	eventsService service.TgEventsServicer
}

func NewTgEventsControl(es service.TgEventsServicer) *TgEventsControl {
	tec := &TgEventsControl{eventsService: es}
	return tec
}

func (c *TgEventsControl) Fetch(ctx context.Context, limit int) ([]entities.Event, error) {
	return c.eventsService.GetUpdates(ctx, limit)
}

func (c *TgEventsControl) Process(ctx context.Context, event entities.Event) error {
	switch event.Type {
	case entities.Message:
		return c.eventsService.ProcessMessage(ctx, event)
	default:
		return e.WrapIfErr("cannot process message", service.ErrUnknownEventType)
	}
}
