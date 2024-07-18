package app

import (
	"app/internal/lib/e"
	"app/internal/modules"
	tgclcontroller "app/internal/modules/telegram/client/controller"
	tgclvservice "app/internal/modules/telegram/client/service"
	"app/internal/modules/telegram/entities"
	tgevcontroller "app/internal/modules/telegram/events/controller"
	tgevservice "app/internal/modules/telegram/events/service"
	"context"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const batchSize = 100

type Config struct {
	tgHost     string
	tgApiToken string
}

type App struct {
	config      Config
	signalChan  chan os.Signal
	services    *modules.Services
	controllers *modules.Controllers
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
		events, err := a.controllers.TgEvents.Fetch(context.Background(), batchSize)
		if err != nil {
			log.Println("error fetching updates: ", err.Error())

			continue
		}

		if len(events) == 0 {
			time.Sleep(1 * time.Second)

			continue
		}

		if err := a.handleEvents(context.Background(), events); err != nil {
			log.Println("error handling events: ", err.Error())

			continue
		}

	}
}

func (a *App) handleEvents(ctx context.Context, events []entities.Event) error {
	for _, event := range events {
		log.Printf("got new event: %s", event.Text)

		if err := a.controllers.TgEvents.Process(ctx, event); err != nil {
			log.Printf("can't handle event: %s", err.Error())

			continue
		}
	}

	return nil
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

	return nil
}

func (a *App) init() error {

	a.config = Config{}

	if err := a.readConfig(); err != nil {
		return err
	}

	tgClientService := tgclvservice.NewTgService(a.config.tgHost, a.config.tgApiToken)
	tgClientController := tgclcontroller.NewTgControl(tgClientService)

	tgEventsService := tgevservice.NewTgEventsService(tgClientController)
	tgEventsController := tgevcontroller.NewTgEventsControl(tgEventsService)

	a.services = modules.NewServices(tgClientService, tgEventsService)
	a.controllers = modules.NewControllers(tgClientController, tgEventsController)

	a.signalChan = make(chan os.Signal, 1)
	signal.Notify(a.signalChan, syscall.SIGINT, syscall.SIGTERM)

	return nil
}
