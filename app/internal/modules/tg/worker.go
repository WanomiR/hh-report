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
	"time"
)

type Worker struct {
	ChatId          int
	IsWorking       bool
	StopWorking     chan bool
	Queries         []Query
	storage         storage.Storage
	vacancies       map[string]time.Time
	workingInterval time.Duration
	tgClient        Telegramer
	hhClient        hh.HeadHunterer
}

func NewWorker(chatId int, interval time.Duration, store storage.Storage, tgClient Telegramer, hhClient hh.HeadHunterer) *Worker {
	w := &Worker{
		ChatId:          chatId,
		StopWorking:     make(chan bool),
		Queries:         make([]Query, 0),
		vacancies:       make(map[string]time.Time),
		storage:         store,
		workingInterval: interval,
		tgClient:        tgClient,
		hhClient:        hhClient,
	}
	w.InitQueries()
	return w
}

func (w *Worker) Work() {
	w.IsWorking = true

	workTicker := time.NewTicker(w.workingInterval)
	cleanTicker := time.NewTicker(time.Hour * 24)

	for {
		select {
		case <-workTicker.C:
			currHour := time.Now().Hour()
			if currHour > 7 && currHour < 21 {
				for _, query := range w.Queries {
					go w.DoSearch(query)
				}
			}
		case <-cleanTicker.C:
			go w.CleanVacancies()

		case <-w.StopWorking:
			w.IsWorking = false
			return
		}
	}
}

func (w *Worker) DoSearch(q Query) {
	vacancies, err := w.hhClient.GetVacancies(q.Area, q.Role, q.Text, q.Experience, 1)
	if err != nil {
		log.Println(e.WrapIfErr(fmt.Sprintf("error getting vacancies for chat %d", w.ChatId), err).Error())
	}

	for _, v := range vacancies {
		if _, ok := w.vacancies[v.ID]; !ok {
			msg := fmt.Sprintf("Found new vacancy for <i>%s</i> with eperience <i>%s</i>:\nhttps://hh.ru/vacancy/%s", q.Text, q.Experience, v.ID)
			w.tgClient.SendMessage(w.ChatId, msg)
			w.vacancies[v.ID] = time.Now()
		}
	}
	log.Printf("conducted search: found %d vacancies for %s with experience %s\n", len(vacancies), q.Text, q.Experience)
}

func (w *Worker) CleanVacancies() {
	for id, createdAt := range w.vacancies {
		if createdAt.Before(time.Now().Add(-time.Hour * 72)) {
			delete(w.vacancies, id)
			log.Printf("deleted vacancy %s for chat %d\n", id, w.ChatId)
		}
	}
}

func (w *Worker) HandleAddQuery(query string) (err error) {
	defer func() { err = e.WrapIfErr("couldn't handle query", err) }()

	area, role, text, experience, err := w.ParseAddQuery(query)
	if err != nil {
		return err
	}

	file := storage.NewFile(w.ChatId, fmt.Sprintf("%s %s %s %s", area, role, text, experience))
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

	w.AppendAddQuery(area, role, text, experience)
	return nil
}

func (w *Worker) ParseAddQuery(regexMatch string) (area string, role string, text string, exp string, err error) {
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

func (w *Worker) AppendAddQuery(area, role, text, experience string) {
	q := Query{area, role, text, experience}
	w.Queries = append(w.Queries, q)
}

func (w *Worker) RemoveQuery(regexMatch string) (err error) {
	defer func() { err = e.WrapIfErr("couldn't remove query", err) }()

	queryId := strings.Split(regexMatch, " ")[1]
	id, err := strconv.Atoi(queryId)
	if err != nil {
		return err
	}

	id -= 1

	if len(w.Queries) == 0 {
		return errors.New("queries list is empty")
	} else if id < 0 || id >= len(w.Queries) {
		return errors.New("index out of range")
	}

	q := w.Queries[id]
	file := storage.NewFile(w.ChatId, fmt.Sprintf("%s %s %s %s", q.Area, q.Role, q.Text, q.Experience))

	if err = w.storage.Remove(file); err != nil {
		return err
	}

	w.Queries = append(w.Queries[:id], w.Queries[id+1:]...)
	log.Println("query removed:", file.Query)

	return nil
}

func (w *Worker) InitQueries() {
	files, err := w.storage.ReadAll(w.ChatId)
	if err != nil {
		log.Println(e.WrapIfErr("couldn't read queries for chat "+strconv.Itoa(w.ChatId), err).Error())
	}

	for _, file := range files {
		parts := strings.Split(file.Query, " ")
		if len(parts) != 4 {
			log.Println("expected 4 parts, got", file.Query)
			continue
		}

		query := Query{Area: parts[0], Role: parts[1], Text: parts[2], Experience: parts[3]}
		w.Queries = append(w.Queries, query)
	}

	log.Println("read", len(w.Queries), "queries for chat", w.ChatId)
}

func (w *Worker) ListQueries() (queries []string) {
	queries = make([]string, 0, len(w.Queries))

	for i, q := range w.Queries {
		queryText := fmt.Sprintf("%d – area: <i>%s</i>, role: <i>%s</i>, text: <i>%s</i>, experience: <i>%s</i>;", i+1, q.Area, q.Role, q.Text, q.Experience)
		queries = append(queries, queryText)
	}

	return queries
}
