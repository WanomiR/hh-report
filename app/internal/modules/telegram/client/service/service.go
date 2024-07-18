package service

import (
	"app/internal/lib/e"
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
)

type TgClientService struct {
	host     string
	basePath string
	client   *http.Client
}

func NewTgService(host string, token string) *TgClientService {
	return &TgClientService{
		host:     host,          // api.telegram.org
		basePath: "bot" + token, // app<token>
		client:   new(http.Client),
	}
}

func (s *TgClientService) DoRequest(ctx context.Context, tgMethod string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("cannot do request", err) }()

	// https://api.telegram.org/bot<token>/METHOD_NAME
	requestUrl := url.URL{
		Scheme: "https",
		Host:   s.host,
		Path:   path.Join(s.basePath, tgMethod),
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
