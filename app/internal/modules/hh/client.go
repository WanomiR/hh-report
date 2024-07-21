package hh

import (
	"app/internal/lib/e"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
)

type HeadHunterer interface {
	GetVacancies(area, role, text, experience string) ([]Vacancy, error)
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

func (c *Client) GetVacancies(area, role, text, experience string) (vacancies []Vacancy, err error) {
	defer func() { err = e.WrapIfErr("couldn't get vacancies", err) }()

	query := url.Values{
		"area":              []string{area},
		"experience":        []string{experience},
		"text":              []string{text},
		"professional_role": []string{role},
		"period":            []string{"3"},
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

	log.Println(req.URL.String())

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
