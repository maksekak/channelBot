package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(update.Message)
	case update.CallbackQuery != nil:

	}
}
func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text
	userId := message.From.ID
	if user == nil {
		return
	}

	log.Printf("%s wrote %s %s", user.FirstName, userId, text)
}
func sendReply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	bot.Send(msg)
}
