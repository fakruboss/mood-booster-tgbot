package favorites

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/you/moodbot/models"
)

// FavoriteManager manages user favorites with file persistence
type FavoriteManager struct {
	favorites map[int64][]models.Favorite // userID -> favorites (cache)
	mutex     sync.RWMutex
	dataDir   string // directory to store user files
}

// NewFavoriteManager creates a new favorite manager with file persistence
func NewFavoriteManager() *FavoriteManager {
	dataDir := "user_data"
	
	// Create user_data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create user_data directory: %v\n", err)
	}
	
	return &FavoriteManager{
		favorites: make(map[int64][]models.Favorite),
		dataDir:   dataDir,
	}
}

// AddFavorite adds a favorite for a user
func (fm *FavoriteManager) AddFavorite(userID int64, contentType, content, author, setup, punchline, imageURL string) string {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// Load current favorites from file if not in cache
	if _, exists := fm.favorites[userID]; !exists {
		fm.favorites[userID] = fm.loadUserFavorites(userID)
	}

	// Generate unique ID
	id := generateID()
	
	favorite := models.Favorite{
		ID:        id,
		UserID:    userID,
		Type:      contentType,
		Content:   content,
		Author:    author,
		Setup:     setup,
		Punchline: punchline,
		ImageURL:  imageURL,
		SavedAt:   time.Now().Unix(),
	}

	// Add to cache
	fm.favorites[userID] = append(fm.favorites[userID], favorite)
	
	// Save to file
	if err := fm.saveUserFavorites(userID, fm.favorites[userID]); err != nil {
		fmt.Printf("Error saving favorites for user %d: %v\n", userID, err)
	}
	
	return id
}

// GetUserFavorites returns all favorites for a user
func (fm *FavoriteManager) GetUserFavorites(userID int64) []models.Favorite {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	// Load from file if not in cache
	if _, exists := fm.favorites[userID]; !exists {
		fm.favorites[userID] = fm.loadUserFavorites(userID)
	}

	if favorites, exists := fm.favorites[userID]; exists {
		// Return a copy to avoid race conditions
		result := make([]models.Favorite, len(favorites))
		copy(result, favorites)
		return result
	}
	return []models.Favorite{}
}

// RemoveFavorite removes a favorite by ID for a user
func (fm *FavoriteManager) RemoveFavorite(userID int64, favoriteID string) bool {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// Load current favorites from file if not in cache
	if _, exists := fm.favorites[userID]; !exists {
		fm.favorites[userID] = fm.loadUserFavorites(userID)
	}

	if favorites, exists := fm.favorites[userID]; exists {
		for i, fav := range favorites {
			if fav.ID == favoriteID {
				// Remove the favorite
				fm.favorites[userID] = append(favorites[:i], favorites[i+1:]...)
				
				// Save to file
				if err := fm.saveUserFavorites(userID, fm.favorites[userID]); err != nil {
					fmt.Printf("Error saving favorites after removal for user %d: %v\n", userID, err)
				}
				
				return true
			}
		}
	}
	return false
}

// GetFavoriteCount returns the number of favorites for a user
func (fm *FavoriteManager) GetFavoriteCount(userID int64) int {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	// Load from file if not in cache
	if _, exists := fm.favorites[userID]; !exists {
		fm.favorites[userID] = fm.loadUserFavorites(userID)
	}

	if favorites, exists := fm.favorites[userID]; exists {
		return len(favorites)
	}
	return 0
}

// IsFavoriteCallback checks if the callback data is for favorite operations
func (fm *FavoriteManager) IsFavoriteCallback(data string) bool {
	return len(data) > 9 && data[:9] == "favorite_"
}

// HandleFavoriteCallback processes favorite-related callbacks
func (fm *FavoriteManager) HandleFavoriteCallback(data string, userID int64, contentType, content, author, setup, punchline, imageURL string) string {
	if data == "favorite_add" {
		id := fm.AddFavorite(userID, contentType, content, author, setup, punchline, imageURL)
		return fmt.Sprintf("Added to favorites! (ID: %s)", id)
	}
	
	if len(data) > 16 && data[:16] == "favorite_remove_" {
		favoriteID := data[16:]
		if fm.RemoveFavorite(userID, favoriteID) {
			return "Removed from favorites!"
		}
		return "Favorite not found."
	}
	
	return "Unknown favorite action."
}

// generateID creates a random ID for favorites
func generateID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getUserFilePath returns the file path for a user's favorites
func (fm *FavoriteManager) getUserFilePath(userID int64) string {
	return filepath.Join(fm.dataDir, fmt.Sprintf("user_%d_favorites.json", userID))
}

// loadUserFavorites loads favorites from file for a specific user
func (fm *FavoriteManager) loadUserFavorites(userID int64) []models.Favorite {
	filePath := fm.getUserFilePath(userID)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []models.Favorite{}
	}
	
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading favorites file for user %d: %v\n", userID, err)
		return []models.Favorite{}
	}
	
	// Parse JSON
	var favorites []models.Favorite
	if err := json.Unmarshal(data, &favorites); err != nil {
		fmt.Printf("Error parsing favorites file for user %d: %v\n", userID, err)
		return []models.Favorite{}
	}
	
	return favorites
}

// saveUserFavorites saves favorites to file for a specific user
func (fm *FavoriteManager) saveUserFavorites(userID int64, favorites []models.Favorite) error {
	filePath := fm.getUserFilePath(userID)
	
	// Convert to JSON
	data, err := json.MarshalIndent(favorites, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling favorites: %v", err)
	}
	
	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing favorites file: %v", err)
	}
	
	return nil
}