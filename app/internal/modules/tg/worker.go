package tg

type Worker struct {
	IsWorking     bool
	StopWorking   chan bool
	ChatId        int
	SearchQueries []Query
}
