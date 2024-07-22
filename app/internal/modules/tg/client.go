package tg

import (
	"app/internal/lib/e"
	"app/internal/modules/hh"
	"app/internal/storage"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	methodGetUpdates  = "getUpdates"  // Use this method to receive incoming updates using long polling. Returns an Array of Update objects
	methodSendMessage = "sendMessage" // Use this method to send text messages. On success, the sent Message is returned
)

type Telegramer interface {
	GetUpdates() ([]Update, error)
	ProcessUpdates(updates []Update)
}

type Client struct {
	host     string
	basePath string
	tgClient *http.Client
	hhClient hh.HeadHunterer
	offset   int
	limit    int
	timeout  int
	workers  map[int]*Worker
	storage  storage.Storage
	reAdd    *regexp.Regexp
	reRemove *regexp.Regexp
}

func NewTgClient(host string, token string, batchSize, timeout int, hhClient hh.HeadHunterer, storage storage.Storage) *Client {
	return &Client{
		host:     host,          // api.tg.org
		basePath: "bot" + token, // app<token>
		tgClient: new(http.Client),
		hhClient: hhClient,
		offset:   0,
		limit:    batchSize,
		timeout:  timeout,
		workers:  make(map[int]*Worker),
		storage:  storage,
		reAdd:    regexp.MustCompile(`add: \d+ \d+ \w+ (-|0|1-3|3-6|6)`),
		reRemove: regexp.MustCompile(`remove: \d+`),
	}
}

func (c *Client) GetUpdates() (updates []Update, err error) {
	defer func() { err = e.WrapIfErr("couldn't get updates", err) }()

	query := url.Values{
		"offset":  []string{strconv.Itoa(c.offset)},
		"limit":   []string{strconv.Itoa(c.limit)},
		"timeout": []string{strconv.Itoa(c.timeout)},
	}

	data, err := c.doRequest(methodGetUpdates, query)
	if err != nil {
		return nil, err
	}

	var res UpdatesResponse
	if err = json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	if !res.Ok {
		return nil, fmt.Errorf(res.Description)
	}

	if updates = res.Result; len(updates) == 0 {
		return updates, nil
	}

	c.offset = updates[len(updates)-1].ID + 1

	return updates, nil
}

func (c *Client) ProcessUpdates(updates []Update) {
	for _, update := range updates {

		// ignore everything that is not a message
		if update.Message == nil {
			continue
		}

		c.processMessage(update.Message)
	}
}

func (c *Client) processMessage(message *Message) {
	worker := c.handleWorker(message.Chat.ID)

	log.Println("got new message:", message.Text, fmt.Sprintf("📍 [worker id: %d, isWorking: %v]", worker.ChatId, worker.IsWorking))

	switch {
	case strings.HasPrefix(message.Text, "/"):
		c.processCommand(message.Text, worker)

	case c.reAdd.MatchString(message.Text):
		// adding new query to the worker
		match := c.reAdd.FindStringSubmatch(message.Text)[0]
		// handle possible error
		if err := worker.HandleAddQuery(match); err != nil {
			c.sendMessage(worker.ChatId, e.WrapIfErr("error adding query", err).Error())
		} else {
			c.sendMessage(worker.ChatId, "Query added 👌🏻")
		}

	case c.reRemove.MatchString(message.Text):
		match := c.reRemove.FindStringSubmatch(message.Text)[0]
		// handle possible error
		if err := worker.RemoveQuery(match); err != nil {
			c.sendMessage(worker.ChatId, e.WrapIfErr("error removing query", err).Error())
		} else {
			c.sendMessage(worker.ChatId, "Query removed 🗑️")
		}

	default:
		// just mirror for now
		c.sendMessage(worker.ChatId, message.Text)
	}
}

func (c *Client) handleWorker(chatId int) *Worker {
	worker, ok := c.workers[chatId]
	if !ok {
		worker = &Worker{
			ChatId:      chatId,
			StopWorking: make(chan bool),
			queries:     make([]Query, 0),
			storage:     c.storage,
		}
		worker.InitQueries()
		c.workers[chatId] = worker
	}
	return worker
}

func (c *Client) processCommand(command string, worker *Worker) {
	switch command {

	case "/check":
		data, err := c.hhClient.GetVacancies("1", "96", "golang", "noExperience")
		if err != nil {
			log.Println(err)
		} else {
			log.Println("vacancies found:", len(data))
		}
		c.sendMessage(worker.ChatId, "checked")

	case "/start":
		// TODO:
		//  - show help if there are no queries
		//  - otherwise launch workers on current queries

	case "/help":
		c.sendMessage(worker.ChatId, messageHelp)

	case "/queries":
		if queries := worker.ListQueries(); len(queries) > 0 {
			c.sendMessage(worker.ChatId, "Active queries:")
			for _, query := range worker.ListQueries() {
				c.sendMessage(worker.ChatId, query)
			}
		} else {
			c.sendMessage(worker.ChatId, "No active queries")
		}

	case "/status":
		// TODO: show the number of active workers in this chat

	case "/stop":
		// TODO: stop all workers

	default:
		c.sendMessage(worker.ChatId, "Unknown command")
	}
}

func (c *Client) sendMessage(chatId int, text string) {

	query := url.Values{
		"chat_id":    []string{strconv.Itoa(chatId)},
		"text":       []string{text},
		"parse_mode": []string{"HTML"},
	}

	if _, err := c.doRequest(methodSendMessage, query); err != nil {
		log.Println("couldn't send message", err)
	}
}

func (c *Client) doRequest(method string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("cannot do request", err) }()

	// https://api.telegram.org/bot<token>/METHOD_NAME
	requestUrl := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, method),
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.tgClient.Do(req)
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
