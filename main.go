package main

import (
	"log"
	"os"

	"github.com/you/moodbot/handlers"
	"github.com/you/moodbot/voting"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("set TELEGRAM_BOT_TOKEN env variable")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Initialize managers and handlers
	voteManager := voting.NewVoteManager()
	messageHandler := handlers.NewMessageHandler(bot, voteManager)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					_ = messageHandler.SendMoodKeyboard(update.Message.Chat.ID)
				case "surprise":
					messageHandler.HandleSurprise(update.Message.Chat.ID)
				case "help":
					messageHandler.SendHelpMessage(update.Message.Chat.ID)
				default:
					messageHandler.SendHelpMessage(update.Message.Chat.ID)
				}
			}
			continue
		}

		if update.CallbackQuery != nil {
			data := update.CallbackQuery.Data
			chatID := update.CallbackQuery.Message.Chat.ID
			userID := update.CallbackQuery.From.ID

			// Handle voting
			if voteManager.IsVoteCallback(data) {
				cb := tgbotapi.NewCallback(update.CallbackQuery.ID, voteManager.HandleVote(data, userID))
				_, _ = bot.Request(cb)
				continue
			}

			// Acknowledge callback (remove "loading")
			cb := tgbotapi.NewCallback(update.CallbackQuery.ID, "Working on it...")
			_, _ = bot.Request(cb)

			// Handle mood selection
			messageHandler.HandleMoodSelection(data, chatID)
		}
	}
}