package handlers

import (
	"context"
	"time"

	"github.com/you/moodbot/fetchers"
	"github.com/you/moodbot/models"
	"github.com/you/moodbot/translation"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendMoodKeyboard sends the main mood selection keyboard
func (mh *MessageHandler) SendMoodKeyboard(chatID int64) error {
	userLang := mh.languageManager.GetUserLanguage(chatID)
	
	// Translate the prompt
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	prompt, _ := mh.translator.TranslateText(ctx, "What's your mood today?", userLang)
	
	msg := tgbotapi.NewMessage(chatID, prompt)
	
	// Translate button labels
	funnyText, _ := mh.translator.TranslateText(ctx, "Funny", userLang)
	inspiringText, _ := mh.translator.TranslateText(ctx, "Inspiring", userLang) 
	educationalText, _ := mh.translator.TranslateText(ctx, "Educational", userLang)
	relaxingText, _ := mh.translator.TranslateText(ctx, "Relaxing", userLang)
	adventurousText, _ := mh.translator.TranslateText(ctx, "Adventurous", userLang)
	thoughtfulText, _ := mh.translator.TranslateText(ctx, "Thoughtful", userLang)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(funnyText+" 😂", "funny"),
			tgbotapi.NewInlineKeyboardButtonData(inspiringText+" 💡", "inspiring"),
			tgbotapi.NewInlineKeyboardButtonData(educationalText+" 📚", "educational"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(relaxingText+" 🌿", "relaxing"),
			tgbotapi.NewInlineKeyboardButtonData(adventurousText+" 🌟", "adventurous"),
			tgbotapi.NewInlineKeyboardButtonData(thoughtfulText+" 🤔", "thoughtful"),
		),
	)
	msg.ReplyMarkup = keyboard
	_, err := mh.bot.Send(msg)
	return err
}

// HandleMoodSelection handles mood-based content requests
func (mh *MessageHandler) HandleMoodSelection(data string, chatID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var body string
	var fetchErr error
	var contentType string

	var imageQuery string

	switch data {
	case "funny":
		contentType = "funny"
		imageQuery = "funny"
		body, fetchErr = fetchers.FetchJoke(ctx)
	case "inspiring":
		contentType = "inspiring"
		imageQuery = "inspiration"
		body, fetchErr = fetchers.FetchZenQuote(ctx)
	case "educational":
		contentType = "educational"
		imageQuery = "books"
		body, fetchErr = fetchers.FetchFact(ctx)
	case "relaxing":
		contentType = "relaxing"
		imageQuery = "nature"
		body, fetchErr = fetchers.FetchZenQuote(ctx)
	case "adventurous":
		contentType = "adventurous"
		imageQuery = "adventure"
		body, fetchErr = fetchers.FetchFact(ctx)
	case "thoughtful":
		contentType = "thoughtful"
		imageQuery = "meditation"
		body, fetchErr = fetchers.FetchZenQuote(ctx)
	default:
		body = "I don't know that mood yet."
		contentType = "unknown"
		imageQuery = ""
	}

	if fetchErr != nil || body == "" {
		body = "Sorry, couldn't fetch content right now. Try again."
	}

	// Translate content before sending  
	userLang := mh.languageManager.GetUserLanguage(chatID)
	translatedBody, _ := mh.translator.TranslateText(ctx, body, userLang)
	translatedBody = translation.FormatLanguageSpecificText(translatedBody, userLang)
	
	mh.sendContentWithImage(chatID, translatedBody, imageQuery, contentType)
	
	// Re-show mood keyboard
	_ = mh.SendMoodKeyboard(chatID)
}

// HandleSurprise handles the /surprise command
func (mh *MessageHandler) HandleSurprise(chatID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Random categories with their corresponding content fetchers and image queries
	categories := []models.ContentCategory{
		{Name: "funny", ImageQuery: "funny"},
		{Name: "inspiring", ImageQuery: "inspiration"},
		{Name: "educational", ImageQuery: "books"},
		{Name: "relaxing", ImageQuery: "nature"},
		{Name: "adventurous", ImageQuery: "adventure"},
		{Name: "thoughtful", ImageQuery: "meditation"},
	}

	// Pick random category
	randomIndex := int(time.Now().UnixNano()) % len(categories)
	chosen := categories[randomIndex]

	var body string
	var fetchErr error

	// Fetch content based on category
	switch chosen.Name {
	case "funny":
		body, fetchErr = fetchers.FetchJoke(ctx)
	case "inspiring", "relaxing", "thoughtful":
		body, fetchErr = fetchers.FetchZenQuote(ctx)
	case "educational", "adventurous":
		body, fetchErr = fetchers.FetchFact(ctx)
	default:
		body, fetchErr = fetchers.FetchFact(ctx)
	}

	// Fallback to fact if primary fetch fails
	if fetchErr != nil || body == "" {
		body, fetchErr = fetchers.FetchFact(ctx)
		chosen.ImageQuery = "random"
	}

	if fetchErr != nil || body == "" {
		body = "🎲 Surprise! Something unexpected happened - I couldn't fetch content right now. Try again!"
	}

	// Translate content before sending
	userLang := mh.languageManager.GetUserLanguage(chatID)
	translatedBody, _ := mh.translator.TranslateText(ctx, body, userLang)
	surpriseText, _ := mh.translator.TranslateText(ctx, "🎲 **SURPRISE!** ", userLang)
	
	mh.sendContentWithImage(chatID, surpriseText+translatedBody, chosen.ImageQuery, "surprise")
}