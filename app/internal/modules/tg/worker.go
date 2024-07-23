package tg

import (
	"app/internal/lib/e"
	"app/internal/modules/hh"
	"app/internal/storage"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Worker interface {
	Work()
	DoSearch(Query)
	HandleAddQuery(string) error
	RemoveQuery(string) error
	Queries() []Query
	ChatId() int
	IsWorking() bool
	StopWorking()
}

type WorkingAgent struct {
	chatId          int
	isWorking       bool
	stopWorking     chan bool
	workingInterval time.Duration
	queries         []Query
	vacancies       map[string]time.Time
	mux             *sync.RWMutex
	storage         storage.Storage
	tgClient        Telegramer
	hhClient        hh.HeadHunterer
}

func NewWorkingAgent(chatId int, interval time.Duration, store storage.Storage, tgClient Telegramer, hhClient hh.HeadHunterer) *WorkingAgent {
	w := &WorkingAgent{
		chatId:          chatId,
		stopWorking:     make(chan bool),
		workingInterval: interval,
		queries:         make([]Query, 0),
		vacancies:       make(map[string]time.Time),
		mux:             new(sync.RWMutex),
		storage:         store,
		tgClient:        tgClient,
		hhClient:        hhClient,
	}
	w.initQueries()
	return w
}

func (w *WorkingAgent) Work() {
	w.isWorking = true

	workTicker := time.NewTicker(w.workingInterval)
	cleanTicker := time.NewTicker(time.Hour * 24)

	for {
		select {
		case <-workTicker.C:
			currHour := time.Now().Hour()
			if currHour > 3 && currHour < 19 { // work only during the day, correct for GMT+3
				for _, query := range w.queries {
					go w.DoSearch(query)
				}
			}
		case <-cleanTicker.C:
			go w.cleanVacancies()

		case <-w.stopWorking:
			w.isWorking = false
			return
		}
	}
}

func (w *WorkingAgent) DoSearch(q Query) {
	vacancies, err := w.hhClient.GetVacancies(q.Area, q.Role, q.Text, q.Experience, 1)
	if err != nil {
		log.Println(e.WrapIfErr(fmt.Sprintf("error getting vacancies for chat %d", w.chatId), err).Error())
	}

	for _, v := range vacancies {
		if _, ok := w.vacancies[v.ID]; !ok {
			msg := fmt.Sprintf("Found new vacancy for <i>%s</i> with eperience <i>%s</i>:\nhttps://hh.ru/vacancy/%s", q.Text, q.Experience, v.ID)
			w.tgClient.SendMessage(w.chatId, msg)

			w.mux.Lock()
			w.vacancies[v.ID] = time.Now()
			w.mux.Unlock()
		}
	}
	log.Printf("conducted search: found %s%d%s vacancies for %s%s%s %s%s%s\n", Magenta, len(vacancies), Reset, Green, q.Text, Reset, Yellow, q.Experience, Reset)
}

func (w *WorkingAgent) HandleAddQuery(query string) (err error) {
	defer func() { err = e.WrapIfErr("couldn't handle query", err) }()

	area, role, text, experience, err := w.parseAddQuery(query)
	if err != nil {
		return err
	}

	file := storage.NewFile(w.chatId, fmt.Sprintf("%s %s %s %s", area, role, text, experience))
	exists, err := w.storage.IsExist(file)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("query already exists")
	}

	if err = w.storage.Save(file); err != nil {
		return err
	}

	w.appendAddQuery(area, role, text, experience)
	return nil
}

func (w *WorkingAgent) RemoveQuery(regexMatch string) (err error) {
	defer func() { err = e.WrapIfErr("couldn't remove query", err) }()

	queryId := strings.Split(regexMatch, " ")[1]
	id, err := strconv.Atoi(queryId)
	if err != nil {
		return err
	}

	id -= 1

	if len(w.queries) == 0 {
		return errors.New("queries list is empty")
	} else if id < 0 || id >= len(w.queries) {
		return errors.New("index out of range")
	}

	q := w.queries[id]
	file := storage.NewFile(w.chatId, fmt.Sprintf("%s %s %s %s", q.Area, q.Role, q.Text, q.Experience))

	if err = w.storage.Remove(file); err != nil {
		return err
	}

	w.queries = append(w.queries[:id], w.queries[id+1:]...)
	log.Println("query removed:", file.Query)

	return nil
}

func (w *WorkingAgent) Queries() []Query {
	return w.queries
}

func (w *WorkingAgent) ChatId() int {
	return w.chatId
}

func (w *WorkingAgent) IsWorking() bool {
	return w.isWorking
}

func (w *WorkingAgent) StopWorking() {
	w.stopWorking <- true
}

func (w *WorkingAgent) parseAddQuery(regexMatch string) (area string, role string, text string, exp string, err error) {
	parts := strings.Split(regexMatch, " ")
	if len(parts) != 5 {
		return "", "", "", "", errors.New(fmt.Sprintf("should be exactly 5 parts, got: %d", len(parts)))
	}

	area, role, text, exp = parts[1], parts[2], parts[3], parts[4]

	switch exp {
	case "0":
		exp = "noExperience"
	case "1-3":
		exp = "between1And3"
	case "3-6":
		exp = "between3And6"
	case "6":
		exp = "moreThan6"
	default:
		exp = ""
	}

	return area, role, text, exp, nil
}

func (w *WorkingAgent) appendAddQuery(area, role, text, experience string) {
	q := Query{Area: area, Role: role, Text: text, Experience: experience}
	w.queries = append(w.queries, q)
}

func (w *WorkingAgent) cleanVacancies() {
	for id, createdAt := range w.vacancies {
		if createdAt.Before(time.Now().Add(-time.Hour * 72)) {

			w.mux.RLock()
			delete(w.vacancies, id)
			w.mux.RUnlock()

			log.Printf("deleted vacancy %s for chat %d\n", id, w.chatId)
		}
	}
}

func (w *WorkingAgent) initQueries() {
	files, err := w.storage.ReadAll(w.chatId)
	if err != nil {
		log.Println(e.WrapIfErr("couldn't read queries for chat "+strconv.Itoa(w.chatId), err).Error())
	}

	for _, file := range files {
		parts := strings.Split(file.Query, " ")
		if len(parts) != 4 {
			log.Println("expected 4 parts, got", file.Query)
			continue
		}

		query := Query{Area: parts[0], Role: parts[1], Text: parts[2], Experience: parts[3]}
		w.queries = append(w.queries, query)
	}

	log.Printf("read %s%d%s queries for %s%d%s", Magenta, len(w.queries), Reset, Green, w.chatId, Reset)
}
