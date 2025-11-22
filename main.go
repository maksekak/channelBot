package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var (
	bot           *tgbotapi.BotAPI
	spreadsheetID string
	authCode      string
	sheetName     = "Sheet1"
)

func main() {

	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	spreadsheetID = os.Getenv("sheetId")
	authCode = os.Getenv("authCode")
	bot, err = tgbotapi.NewBotAPI(os.Getenv("token"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for {
		handleUpdate(<-updates)
	}
}
