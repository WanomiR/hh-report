package service

import (
	"app/internal/modules/telegram/entities"
	"context"
)

type TgEventsServicer interface {
	GetUpdates(ctx context.Context, limit int) ([]entities.Event, error)
	ProcessMessage(ctx context.Context, event entities.Event) error
}
