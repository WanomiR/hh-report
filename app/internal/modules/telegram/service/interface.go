package service

import (
	"app/internal/modules/telegram/entities"
	"context"
)

type TgServicer interface {
	GetUpdates(ctx context.Context) ([]entities.Update, error)
	ProcessUpdates(ctx context.Context, updates []entities.Update)
}
