# 🤖 MoodBot

A Telegram bot that delivers personalized content based on your current mood. Get inspirational quotes, funny jokes, interesting facts, and beautiful images tailored to how you're feeling!

## ✨ Features

- **Mood-based Content**: Choose from different moods (Fun, Inspiring, Motivating, Casual) to get relevant content
- **Multiple Content Types**: 
  - Inspirational quotes from ZenQuotes API
  - Funny jokes from Official Joke API  
  - Interesting facts from Useless Facts API
  - Beautiful images from Unsplash API
- **Interactive Voting**: Rate content with thumbs up/down to help improve recommendations
- **Surprise Mode**: Get random content when you're feeling adventurous
- **Clean Interface**: Simple inline keyboard navigation

## 🚀 Commands

- `/start` - Display the main mood selection menu
- `/surprise` - Get random content from any category
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

3. Set your Telegram Bot Token:
   ```bash
   export TELEGRAM_BOT_TOKEN="your_bot_token_here"
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
└── go.mod              # Go module dependencies
```

## 🔧 Architecture

The bot is built with a modular architecture:

- **Main Loop**: Handles Telegram updates and routes commands/callbacks
- **Message Handlers**: Process user interactions and send appropriate responses  
- **Vote Manager**: Tracks user feedback on content
- **API Fetchers**: Retrieve content from external APIs
- **Models**: Define data structures for quotes, jokes, facts, and images

## 🌐 External APIs Used

- **ZenQuotes API**: Inspirational quotes
- **Official Joke API**: Clean, family-friendly jokes
- **Useless Facts API**: Interesting random facts
- **Unsplash API**: High-quality stock photos

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the terms specified in the LICENSE file.

## 🐛 Support

If you encounter any issues or have questions, please open an issue on GitHub.

---

Made with ❤️ for spreading good vibes through technology