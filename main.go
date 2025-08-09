package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ZenQuote struct {
	Q string `json:"q"`
	A string `json:"a"`
}

type Joke struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Setup     string `json:"setup"`
	Punchline string `json:"punchline"`
}

type Fact struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type UnsplashImage struct {
	Urls struct {
		Regular string `json:"regular"`
		Small   string `json:"small"`
	} `json:"urls"`
	AltDescription string `json:"alt_description"`
}

type Vote struct {
	MessageID int
	UserID    int64
	Vote      bool // true = thumbs up, false = thumbs down
}

var (
	votes = make(map[string][]Vote)
	votesMutex sync.RWMutex
)

func fetchZenQuote(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://zenquotes.io/api/random", nil)
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	var q []ZenQuote
	if err := json.NewDecoder(resp.Body).Decode(&q); err != nil { return "", err }
	if len(q) == 0 { return "", fmt.Errorf("no quote") }
	return fmt.Sprintf("“%s” — %s", q[0].Q, q[0].A), nil
}

func fetchJoke(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://official-joke-api.appspot.com/jokes/random", nil)
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	var j Joke
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil { return "", err }
	return fmt.Sprintf("%s\n\n%s", j.Setup, j.Punchline), nil
}

func fetchFact(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://uselessfacts.jsph.pl/api/v2/facts/random?language=en", nil)
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	var f Fact
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil { return "", err }
	// some responses use "text" field, so fallback:
	if f.Text == "" {
		// try generic map
		var m map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&m); err == nil {
			if t, ok := m["text"].(string); ok { f.Text = t }
		}
	}
	return f.Text, nil
}

func fetchUnsplashImage(ctx context.Context, query string) (string, error) {
	accessKey := os.Getenv("UNSPLASH_ACCESS_KEY")
	if accessKey == "" {
		return "", fmt.Errorf("UNSPLASH_ACCESS_KEY not set")
	}

	url := fmt.Sprintf("https://api.unsplash.com/photos/random?query=%s&client_id=%s", query, accessKey)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var img UnsplashImage
	if err := json.NewDecoder(resp.Body).Decode(&img); err != nil {
		return "", err
	}
	return img.Urls.Small, nil
}

func createVotingKeyboard(contentType string, messageID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👍", fmt.Sprintf("vote_%s_%d_up", contentType, messageID)),
			tgbotapi.NewInlineKeyboardButtonData("👎", fmt.Sprintf("vote_%s_%d_down", contentType, messageID)),
		),
	)
}

func handleVote(data string, userID int64) string {
	// Parse vote data: vote_{type}_{messageID}_{up/down}
	parts := strings.Split(data, "_")
	if len(parts) < 4 {
		return "Invalid vote data"
	}
	
	contentType := parts[1]
	messageID := 0
	fmt.Sscanf(parts[2], "%d", &messageID)
	voteType := parts[3]

	vote := Vote{
		MessageID: messageID,
		UserID:    userID,
		Vote:      voteType == "up",
	}

	votesMutex.Lock()
	key := fmt.Sprintf("%s_%d", contentType, messageID)
	
	// Remove existing vote from same user for this content
	existingVotes := votes[key]
	filteredVotes := make([]Vote, 0)
	for _, v := range existingVotes {
		if v.UserID != userID {
			filteredVotes = append(filteredVotes, v)
		}
	}
	
	// Add new vote
	filteredVotes = append(filteredVotes, vote)
	votes[key] = filteredVotes
	votesMutex.Unlock()

	if voteType == "up" {
		return "Thanks for the thumbs up! 👍"
	}
	return "Thanks for the feedback! 👎"
}

func sendMoodKeyboard(bot *tgbotapi.BotAPI, chatID int64) error {
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
	_, err := bot.Send(msg)
	return err
}

func handleSurprise(bot *tgbotapi.BotAPI, chatID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Random categories with their corresponding content fetchers and image queries
	categories := []struct {
		name     string
		fetchFn  func(context.Context) (string, error)
		imageQuery string
	}{
		{"fun", fetchJoke, "funny"},
		{"inspiring", fetchZenQuote, "inspiration"},
		{"motivating", fetchZenQuote, "motivation"},
		{"casual", fetchFact, "nature"},
		{"random_fact", fetchFact, "surprise"},
	}

	// Pick random category
	randomIndex := int(time.Now().UnixNano()) % len(categories)
	chosen := categories[randomIndex]

	var body string
	var fetchErr error
	body, fetchErr = chosen.fetchFn(ctx)
	
	// Fallback to fact if primary fetch fails
	if fetchErr != nil || body == "" {
		body, fetchErr = fetchFact(ctx)
		chosen.imageQuery = "random"
	}

	if fetchErr != nil || body == "" {
		body = "🎲 Surprise! Something unexpected happened - I couldn't fetch content right now. Try again!"
	}

	imageURL, _ := fetchUnsplashImage(ctx, chosen.imageQuery)

	var sentMsg tgbotapi.Message
	surpriseText := "🎲 **SURPRISE!** " + body

	if imageURL != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(imageURL))
		photo.Caption = surpriseText
		photo.ParseMode = "Markdown"
		photo.ReplyMarkup = createVotingKeyboard("surprise", int(time.Now().Unix()))
		msg, err := bot.Send(photo)
		if err != nil {
			// Fallback to text message if photo fails
			textMsg := tgbotapi.NewMessage(chatID, surpriseText)
			textMsg.ParseMode = "Markdown"
			textMsg.ReplyMarkup = createVotingKeyboard("surprise", int(time.Now().Unix()))
			sentMsg, _ = bot.Send(textMsg)
		} else {
			sentMsg = msg
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, surpriseText)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = createVotingKeyboard("surprise", int(sentMsg.MessageID))
		sentMsg, _ = bot.Send(msg)
	}
}

func sendHelpMessage(bot *tgbotapi.BotAPI, chatID int64) {
	helpText := `🤖 **MoodBot Commands:**

/start - Choose your mood and get personalized content
/surprise - Get completely random content (jokes, quotes, or facts)
/help - Show this help message

**How it works:**
• Use /start to select from mood categories (Fun, Inspiring, Motivating, Casual)
• Use /surprise for random content from any category
• Vote on content with 👍 or 👎 buttons
• Each mood comes with matching images from Unsplash

Enjoy your mood-boosting content! 🎉`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					_ = sendMoodKeyboard(bot, update.Message.Chat.ID)
				case "surprise":
					handleSurprise(bot, update.Message.Chat.ID)
				case "help":
					sendHelpMessage(bot, update.Message.Chat.ID)
				default:
					sendHelpMessage(bot, update.Message.Chat.ID)
				}
			}
			continue
		}

		if update.CallbackQuery != nil {
			data := update.CallbackQuery.Data
			chatID := update.CallbackQuery.Message.Chat.ID
			userID := update.CallbackQuery.From.ID

			// Handle voting
			if len(data) > 5 && data[:5] == "vote_" {
				cb := tgbotapi.NewCallback(update.CallbackQuery.ID, handleVote(data, userID))
				_, _ = bot.Request(cb)
				continue
			}

			// acknowledge callback (remove "loading")
			cb := tgbotapi.NewCallback(update.CallbackQuery.ID, "Working on it...")
			_, _ = bot.Request(cb)

			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			var body string
			var fetchErr error
			var imageURL string
			var contentType string

			switch data {
			case "fun":
				contentType = "fun"
				// alternate between joke and fun-fact
				body, fetchErr = fetchJoke(ctx)
				if fetchErr != nil { 
					body, fetchErr = fetchFact(ctx) 
				}
				imageURL, _ = fetchUnsplashImage(ctx, "funny")
			case "inspiring":
				contentType = "inspiring"
				body, fetchErr = fetchZenQuote(ctx)
				imageURL, _ = fetchUnsplashImage(ctx, "inspiration")
			case "motivating":
				contentType = "motivating"
				body, fetchErr = fetchZenQuote(ctx)
				imageURL, _ = fetchUnsplashImage(ctx, "motivation")
			case "casual":
				contentType = "casual"
				body, fetchErr = fetchFact(ctx)
				imageURL, _ = fetchUnsplashImage(ctx, "nature")
			default:
				body = "I don't know that mood yet."
				contentType = "unknown"
			}
			cancel()

			if fetchErr != nil || body == "" {
				body = "Sorry, couldn't fetch content right now. Try again."
			}

			var sentMsg tgbotapi.Message
			if imageURL != "" {
				// Send photo with caption
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(imageURL))
				photo.Caption = body
				photo.ReplyMarkup = createVotingKeyboard(contentType, int(time.Now().Unix()))
				msg, err := bot.Send(photo)
				if err != nil {
					// Fallback to text message if photo fails
					textMsg := tgbotapi.NewMessage(chatID, body)
					textMsg.ReplyMarkup = createVotingKeyboard(contentType, int(time.Now().Unix()))
					sentMsg, _ = bot.Send(textMsg)
				} else {
					sentMsg = msg
				}
			} else {
				// Send text message
				msg := tgbotapi.NewMessage(chatID, body)
				msg.ReplyMarkup = createVotingKeyboard(contentType, int(sentMsg.MessageID))
				sentMsg, _ = bot.Send(msg)
			}

			// optionally re-show keyboard
			_ = sendMoodKeyboard(bot, chatID)
		}
	}
}