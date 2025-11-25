package scheduler

import (
	"log"
	"time"

	"github.com/maksekak/channelBot/cmd/internal/models"
	"github.com/maksekak/channelBot/cmd/internal/storage"
	"github.com/maksekak/channelBot/cmd/internal/telegram"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	storage  storage.Storage
	telegram *telegram.Client
	cron     *cron.Cron
}

func NewScheduler(storage storage.Storage, telegram *telegram.Client) *Scheduler {
	return &Scheduler{
		storage:  storage,
		telegram: telegram,
		cron:     cron.New(),
	}
}

func (s *Scheduler) Start() {
	// Проверка запланированных постов каждую минуту
	s.cron.AddFunc("* * * * *", s.processScheduledPosts)
	s.cron.Start()
	log.Println("Scheduler started")
}

func (s *Scheduler) processScheduledPosts() {
	now := time.Now()
	posts, err := s.storage.GetScheduledPosts(now)
	if err != nil {
		log.Printf("Error getting scheduled posts: %v", err)
		return
	}

	for _, post := range posts {
		log.Printf("Processing scheduled post ID: %d", post.ID)
		s.sendPost(post)

		// Обновление статуса поста
		post.Status = "sent"
		now := time.Now()
		post.SentAt = &now
		s.storage.UpdatePost(&post)
	}
}

func (s *Scheduler) sendPost(post models.Post) {
	channels, err := s.storage.GetActiveChannels()
	if err != nil {
		log.Printf("Error getting active channels: %v", err)
		return
	}

	for _, channel := range channels {
		messageID, err := s.telegram.SendMessage(channel.TelegramID, post)
		status := "sent"
		errorMsg := ""
		if err != nil {
			status = "error"
			errorMsg = err.Error()
			log.Printf("Error sending to channel %s: %v", channel.Username, err)
		}

		postChannel := models.PostChannel{
			PostID:    post.ID,
			ChannelID: channel.ID,
			MessageID: messageID,
			Status:    status,
			Error:     errorMsg,
			SentAt:    time.Now(),
		}
		if err := s.storage.CreatePostChannel(&postChannel); err != nil {
			log.Printf("Error saving post channel: %v", err)
		}
	}
}
