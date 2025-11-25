package models

import (
	"encoding/json"
	"time"
)

type Channel struct {
	ID         int       `json:"id" db:"id"`
	TelegramID int64     `json:"telegram_id" db:"telegram_id"`
	Username   string    `json:"username" db:"username"`
	Title      string    `json:"title" db:"title"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	Timezone   string    `json:"timezone" db:"timezone"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type Post struct {
	ID           int             `json:"id" db:"id"`
	Content      string          `json:"content" db:"content"`
	MediaType    string          `json:"media_type" db:"media_type"`
	MediaPath    string          `json:"media_path" db:"media_path"`
	Buttons      json.RawMessage `json:"buttons" db:"buttons"`
	ScheduleTime *time.Time      `json:"schedule_time" db:"schedule_time"`
	Status       string          `json:"status" db:"status"`
	CreatedBy    string          `json:"created_by" db:"created_by"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	SentAt       *time.Time      `json:"sent_at" db:"sent_at"`
}

type PostChannel struct {
	ID        int       `json:"id" db:"id"`
	PostID    int       `json:"post_id" db:"post_id"`
	ChannelID int       `json:"channel_id" db:"channel_id"`
	MessageID int       `json:"message_id" db:"message_id"`
	Status    string    `json:"status" db:"status"`
	Error     string    `json:"error" db:"error"`
	SentAt    time.Time `json:"sent_at" db:"sent_at"`
}

type Button struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type Statistics struct {
	TotalPosts     int `json:"total_posts"`
	Successful     int `json:"successful"`
	Failed         int `json:"failed"`
	Scheduled      int `json:"scheduled"`
	TotalChannels  int `json:"total_channels"`
	ActiveChannels int `json:"active_channels"`
}
