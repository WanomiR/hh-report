package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
	"time"
)

var bot *tgbotapi.BotAPI
var stopChan = make(chan bool)
var isTicking bool

func main() {
	var err error

	if err = godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	bot, err = tgbotapi.NewBotAPI(os.Getenv("TG_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		handleUpdate(update)
	}

}

func handleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(update.Message)
	}
}

func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text
	chatID := message.Chat.ID

	if user == nil {
		return
	}

	var err error
	if strings.HasPrefix(text, "/") {
		err = handleCommand(chatID, text)
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.Entities = message.Entities
		_, err = bot.Send(msg)
	}

	if err != nil {
		log.Printf("Error handling command: %v", err)
	}

}

func handleCommand(chatId int64, command string) error {
	switch {
	case !isTicking && command == "/start":
		isTicking = true
		go startTicker(chatId)
		msg := tgbotapi.NewMessage(chatId, "ticker started!")
		bot.Send(msg)
	case isTicking && command == "/stop":
		stopChan <- true
		isTicking = false
		msg := tgbotapi.NewMessage(chatId, "ticker stopped!")
		bot.Send(msg)
	}

	return nil
}

func startTicker(chatId int64) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			msg := tgbotapi.NewMessage(chatId, "tick")
			_, _ = bot.Send(msg)
		case <-stopChan:
			ticker.Stop()
			return
		}
	}
}
