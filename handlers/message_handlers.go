package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/you/moodbot/favorites"
	"github.com/you/moodbot/fetchers"
	"github.com/you/moodbot/models"
	"github.com/you/moodbot/voting"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageHandler handles all bot message interactions
type MessageHandler struct {
	bot             *tgbotapi.BotAPI
	voteManager     *voting.VoteManager
	favoriteManager *favorites.FavoriteManager
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(bot *tgbotapi.BotAPI, voteManager *voting.VoteManager, favoriteManager *favorites.FavoriteManager) *MessageHandler {
	return &MessageHandler{
		bot:             bot,
		voteManager:     voteManager,
		favoriteManager: favoriteManager,
	}
}

// SendMoodKeyboard sends the main mood selection keyboard
func (mh *MessageHandler) SendMoodKeyboard(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "What's your mood today?")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Fun 🎉", "fun"),
			tgbotapi.NewInlineKeyboardButtonData("Inspiring 💡", "inspiring"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Motivating 💪", "motivating"),
			tgbotapi.NewInlineKeyboardButtonData("Casual 😌", "casual"),
		),
	)
	msg.ReplyMarkup = keyboard
	_, err := mh.bot.Send(msg)
	return err
}

// HandleSurprise handles the /surprise command
func (mh *MessageHandler) HandleSurprise(chatID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Random categories with their corresponding content fetchers and image queries
	categories := []models.ContentCategory{
		{Name: "fun", ImageQuery: "funny"},
		{Name: "inspiring", ImageQuery: "inspiration"},
		{Name: "motivating", ImageQuery: "motivation"},
		{Name: "casual", ImageQuery: "nature"},
		{Name: "random_fact", ImageQuery: "surprise"},
	}

	// Pick random category
	randomIndex := int(time.Now().UnixNano()) % len(categories)
	chosen := categories[randomIndex]

	var body string
	var fetchErr error

	// Fetch content based on category
	switch chosen.Name {
	case "fun":
		body, fetchErr = fetchers.FetchJoke(ctx)
	case "inspiring", "motivating":
		body, fetchErr = fetchers.FetchZenQuote(ctx)
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

	mh.sendContentWithImage(chatID, "🎲 **SURPRISE!** "+body, chosen.ImageQuery, "surprise")
}

// SendHelpMessage sends the help message with available commands
func (mh *MessageHandler) SendHelpMessage(chatID int64) {
	helpText := `🤖 **MoodBot Commands:**

/start - Choose your mood and get personalized content
/surprise - Get completely random content (jokes, quotes, or facts)
/favorites - View and manage your saved favorites
/help - Show this help message

**How it works:**
• Use /start to select from mood categories (Fun, Inspiring, Motivating, Casual)
• Use /surprise for random content from any category
• Vote on content with 👍 or 👎 buttons
• Save content you love with the ⭐ favorite button
• Each mood comes with matching images from Unsplash

Enjoy your mood-boosting content! 🎉`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"
	mh.bot.Send(msg)
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
	case "fun":
		contentType = "fun"
		imageQuery = "funny"
		// Alternate between joke and fun-fact
		body, fetchErr = fetchers.FetchJoke(ctx)
		if fetchErr != nil {
			body, fetchErr = fetchers.FetchFact(ctx)
		}
	case "inspiring":
		contentType = "inspiring"
		imageQuery = "inspiration"
		body, fetchErr = fetchers.FetchZenQuote(ctx)
	case "motivating":
		contentType = "motivating"
		imageQuery = "motivation"
		body, fetchErr = fetchers.FetchZenQuote(ctx)
	case "casual":
		contentType = "casual"
		imageQuery = "nature"
		body, fetchErr = fetchers.FetchFact(ctx)
	default:
		body = "I don't know that mood yet."
		contentType = "unknown"
		imageQuery = ""
	}

	if fetchErr != nil || body == "" {
		body = "Sorry, couldn't fetch content right now. Try again."
	}

	mh.sendContentWithImage(chatID, body, imageQuery, contentType)
	
	// Re-show mood keyboard
	_ = mh.SendMoodKeyboard(chatID)
}

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