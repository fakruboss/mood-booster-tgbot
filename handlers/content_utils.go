package handlers

import (
	"context"
	"time"

	"github.com/you/moodbot/fetchers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// sendContentWithImage is a helper method to send content with optional image
func (mh *MessageHandler) sendContentWithImage(chatID int64, text, imageQuery, contentType string) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var imageURL string
	if imageQuery != "" {
		imageURL, _ = fetchers.FetchUnsplashImage(ctx, imageQuery)
	}

	messageID := int(time.Now().Unix())
	
	if imageURL != "" {
		// Send photo with caption
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(imageURL))
		photo.Caption = text
		photo.ParseMode = "Markdown"
		photo.ReplyMarkup = mh.voteManager.CreateVotingKeyboard(contentType, messageID)
		_, err := mh.bot.Send(photo)
		if err != nil {
			// Fallback to text message if photo fails
			mh.sendTextMessage(chatID, text, contentType, messageID)
		}
	} else {
		// Send text message
		mh.sendTextMessage(chatID, text, contentType, messageID)
	}
}

// sendTextMessage is a helper method to send text-only messages
func (mh *MessageHandler) sendTextMessage(chatID int64, text, contentType string, messageID int) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = mh.voteManager.CreateVotingKeyboard(contentType, messageID)
	mh.bot.Send(msg)
}