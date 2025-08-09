package voting

import (
	"fmt"
	"strings"
	"sync"

	"github.com/you/moodbot/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// VoteManager handles all voting-related functionality
type VoteManager struct {
	votes      map[string][]models.Vote
	votesMutex sync.RWMutex
}

// NewVoteManager creates a new vote manager instance
func NewVoteManager() *VoteManager {
	return &VoteManager{
		votes:      make(map[string][]models.Vote),
		votesMutex: sync.RWMutex{},
	}
}

// CreateVotingKeyboard creates inline keyboard with voting buttons
func (vm *VoteManager) CreateVotingKeyboard(contentType string, messageID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👍", fmt.Sprintf("vote_%s_%d_up", contentType, messageID)),
			tgbotapi.NewInlineKeyboardButtonData("👎", fmt.Sprintf("vote_%s_%d_down", contentType, messageID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐", "favorite_add"),
		),
	)
}

// HandleVote processes a vote from callback data
func (vm *VoteManager) HandleVote(data string, userID int64) string {
	// Parse vote data: vote_{type}_{messageID}_{up/down}
	parts := strings.Split(data, "_")
	if len(parts) < 4 {
		return "Invalid vote data"
	}

	contentType := parts[1]
	messageID := 0
	fmt.Sscanf(parts[2], "%d", &messageID)
	voteType := parts[3]

	vote := models.Vote{
		MessageID: messageID,
		UserID:    userID,
		Vote:      voteType == "up",
	}

	vm.votesMutex.Lock()
	key := fmt.Sprintf("%s_%d", contentType, messageID)

	// Remove existing vote from same user for this content
	existingVotes := vm.votes[key]
	filteredVotes := make([]models.Vote, 0)
	for _, v := range existingVotes {
		if v.UserID != userID {
			filteredVotes = append(filteredVotes, v)
		}
	}

	// Add new vote
	filteredVotes = append(filteredVotes, vote)
	vm.votes[key] = filteredVotes
	vm.votesMutex.Unlock()

	if voteType == "up" {
		return "Thanks for the thumbs up! 👍"
	}
	return "Thanks for the feedback! 👎"
}

// GetVoteStats returns voting statistics for analytics (future use)
func (vm *VoteManager) GetVoteStats(contentType string, messageID int) (int, int) {
	vm.votesMutex.RLock()
	defer vm.votesMutex.RUnlock()

	key := fmt.Sprintf("%s_%d", contentType, messageID)
	votes := vm.votes[key]

	upVotes, downVotes := 0, 0
	for _, vote := range votes {
		if vote.Vote {
			upVotes++
		} else {
			downVotes++
		}
	}

	return upVotes, downVotes
}

// IsVoteCallback checks if callback data is a vote
func (vm *VoteManager) IsVoteCallback(data string) bool {
	return len(data) > 5 && data[:5] == "vote_"
}