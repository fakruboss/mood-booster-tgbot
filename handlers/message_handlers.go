package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/you/moodbot/favorites"
	"github.com/you/moodbot/translation"
	"github.com/you/moodbot/voting"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageHandler handles all bot message interactions
type MessageHandler struct {
	bot             *tgbotapi.BotAPI
	voteManager     *voting.VoteManager
	favoriteManager *favorites.FavoriteManager
	translator      *translation.Translator
	languageManager *translation.LanguageManager
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(bot *tgbotapi.BotAPI, voteManager *voting.VoteManager, favoriteManager *favorites.FavoriteManager, translator *translation.Translator, languageManager *translation.LanguageManager) *MessageHandler {
	return &MessageHandler{
		bot:             bot,
		voteManager:     voteManager,
		favoriteManager: favoriteManager,
		translator:      translator,
		languageManager: languageManager,
	}
}

// SendHelpMessage sends the help message with available commands
func (mh *MessageHandler) SendHelpMessage(chatID int64) {
	userLang := mh.languageManager.GetUserLanguage(chatID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	helpText := `🤖 **MoodBot Commands:**

/start - Choose your mood and get personalized content
/surprise - Get completely random content (jokes, quotes, or facts)
/favorites - View and manage your saved favorites
/language - Change your language preference
/help - Show this help message

**How it works:**
• Use /start to select from mood categories (Fun, Inspiring, Motivating, Casual)
• Use /surprise for random content from any category
• Vote on content with 👍 or 👎 buttons
• Save content you love with the ⭐ favorite button
• Each mood comes with matching images from Unsplash

Enjoy your mood-boosting content! 🎉`

	translatedHelp, _ := mh.translator.TranslateText(ctx, helpText, userLang)
	
	msg := tgbotapi.NewMessage(chatID, translatedHelp)
	msg.ParseMode = "Markdown"
	mh.bot.Send(msg)
}

// SendFavorites displays user's saved favorites
func (mh *MessageHandler) SendFavorites(chatID int64, userID int64) {
	favorites := mh.favoriteManager.GetUserFavorites(userID)
	
	if len(favorites) == 0 {
		msg := tgbotapi.NewMessage(chatID, "⭐ You haven't saved any favorites yet!\n\nUse the ⭐ button on content you like to save it here.")
		mh.bot.Send(msg)
		return
	}
	
	headerMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⭐ **Your Favorites** (%d saved)", len(favorites)))
	headerMsg.ParseMode = "Markdown"
	mh.bot.Send(headerMsg)
	
	for i, fav := range favorites {
		var text string
		var keyboard tgbotapi.InlineKeyboardMarkup
		
		switch fav.Type {
		case "quote":
			text = fmt.Sprintf("💡 *Quote #%d*\n\n%s\n\n*— %s*", i+1, fav.Content, fav.Author)
		case "joke":
			text = fmt.Sprintf("😄 *Joke #%d*\n\n%s\n\n%s", i+1, fav.Setup, fav.Punchline)
		case "fact":
			text = fmt.Sprintf("🧠 *Fact #%d*\n\n%s", i+1, fav.Content)
		case "image":
			text = fmt.Sprintf("🖼️ *Image #%d*\n\n%s", i+1, fav.Content)
		default:
			text = fmt.Sprintf("📄 *Content #%d*\n\n%s", i+1, fav.Content)
		}
		
		// Add remove button
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🗑️ Remove", fmt.Sprintf("favorite_remove_%s", fav.ID)),
			),
		)
		
		if fav.ImageURL != "" && fav.Type == "image" {
			// Send as photo if it has an image
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(fav.ImageURL))
			photo.Caption = text
			photo.ParseMode = "Markdown"
			photo.ReplyMarkup = keyboard
			_, err := mh.bot.Send(photo)
			if err != nil {
				// Fallback to text if photo fails
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = "Markdown"
				msg.ReplyMarkup = keyboard
				mh.bot.Send(msg)
			}
		} else {
			// Send as text message
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = keyboard
			mh.bot.Send(msg)
		}
	}
}

// HandleFavoriteCallback processes favorite-related callbacks
func (mh *MessageHandler) HandleFavoriteCallback(data string, userID int64, originalContent string) string {
	// For adding favorites, we need to parse the original content
	// This is a simplified implementation - in a real app, you'd store more context
	if data == "favorite_add" {
		// Extract content type and details from the original message
		contentType := "general"
		content := originalContent
		author := ""
		setup := ""
		punchline := ""
		imageURL := ""
		
		// Try to detect content type from format
		if strings.Contains(content, "—") {
			contentType = "quote"
			parts := strings.Split(content, "—")
			if len(parts) >= 2 {
				content = strings.TrimSpace(parts[0])
				author = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(content, "\n\n") && !strings.Contains(content, "SURPRISE") {
			contentType = "joke"
			parts := strings.Split(content, "\n\n")
			if len(parts) >= 2 {
				setup = strings.TrimSpace(parts[0])
				punchline = strings.TrimSpace(parts[1])
				content = fmt.Sprintf("%s %s", setup, punchline)
			}
		} else {
			contentType = "fact"
		}
		
		return mh.favoriteManager.HandleFavoriteCallback(data, userID, contentType, content, author, setup, punchline, imageURL)
	}
	
	return mh.favoriteManager.HandleFavoriteCallback(data, userID, "", "", "", "", "", "")
}