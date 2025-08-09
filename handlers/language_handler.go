package handlers

import (
	"fmt"

	"github.com/you/moodbot/translation"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendLanguageKeyboard sends language selection keyboard (excluding current language)
func (mh *MessageHandler) SendLanguageKeyboard(chatID int64) {
	currentLang := mh.languageManager.GetUserLanguage(chatID)
	
	// Create keyboard with available languages (excluding current)
	var buttons [][]tgbotapi.InlineKeyboardButton
	
	if currentLang != translation.English {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("English 🇺🇸", "lang_en"),
		})
	}
	
	if currentLang != translation.Hindi {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("हिंदी 🇮🇳", "lang_hi"),
		})
	}
	
	if currentLang != translation.Tamil {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("தமிழ் 🇮🇳", "lang_ta"),
		})
	}
	
	// If somehow no other languages available, show current status
	if len(buttons) == 0 {
		var currentLangName string
		switch currentLang {
		case translation.English:
			currentLangName = "English 🇺🇸"
		case translation.Hindi:
			currentLangName = "हिंदी 🇮🇳"
		case translation.Tamil:
			currentLangName = "தமிழ் 🇮🇳"
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Current language: %s", currentLangName))
		mh.bot.Send(msg)
		return
	}
	
	msg := tgbotapi.NewMessage(chatID, "🌍 Choose your preferred language / अपनी भाषा चुनें / உங்கள் மொழியைத் தேர்ந்தெடுக்கவும்:")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	msg.ReplyMarkup = keyboard
	mh.bot.Send(msg)
}

// HandleLanguageSelection handles language selection callbacks
func (mh *MessageHandler) HandleLanguageSelection(data string, chatID, userID int64) string {
	var selectedLang translation.Language
	var response string
	
	switch data {
	case "lang_en":
		selectedLang = translation.English
		response = "✅ Language set to English!"
	case "lang_hi":
		selectedLang = translation.Hindi
		response = "✅ भाषा हिंदी में सेट की गई!"
	case "lang_ta":
		selectedLang = translation.Tamil
		response = "✅ மொழி தமிழில் அமைக்கப்பட்டது!"
	default:
		return "Unknown language selection"
	}
	
	err := mh.languageManager.SetUserLanguage(userID, selectedLang)
	if err != nil {
		return "Error saving language preference"
	}
	
	return response
}