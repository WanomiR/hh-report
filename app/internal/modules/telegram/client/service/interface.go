package service

import (
	"context"
	"net/url"
)

type TgClientServicer interface {
	DoRequest(ctx context.Context, method string, query url.Values) ([]byte, error)
}
