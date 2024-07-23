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
	"time"
)

const (
	Reset   = "\033[0m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Magenta = "\033[35m"
)

const (
	methodGetUpdates  = "getUpdates"  // Use this method to receive incoming updates using long polling. Returns an Array of Update objects
	methodSendMessage = "SendMessage" // Use this method to send text messages. On success, the sent Message is returned
)

type Telegramer interface {
	GetUpdates() ([]Update, error)
	ProcessUpdates(updates []Update)
	SendMessage(chatId int, text string)
}

type Client struct {
	host     string
	basePath string
	tgClient *http.Client
	offset   int
	limit    int
	timeout  int

	hhClient hh.HeadHunterer

	workers  map[int]Worker
	interval time.Duration
	storage  storage.Storage

	reAdd    *regexp.Regexp
	reRemove *regexp.Regexp
}

func NewTgClient(host string, token string, batchSize, timeout int, hhClient hh.HeadHunterer, storage storage.Storage, workingInterval time.Duration) *Client {
	return &Client{
		host:     host,          // api.tg.org
		basePath: "bot" + token, // app<token>
		tgClient: new(http.Client),
		offset:   0,
		limit:    batchSize,
		timeout:  timeout,

		hhClient: hhClient,

		workers:  make(map[int]Worker),
		interval: workingInterval,
		storage:  storage,

		reAdd:    regexp.MustCompile(`add: \d+ \d+ [a-zA-Zа-яА-Я-]+ (-|0|1-3|3-6|6)`),
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

	log.Printf("got message %s%v%s from %s%d%s", Magenta, message.Text, Reset, Green, message.Chat.ID, Reset)

	switch {
	case strings.HasPrefix(message.Text, "/"):
		c.processCommand(message.Text, worker)

	case c.reAdd.MatchString(message.Text):
		// adding new query to the wr
		match := c.reAdd.FindStringSubmatch(message.Text)[0]
		// handle possible error
		if err := worker.HandleAddQuery(match); err != nil {
			c.SendMessage(worker.ChatId(), e.WrapIfErr("error adding query", err).Error())
		} else {
			c.SendMessage(worker.ChatId(), "Query added 👌🏻")
		}

	case c.reRemove.MatchString(message.Text):
		match := c.reRemove.FindStringSubmatch(message.Text)[0]
		// handle possible error
		if err := worker.RemoveQuery(match); err != nil {
			c.SendMessage(worker.ChatId(), e.WrapIfErr("error removing query", err).Error())
		} else {
			c.SendMessage(worker.ChatId(), "Query removed 🗑️")
		}

	default:
		// just mirror for now
		c.SendMessage(worker.ChatId(), message.Text)
	}
}

func (c *Client) handleWorker(chatId int) Worker {
	worker, ok := c.workers[chatId]
	if !ok {
		worker = NewWorkingAgent(chatId, c.interval, c.storage, c, c.hhClient)
		c.workers[chatId] = worker
		log.Printf("new worker created %s%d%s", Green, chatId, Reset)
	}
	return worker
}

func (c *Client) processCommand(command string, worker Worker) {
	switch command {

	case "/check":
		for _, q := range worker.Queries() {
			worker.DoSearch(q)
		}
		c.SendMessage(worker.ChatId(), fmt.Sprintf("Checked %d queries 👌🏻", len(worker.Queries())))

	case "/start":
		if len(worker.Queries()) == 0 {
			c.SendMessage(worker.ChatId(), messageNoQueries+"\n\n"+messageAddQuery)
		} else if !worker.IsWorking() {
			go worker.Work()
			c.SendMessage(worker.ChatId(), "Worker started!")
		}

	case "/stop":
		if worker.IsWorking() {
			worker.StopWorking()
			c.SendMessage(worker.ChatId(), "Worker stopped.")
		}

	case "/help":
		c.SendMessage(worker.ChatId(), messageHelp)

	case "/queries":
		if queries := worker.Queries(); len(queries) > 0 {
			c.SendMessage(worker.ChatId(), "Active queries:")
			for i, q := range queries {
				msg := fmt.Sprintf("%d – area: <i>%s</i>, role: <i>%s</i>, text: <i>%s</i>, experience: <i>%s</i>;", i+1, q.Area, q.Role, q.Text, q.Experience)
				c.SendMessage(worker.ChatId(), msg)
			}
		} else {
			c.SendMessage(worker.ChatId(), messageNoQueries)
		}

	case "/status":
		if worker.IsWorking() {
			msg := fmt.Sprintf("Working on %d queries with interval %v", len(worker.Queries()), c.interval)
			c.SendMessage(worker.ChatId(), msg)
		} else {
			c.SendMessage(worker.ChatId(), "Worker not started.")
		}

	default:
		c.SendMessage(worker.ChatId(), "Unknown command")
	}
}

func (c *Client) SendMessage(chatId int, text string) {

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
