package tg

import (
	"app/internal/lib/e"
	"app/internal/storage"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Worker struct {
	IsWorking   bool
	StopWorking chan bool
	ChatId      int
	queries     []Query
	storage     storage.Storage
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
		return "", "", "", "", errors.New(fmt.Sprintf("should be exactly 4 parts, got: %d", len(parts)))
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
	w.queries = append(w.queries, q)
}

func (w *Worker) RemoveQuery(regexMatch string) (err error) {
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
	file := storage.NewFile(w.ChatId, fmt.Sprintf("%s %s %s %s", q.Area, q.Role, q.Text, q.Experience))

	if err = w.storage.Remove(file); err != nil {
		return err
	}

	w.queries = append(w.queries[:id], w.queries[id+1:]...)
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
		w.queries = append(w.queries, query)
	}

	log.Println("read", len(w.queries), "queries for chat", w.ChatId)
}

func (w *Worker) ListQueries() (queries []string) {
	queries = make([]string, 0, len(w.queries))

	for i, q := range w.queries {
		queryText := fmt.Sprintf("%d – area: <i>%s</i>, role: <i>%s</i>, text: <i>%s</i>, experience: <i>%s</i>;", i+1, q.Area, q.Role, q.Text, q.Experience)
		queries = append(queries, queryText)
	}

	return queries
}
