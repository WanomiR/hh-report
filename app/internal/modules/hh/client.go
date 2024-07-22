package hh

import (
	"app/internal/lib/e"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HeadHunterer interface {
	GetVacancies(area, role, text, experience string, period int) ([]Vacancy, error)
}

type Client struct {
	host   string
	client *http.Client
}

func NewHhClient(host string) *Client {
	return &Client{
		host:   host, // api.hh.ru
		client: new(http.Client),
	}
}

func (c *Client) GetVacancies(area, role, text, experience string, period int) (vacancies []Vacancy, err error) {
	defer func() { err = e.WrapIfErr("couldn't get vacancies", err) }()

	dateFrom := time.Now().AddDate(0, 0, -period).Format("2006-01-02")

	query := url.Values{
		"area":              []string{area},
		"text":              []string{text},
		"professional_role": []string{role},
		"date_from":         []string{dateFrom},
	}

	if experience != "" {
		experience = strings.Replace(experience, "-", " ", -1)
		query["experience"] = []string{experience}
	}

	data, err := c.doRequest("vacancies", query)
	if err != nil {
		return nil, err
	}

	var resp VacanciesResponse
	if err = json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

func (c *Client) doRequest(method string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("couldn't do request", err) }()

	requestUrl := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   method, // vacancies or employer or others
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
