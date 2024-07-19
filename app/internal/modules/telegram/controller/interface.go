package controller

import "context"

type TgController interface {
	Serve(ctx context.Context)
}
