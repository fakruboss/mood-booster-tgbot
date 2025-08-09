# 🤖 MoodBot

A Telegram bot that delivers personalized content based on your current mood. Get inspirational quotes, funny jokes, interesting facts, and beautiful images tailored to how you're feeling!

## ✨ Features

- **Mood-based Content**: Choose from different moods (Fun, Inspiring, Motivating, Casual) to get relevant content
- **Multi-language Support**: Content available in English, Hindi (हिंदी), and Tamil (தமிழ்) with automatic translation
- **Multiple Content Types**: 
  - Inspirational quotes from ZenQuotes API
  - Funny jokes from Official Joke API  
  - Interesting facts from Useless Facts API
  - Beautiful images from Unsplash API
- **Interactive Voting**: Rate content with thumbs up/down to help improve recommendations
- **Personal Favorites**: Save content you love with the ⭐ button and access them anytime with `/favorites`
- **Surprise Mode**: Get random content when you're feeling adventurous
- **Clean Interface**: Simple inline keyboard navigation

## 🚀 Commands

- `/start` - Display the main mood selection menu
- `/surprise` - Get random content from any category
- `/favorites` - View and manage your saved favorites
- `/language` - Change your language preference (shows available languages excluding current)
- `/help` - Show available commands and usage instructions

## 🛠️ Setup

### Prerequisites

- Go 1.24.3 or higher
- A Telegram Bot Token (get one from [@BotFather](https://t.me/BotFather))
- API keys for external services (optional, for enhanced functionality)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/you/moodbot.git
   cd moodbot
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set your environment variables:
   ```bash
   export TELEGRAM_BOT_TOKEN="your_bot_token_here"
   # Optional: For images  
   export UNSPLASH_ACCESS_KEY="your_unsplash_access_key"
   ```

4. Build and run:
   ```bash
   go build -o moodbot
   ./moodbot
   ```

## 📁 Project Structure

```
moodbot/
├── main.go              # Bot entry point and update handling
├── handlers/            # Message and callback handlers
│   └── message_handlers.go
├── models/              # Data structures and types
│   └── models.go
├── fetchers/            # External API clients
│   └── api_clients.go
├── voting/              # Vote management system
│   └── vote_manager.go
├── favorites/           # Favorite content management
│   └── favorite_manager.go
├── translation/         # Multi-language support
│   ├── translator.go   
│   └── language_manager.go
└── go.mod              # Go module dependencies
```

## 🔧 Architecture

The bot is built with a modular architecture:

- **Main Loop**: Handles Telegram updates and routes commands/callbacks
- **Message Handlers**: Process user interactions and send appropriate responses  
- **Vote Manager**: Tracks user feedback on content
- **Favorite Manager**: Stores and retrieves user's saved content
- **Translation System**: Provides multi-language support with user preferences
- **API Fetchers**: Retrieve content from external APIs
- **Models**: Define data structures for quotes, jokes, facts, images, and favorites

## 🌐 External APIs Used

- **ZenQuotes API**: Inspirational quotes
- **Official Joke API**: Clean, family-friendly jokes
- **Useless Facts API**: Interesting random facts
- **Unsplash API**: High-quality stock photos
- **MyMemory Translation API**: Free multi-language translation support

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the terms specified in the LICENSE file.

## 🐛 Support

If you encounter any issues or have questions, please open an issue on GitHub.

---

Made with ❤️ for spreading good vibes through technology