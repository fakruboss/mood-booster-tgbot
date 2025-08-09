package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/you/moodbot/models"
)

// ZenQuotesFetcher handles fetching quotes from ZenQuotes API
func FetchZenQuote(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://zenquotes.io/api/random", nil)
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var q []models.ZenQuote
	if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
		return "", err
	}
	if len(q) == 0 {
		return "", fmt.Errorf("no quote")
	}
	return fmt.Sprintf("\"%s\" — %s", q[0].Q, q[0].A), nil
}

// JokeFetcher handles fetching jokes from Official Joke API
func FetchJoke(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://official-joke-api.appspot.com/jokes/random", nil)
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var j models.Joke
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n\n%s", j.Setup, j.Punchline), nil
}

// FactFetcher handles fetching facts from Useless Facts API
func FetchFact(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://uselessfacts.jsph.pl/api/v2/facts/random?language=en", nil)
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var f models.Fact
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		return "", err
	}
	
	// Some responses use "text" field, so fallback:
	if f.Text == "" {
		// Try generic map
		var m map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&m); err == nil {
			if t, ok := m["text"].(string); ok {
				f.Text = t
			}
		}
	}
	return f.Text, nil
}

// UnsplashFetcher handles fetching images from Unsplash API
func FetchUnsplashImage(ctx context.Context, query string) (string, error) {
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

	var img models.UnsplashImage
	if err := json.NewDecoder(resp.Body).Decode(&img); err != nil {
		return "", err
	}
	return img.Urls.Small, nil
}