package controller

import (
	"app/internal/modules/telegram/entities"
	"context"
)

type TgEventsController interface {
	Fetch(ctx context.Context, limit int) ([]entities.Event, error)
	Process(ctx context.Context, event entities.Event) error
}
