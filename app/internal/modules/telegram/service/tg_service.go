package service

import (
	"app/internal/lib/e"
	"app/internal/modules/telegram/entities"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	methodGetUpdates  = "getUpdates"  // Use this method to receive incoming updates using long polling. Returns an Array of Update objects
	methodSendMessage = "sendMessage" // Use this method to send text messages. On success, the sent Message is returned
)

type TgServicer interface {
	GetUpdates(ctx context.Context) ([]entities.Update, error)
	ProcessUpdates(ctx context.Context, updates []entities.Update)
}

type TgService struct {
	host     string
	basePath string
	client   *http.Client
	offset   int
	limit    int
	timeout  int
}

func NewTgService(host string, token string, batchSize, timeout int) *TgService {
	return &TgService{
		host:     host,          // api.telegram.org
		basePath: "bot" + token, // app<token>
		client:   new(http.Client),
		offset:   0,
		limit:    batchSize,
		timeout:  timeout,
	}
}

func (s *TgService) GetUpdates(ctx context.Context) (updates []entities.Update, err error) {
	defer func() { err = e.WrapIfErr("couldn't get updates", err) }()

	query := url.Values{
		"offset":  []string{strconv.Itoa(s.offset)},
		"limit":   []string{strconv.Itoa(s.limit)},
		"timeout": []string{strconv.Itoa(s.timeout)},
	}

	data, err := s.doRequest(ctx, methodGetUpdates, query)
	if err != nil {
		return nil, err
	}

	var res entities.UpdatesResponse
	if err = json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	if !res.Ok {
		return nil, fmt.Errorf(res.Description)
	}

	if updates = res.Result; len(updates) == 0 {
		return updates, nil
	}

	s.offset = updates[len(updates)-1].ID + 1

	return updates, nil
}

func (s *TgService) ProcessUpdates(ctx context.Context, updates []entities.Update) {
	for _, update := range updates {

		// ignore everything that is not a message
		if update.Message == nil {
			continue
		}

		message := update.Message
		log.Println("got new message: ", message.Text)

		if err := s.processMessage(ctx, message); err != nil {
			log.Println(err.Error())
		}
	}

}

func (s *TgService) processMessage(ctx context.Context, message *entities.IncomingMessage) (err error) {

	if strings.HasPrefix(message.Text, "/") {
		switch message.Text {
		case "/start":
			if err = s.sendMessage(ctx, message.Chat.ID, "Hello!"); err != nil {
				return err
			}
		case "/help":
			if err = s.sendMessage(ctx, message.Chat.ID, "Here is your help"); err != nil {
				return err
			}
		default:
			if err = s.sendMessage(ctx, message.Chat.ID, "Unknown command"); err != nil {
				return err
			}
		}
	} else {
		if err = s.sendMessage(ctx, message.Chat.ID, message.Text); err != nil {
			return err
		}
	}

	return nil

}

func (s *TgService) sendMessage(ctx context.Context, chatId int, text string) error {

	query := url.Values{
		"chat_id": []string{strconv.Itoa(chatId)},
		"text":    []string{text},
	}

	if _, err := s.doRequest(ctx, methodSendMessage, query); err != nil {
		return e.WrapIfErr("couldn't send message", err)
	}

	return nil
}

func (s *TgService) doRequest(ctx context.Context, method string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("cannot do request", err) }()

	// https://api.telegram.org/bot<token>/METHOD_NAME
	requestUrl := url.URL{
		Scheme: "https",
		Host:   s.host,
		Path:   path.Join(s.basePath, method),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
