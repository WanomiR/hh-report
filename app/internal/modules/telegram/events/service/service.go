package service

import (
	"app/internal/lib/e"
	tgcontroller "app/internal/modules/telegram/client/controller"
	"app/internal/modules/telegram/entities"
	"context"
	"errors"
	"strings"
)

const MsgHelp = `I can save and keep you pages. Also I can offer you them to read.

In order to save the page, just send me al link to it.

In order to get a random page from your list, send me command /rnd.
Caution! After that, this page will be removed from your list!`

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
	CmdHelp             = "/help"
	CmdStart            = "/start"
	MsgHello            = "Hi there! 👾\n\n" + MsgHelp
	MsgUnknownCommand   = "Unknown command 🤔"
)

type TgEventsService struct {
	offset   int
	timeout  int
	tgClient tgcontroller.TgClientController
}

func NewTgEventsService(tgClientController tgcontroller.TgClientController) *TgEventsService {
	tes := &TgEventsService{
		tgClient: tgClientController,
		timeout:  0,
	}
	return tes
}

func (s *TgEventsService) GetUpdates(ctx context.Context, limit int) ([]entities.Event, error) {
	updates, err := s.tgClient.GetUpdates(ctx, s.offset, limit, s.timeout)
	if err != nil {
		return nil, e.WrapIfErr("couldn't get updates", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	events := make([]entities.Event, 0, len(updates))
	for _, update := range updates {
		events = append(events, s.updateToEvent(update))
	}

	s.offset = updates[len(updates)-1].ID + 1

	return events, nil
}

func (s *TgEventsService) ProcessMessage(ctx context.Context, event entities.Event) (err error) {
	meta, err := s.handleMeta(event)
	if err != nil {
		return e.WrapIfErr("cannot process message", err)
	}

	if strings.HasPrefix(event.Text, "/") {
		switch event.Text {
		case CmdStart:
			if err = s.tgClient.SendMessage(ctx, meta.ChatID, MsgHello); err != nil {
				return err
			}
		case CmdHelp:
			if err = s.tgClient.SendMessage(ctx, meta.ChatID, MsgHelp); err != nil {
				return err
			}
		default:
			if err = s.tgClient.SendMessage(ctx, meta.ChatID, MsgUnknownCommand); err != nil {
				return err
			}
		}
	} else {
		if err = s.tgClient.SendMessage(ctx, meta.ChatID, event.Text); err != nil {
			return err
		}
	}

	return nil

}

func (s *TgEventsService) handleMeta(event entities.Event) (entities.Meta, error) {
	res, ok := event.Meta.(entities.Meta)
	if !ok {
		return entities.Meta{}, e.WrapIfErr("couldn't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

func (s *TgEventsService) updateToEvent(update entities.Update) entities.Event {
	updType := s.updateType(update)

	res := entities.Event{
		Type: updType,
		Text: s.handleText(update),
	}

	if updType == entities.Message {
		res.Meta = entities.Meta{
			ChatID:   update.Message.Chat.ID,
			Username: update.Message.From.Username,
		}
	}

	return res
}

func (s *TgEventsService) handleText(update entities.Update) string {
	if update.Message == nil {
		return ""
	}

	return update.Message.Text
}

func (s *TgEventsService) updateType(upd entities.Update) entities.Type {
	if upd.Message == nil {
		return entities.Unknown
	}

	return entities.Message
}
