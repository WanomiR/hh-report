package service

import (
	"app/internal/lib/e"
	"context"
	"io"
	"net/http"
	"net/url"
)

type HhServicer interface{}

type HhService struct {
	host   string
	client *http.Client
}

func NewHhService(host string, basePath string) *HhService {
	return &HhService{
		host: host, // api.hh.ru
	}
}

func (s *HhService) doRequest(ctx context.Context, method string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("couldn't do request", err) }()

	requestUrl := url.URL{
		Scheme: "https",
		Host:   s.host,
		Path:   method, // vacancies or employer or others
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

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
