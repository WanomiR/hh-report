package app

import (
	"app/internal/lib/e"
	"app/internal/modules"
	tgcontroller "app/internal/modules/telegram/controller"
	tgservice "app/internal/modules/telegram/service"
	"context"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	tgHost         string
	tgApiToken     string
	hhHost         string
	hhClientId     string
	hhClientSecret string
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

func (a *App) Run(ctx context.Context) {
	log.Println("Firing up the app...")

	go a.controllers.Tg.Serve(ctx)

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
	a.config.hhClientId = os.Getenv("HH_CLIENT_ID")
	a.config.hhClientSecret = os.Getenv("HH_CLIENT_SECRET")

	return nil
}

func (a *App) init() error {

	a.config = Config{}

	if err := a.readConfig(); err != nil {
		return err
	}

	tgs := tgservice.NewTgService(a.config.tgHost, a.config.tgApiToken, 100, 0)
	tgc := tgcontroller.NewTgControl(tgs)

	a.services = modules.NewServices(tgs)
	a.controllers = modules.NewControllers(tgc)

	a.signalChan = make(chan os.Signal, 1)
	signal.Notify(a.signalChan, syscall.SIGINT, syscall.SIGTERM)

	return nil
}
