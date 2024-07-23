package app

import (
	"app/internal/lib/e"
	"app/internal/modules/hh"
	"app/internal/modules/tg"
	"app/internal/storage"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	tgHost     string
	tgApiToken string
	hhHost     string
}

type App struct {
	config     Config
	signalChan chan os.Signal
	storage    storage.Storage
	wAgent     tg.Worker
	hhClient   hh.HeadHunterer
	tgClient   tg.Telegramer
}

func NewApp() (a *App, err error) {
	defer func() { err = e.WrapIfErr("failed to init app", err) }()

	a = &App{}

	if err = a.init(); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Signal() <-chan os.Signal {
	return a.signalChan
}

func (a *App) Run() {
	log.Println("Firing up the app...")

	for {
		// get updates every second
		time.Sleep(1 * time.Second)

		updates, err := a.tgClient.GetUpdates()
		if err != nil {
			log.Println(err.Error()) // skip if an error
			continue
		}

		if len(updates) == 0 { // skip if no updates
			continue
		}

		a.tgClient.ProcessUpdates(updates)
	}

}

func (a *App) readConfig(envPath ...string) (err error) {
	if len(envPath) > 0 {
		err = godotenv.Load(envPath[0])
	} else {
		err = godotenv.Load()
	}

	if err != nil {
		return e.WrapIfErr("can't read .env file", err)
	}

	a.config.tgHost = os.Getenv("TG_HOST")
	a.config.tgApiToken = os.Getenv("TG_API_TOKEN")
	a.config.hhHost = os.Getenv("HH_HOST")

	return nil
}

func (a *App) init() error {

	a.config = Config{}

	if err := a.readConfig(); err != nil {
		return err
	}

	a.hhClient = hh.NewHhClient(a.config.hhHost)
	a.storage = storage.NewQueriesStorage("storage")
	a.tgClient = tg.NewTgClient(
		a.config.tgHost, a.config.tgApiToken, 100, 0, a.hhClient, a.storage, time.Minute*10,
	)

	a.signalChan = make(chan os.Signal, 1)
	signal.Notify(a.signalChan, syscall.SIGINT, syscall.SIGTERM)

	return nil
}
