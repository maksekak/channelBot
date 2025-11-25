package storage

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/maksekak/channelBot/internal/models"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) GetChannels() ([]models.Channel, error) {
	query := `SELECT id, telegram_id, username, title, is_active, timezone, created_at FROM channels ORDER BY created_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []models.Channel
	for rows.Next() {
		var channel models.Channel
		err := rows.Scan(
			&channel.ID,
			&channel.TelegramID,
			&channel.Username,
			&channel.Title,
			&channel.IsActive,
			&channel.Timezone,
			&channel.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

func (s *PostgresStorage) CreateChannel(channel *models.Channel) error {
	query := `INSERT INTO channels (telegram_id, username, title, is_active, timezone, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return s.db.QueryRow(query,
		channel.TelegramID,
		channel.Username,
		channel.Title,
		channel.IsActive,
		channel.Timezone,
		time.Now(),
	).Scan(&channel.ID)
}

func (s *PostgresStorage) CreatePost(post *models.Post) error {
	query := `INSERT INTO posts (content, media_type, media_path, buttons, schedule_time, status, created_by, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	return s.db.QueryRow(query,
		post.Content,
		post.MediaType,
		post.MediaPath,
		post.Buttons,
		post.ScheduleTime,
		post.Status,
		post.CreatedBy,
		time.Now(),
	).Scan(&post.ID)
}

func (s *PostgresStorage) GetScheduledPosts(before time.Time) ([]models.Post, error) {
	query := `SELECT id, content, media_type, media_path, buttons, schedule_time, status, created_by, created_at 
              FROM posts WHERE schedule_time <= $1 AND status = 'scheduled'`

	rows, err := s.db.Query(query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.Content,
			&post.MediaType,
			&post.MediaPath,
			&post.Buttons,
			&post.ScheduleTime,
			&post.Status,
			&post.CreatedBy,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *PostgresStorage) CreatePostChannel(pc *models.PostChannel) error {
	query := `INSERT INTO post_channels (post_id, channel_id, message_id, status, error, sent_at) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return s.db.QueryRow(query,
		pc.PostID,
		pc.ChannelID,
		pc.MessageID,
		pc.Status,
		pc.Error,
		time.Now(),
	).Scan(&pc.ID)
}

func (s *PostgresStorage) GetStatistics(days int) (*models.Statistics, error) {
	stats := &models.Statistics{}

	// Общее количество постов
	err := s.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&stats.TotalPosts)
	if err != nil {
		return nil, err
	}

	// Успешные отправки
	err = s.db.QueryRow("SELECT COUNT(*) FROM post_channels WHERE status = 'sent'").Scan(&stats.Successful)
	if err != nil {
		return nil, err
	}

	// Неудачные отправки
	err = s.db.QueryRow("SELECT COUNT(*) FROM post_channels WHERE status = 'error'").Scan(&stats.Failed)
	if err != nil {
		return nil, err
	}

	// Запланированные посты
	err = s.db.QueryRow("SELECT COUNT(*) FROM posts WHERE status = 'scheduled'").Scan(&stats.Scheduled)
	if err != nil {
		return nil, err
	}

	// Каналы
	err = s.db.QueryRow("SELECT COUNT(*) FROM channels").Scan(&stats.TotalChannels)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow("SELECT COUNT(*) FROM channels WHERE is_active = true").Scan(&stats.ActiveChannels)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
