package tg

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Worker struct {
	IsWorking   bool
	StopWorking chan bool
	ChatId      int
	queries     []Query
}

func (w *Worker) HandleAddQuery(query string) error {
	area, role, text, experience, err := w.ParseAddQuery(query)
	if err != nil {
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

	w.queries = append(w.queries[:id], w.queries[id+1:]...)

	return nil
}

func (w *Worker) ListQueries() (queries []string) {
	queries = make([]string, 0, len(w.queries))

	for i, q := range w.queries {
		queryText := fmt.Sprintf("%d. Area: %s, role: %s, text: %s, experience: %s;", i+1, q.Area, q.ProfessionalRole, q.Text, q.Experience)
		queries = append(queries, queryText)
	}

	return queries
}
