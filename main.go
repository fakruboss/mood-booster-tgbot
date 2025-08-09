package main

import (
	"log"
	"os"

	"github.com/you/moodbot/favorites"
	"github.com/you/moodbot/handlers"
	"github.com/you/moodbot/translation"
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
	favoriteManager := favorites.NewFavoriteManager()
	translator := translation.NewTranslator()
	languageManager := translation.NewLanguageManager("user_data")
	messageHandler := handlers.NewMessageHandler(bot, voteManager, favoriteManager, translator, languageManager)

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
				case "favorites", "favorite":
					messageHandler.SendFavorites(update.Message.Chat.ID, update.Message.From.ID)
				case "language", "lang":
					messageHandler.SendLanguageKeyboard(update.Message.Chat.ID)
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

			// Handle language selection
			if data == "lang_en" || data == "lang_hi" || data == "lang_ta" {
				response := messageHandler.HandleLanguageSelection(data, chatID, userID)
				cb := tgbotapi.NewCallback(update.CallbackQuery.ID, response)
				_, _ = bot.Request(cb)
				continue
			}

			// Handle favorites
			if favoriteManager.IsFavoriteCallback(data) {
				originalContent := ""
				if update.CallbackQuery.Message.Caption != "" {
					originalContent = update.CallbackQuery.Message.Caption
				} else {
					originalContent = update.CallbackQuery.Message.Text
				}
				cb := tgbotapi.NewCallback(update.CallbackQuery.ID, messageHandler.HandleFavoriteCallback(data, userID, originalContent))
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