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
	"time"
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
	chats    map[int]*entities.TgChat
}

func NewTgService(host string, token string, batchSize, timeout int) *TgService {
	return &TgService{
		host:     host,          // api.telegram.org
		basePath: "bot" + token, // app<token>
		client:   new(http.Client),
		offset:   0,
		limit:    batchSize,
		timeout:  timeout,
		chats:    make(map[int]*entities.TgChat),
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

		s.processMessage(ctx, update.Message)
	}
}

func (s *TgService) processMessage(ctx context.Context, message *entities.Message) {
	chat := s.handleChat(message.Chat.ID)

	log.Println("got new message:", message.Text, fmt.Sprintf("ℹ️ [chat id: %d, isTicking: %v]", chat.Id, chat.IsTicking))

	if strings.HasPrefix(message.Text, "/") {
		switch message.Text {
		case "/start":
			if !chat.IsTicking {
				chat.IsTicking = true
				go func() {
					s.sendMessage(ctx, chat.Id, "Ticker has started!")
					ticker := time.NewTicker(2 * time.Second)
					for {
						select {
						case <-ticker.C:
							s.sendMessage(ctx, chat.Id, "tick")
						case <-chat.StopTicking:
							return
						}
					}
				}()
			}
		case "/stop":
			if chat.IsTicking {
				chat.IsTicking = false
				chat.StopTicking <- true
				s.sendMessage(ctx, chat.Id, "Ticker has stopped!")
			}
		default:
			s.sendMessage(ctx, chat.Id, "Unknown command")
		}
	} else {
		s.sendMessage(ctx, chat.Id, message.Text)
	}
}

func (s *TgService) handleChat(chatId int) *entities.TgChat {
	chat, ok := s.chats[chatId]
	if !ok {
		chat = &entities.TgChat{Id: chatId, StopTicking: make(chan bool)}
		s.chats[chatId] = chat
	}
	return chat
}

func (s *TgService) sendMessage(ctx context.Context, chatId int, text string) {

	query := url.Values{
		"chat_id": []string{strconv.Itoa(chatId)},
		"text":    []string{text},
	}

	if _, err := s.doRequest(ctx, methodSendMessage, query); err != nil {
		log.Println("couldn't send message", err)
	}
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
