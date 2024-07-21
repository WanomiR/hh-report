package controller

import (
	"app/internal/modules/telegram/service"
	"context"
	"log"
	"time"
)

type TgController interface {
	Serve(ctx context.Context)
}

type TgControl struct {
	service service.TgServicer
}

func NewTgControl(service service.TgServicer) *TgControl {
	return &TgControl{service: service}
}

func (c *TgControl) Serve(ctx context.Context) {

	for {
		// get updates every second
		time.Sleep(1 * time.Second)

		updates, err := c.service.GetUpdates(ctx)
		if err != nil {
			log.Println(err.Error()) // skip if an error
			continue
		}

		if len(updates) == 0 { // skip if no updates
			continue
		}

		c.service.ProcessUpdates(ctx, updates)
	}
}
