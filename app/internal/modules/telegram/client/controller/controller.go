package controller

import (
	"app/internal/lib/e"
	"app/internal/modules/telegram/client/service"
	"app/internal/modules/telegram/entities"
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

const (
	methodGetUpdates  = "getUpdates"  // Use this method to receive incoming updates using long polling. Returns an Array of Update objects
	methodSendMessage = "sendMessage" // Use this method to send text messages. On success, the sent Message is returned
)

type TgClientControl struct {
	service service.TgClientServicer
}

func NewTgControl(service service.TgClientServicer) *TgClientControl {
	return &TgClientControl{service: service}
}

// GetUpdates godoc
func (c *TgClientControl) GetUpdates(ctx context.Context, offset, limit, timeout int) (updates []entities.Update, err error) {
	defer func() { err = e.WrapIfErr("cannot get updates", err) }()

	query := url.Values{
		"offset":  []string{strconv.Itoa(offset)},
		"limit":   []string{strconv.Itoa(limit)},
		"timeout": []string{strconv.Itoa(timeout)},
	}

	data, err := c.service.DoRequest(ctx, methodGetUpdates, query)
	if err != nil {
		return nil, err
	}

	var res entities.UpdatesResponse
	if err = json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (c *TgClientControl) SendMessage(ctx context.Context, chatID int, text string) error {

	query := url.Values{
		"chat_id": []string{strconv.Itoa(chatID)},
		"text":    []string{text},
	}

	if _, err := c.service.DoRequest(ctx, methodSendMessage, query); err != nil {
		return e.WrapIfErr("cannot send message", err)
	}

	return nil
}
