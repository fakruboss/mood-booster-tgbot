package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Language represents supported languages
type Language string

const (
	English Language = "en"
	Hindi   Language = "hi"
	Tamil   Language = "ta"
)

// MyMemoryResponse represents MyMemory Translation API response
type MyMemoryResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
	ResponseStatus int `json:"responseStatus"`
}

// Translator handles translation operations
type Translator struct {
	client *http.Client
}

// NewTranslator creates a new translator instance
func NewTranslator() *Translator {
	return &Translator{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// TranslateText translates text to the specified language using MyMemory API
func (t *Translator) TranslateText(ctx context.Context, text string, targetLang Language) (string, error) {
	// If target language is English, return original text
	if targetLang == English {
		return text, nil
	}

	// MyMemory Translation API endpoint (free, no API key required)
	// Limit text length to avoid issues with very long content
	if len(text) > 500 {
		text = text[:500] + "..."
	}

	// Build API URL with parameters
	apiURL := fmt.Sprintf("https://api.mymemory.translated.net/get?q=%s&langpair=%s|%s",
		url.QueryEscape(text), "en", string(targetLang))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return text, err // Fallback to original text
	}

	// Make request
	resp, err := t.client.Do(req)
	if err != nil {
		return text, err // Fallback to original text
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return text, fmt.Errorf("translation API error: %d", resp.StatusCode)
	}

	// Parse response
	var translationResp MyMemoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&translationResp); err != nil {
		return text, err // Fallback to original text
	}

	// Check if translation was successful (status 200 = OK)
	if translationResp.ResponseStatus != 200 {
		return text, fmt.Errorf("translation failed with status: %d", translationResp.ResponseStatus)
	}

	translatedText := translationResp.ResponseData.TranslatedText
	if translatedText == "" {
		return text, fmt.Errorf("empty translation returned")
	}

	return translatedText, nil
}

// GetLanguageFromCommand determines language from user command or preference
func GetLanguageFromCommand(command string) Language {
	command = strings.ToLower(command)
	
	switch {
	case strings.Contains(command, "tamil") || strings.Contains(command, "ta"):
		return Tamil
	case strings.Contains(command, "hindi") || strings.Contains(command, "hi"):
		return Hindi
	default:
		return English
	}
}

// FormatLanguageSpecificText formats text based on language requirements
func FormatLanguageSpecificText(text string, lang Language) string {
	switch lang {
	case Tamil:
		// Add Tamil-specific formatting if needed
		return "🇮🇳 " + text
	case Hindi:
		// Add Hindi-specific formatting if needed
		return "🇮🇳 " + text
	default:
		return text
	}
}

// IsTranslationSupported checks if translation is available
func (t *Translator) IsTranslationSupported() bool {
	return true // MyMemory API is always available (no API key needed)
}