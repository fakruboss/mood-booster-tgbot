package models

// ZenQuote represents a quote from ZenQuotes API
type ZenQuote struct {
	Q string `json:"q"`
	A string `json:"a"`
}

// Joke represents a joke from Official Joke API
type Joke struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Setup     string `json:"setup"`
	Punchline string `json:"punchline"`
}

// Fact represents a fact from Useless Facts API
type Fact struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// UnsplashImage represents an image from Unsplash API
type UnsplashImage struct {
	Urls struct {
		Regular string `json:"regular"`
		Small   string `json:"small"`
	} `json:"urls"`
	AltDescription string `json:"alt_description"`
}

// Vote represents a user's vote on content
type Vote struct {
	MessageID int
	UserID    int64
	Vote      bool // true = thumbs up, false = thumbs down
}

// ContentCategory represents different content types with their fetch functions
type ContentCategory struct {
	Name       string
	ImageQuery string
}