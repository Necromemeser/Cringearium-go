package domain

import "time"

type Conversation struct {
	ID        int64
	UserID    int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Message struct {
	ID             int64
	ConversationID int64
	UserID         *int64
	Content        string
	IsAIResponse   bool
	CreatedAt      time.Time
}

type AdminStats struct {
	Conversations int64
	Messages      int64
	UserMessages  int64
	AIMessages    int64
	ActiveUsers   int64
}
