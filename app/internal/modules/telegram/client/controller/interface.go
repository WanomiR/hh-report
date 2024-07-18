package controller

import (
	"app/internal/modules/telegram/entities"
	"context"
)

type TgClientController interface {
	GetUpdates(ctx context.Context, offset, limit, timeout int) ([]entities.Update, error)
	SendMessage(ctx context.Context, chatID int, text string) error
}
