package translation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// UserLanguagePreference stores user language preferences
type UserLanguagePreference struct {
	UserID   int64    `json:"user_id"`
	Language Language `json:"language"`
}

// LanguageManager manages user language preferences
type LanguageManager struct {
	preferences map[int64]Language
	mutex       sync.RWMutex
	dataDir     string
}

// NewLanguageManager creates a new language manager
func NewLanguageManager(dataDir string) *LanguageManager {
	lm := &LanguageManager{
		preferences: make(map[int64]Language),
		dataDir:     dataDir,
	}
	
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create language data directory: %v\n", err)
	}
	
	// Load existing preferences
	lm.loadPreferences()
	
	return lm
}

// SetUserLanguage sets the language preference for a user
func (lm *LanguageManager) SetUserLanguage(userID int64, language Language) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	lm.preferences[userID] = language
	return lm.savePreferences()
}

// GetUserLanguage gets the language preference for a user (defaults to English)
func (lm *LanguageManager) GetUserLanguage(userID int64) Language {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()
	
	if lang, exists := lm.preferences[userID]; exists {
		return lang
	}
	return English // Default to English
}

// GetSupportedLanguages returns list of supported languages with their display names
func (lm *LanguageManager) GetSupportedLanguages() map[Language]string {
	return map[Language]string{
		English: "English 🇺🇸",
		Hindi:   "हिंदी 🇮🇳",
		Tamil:   "தமிழ் 🇮🇳",
	}
}

// loadPreferences loads user preferences from file
func (lm *LanguageManager) loadPreferences() {
	filePath := filepath.Join(lm.dataDir, "language_preferences.json")
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		// File doesn't exist, that's okay
		return
	}
	
	var prefs []UserLanguagePreference
	if err := json.Unmarshal(data, &prefs); err != nil {
		fmt.Printf("Warning: Could not load language preferences: %v\n", err)
		return
	}
	
	for _, pref := range prefs {
		lm.preferences[pref.UserID] = pref.Language
	}
}

// savePreferences saves user preferences to file
func (lm *LanguageManager) savePreferences() error {
	filePath := filepath.Join(lm.dataDir, "language_preferences.json")
	
	var prefs []UserLanguagePreference
	for userID, lang := range lm.preferences {
		prefs = append(prefs, UserLanguagePreference{
			UserID:   userID,
			Language: lang,
		})
	}
	
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filePath, data, 0644)
}