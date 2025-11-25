package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maksekak/channelBot/config"
	"github.com/maksekak/channelBot/internal/auth"
	"github.com/maksekak/channelBot/internal/models"
	"github.com/maksekak/channelBot/internal/storage"
	"github.com/maksekak/channelBot/internal/telegram"
)

type Handler struct {
	storage  storage.Storage
	telegram *telegram.Client
	auth     *auth.Auth
	config   *config.Config
}

func NewHandler(storage storage.Storage, telegram *telegram.Client, auth *auth.Auth, config *config.Config) *Handler {
	return &Handler{
		storage:  storage,
		telegram: telegram,
		auth:     auth,
		config:   config,
	}
}

func (h *Handler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (h *Handler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username != h.config.AdminUsername || password != h.config.AdminPassword {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Error": "Invalid credentials",
		})
		return
	}

	token, err := h.auth.GenerateToken(username)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Error": "Failed to create session",
		})
		return
	}

	h.auth.SetTokenCookie(c.Writer, token)
	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (h *Handler) Logout(c *gin.Context) {
	h.auth.ClearTokenCookie(c.Writer)
	c.Redirect(http.StatusFound, "/admin/login")
}

func (h *Handler) Dashboard(c *gin.Context) {
	stats, err := h.storage.GetStatistics(7)
	if err != nil {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"Error": "Failed to load statistics",
		})
		return
	}

	posts, err := h.storage.GetPosts(10, 0)
	if err != nil {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"Error": "Failed to load posts",
		})
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Stats": stats,
		"Posts": posts,
	})
}

func (h *Handler) CreatePost(c *gin.Context) {
	content := c.PostForm("content")
	mediaType := c.PostForm("media_type")
	scheduleTimeStr := c.PostForm("schedule_time")
	sendNow := c.PostForm("send_now") == "true"

	var scheduleTime *time.Time
	if scheduleTimeStr != "" {
		t, err := time.Parse("2006-01-02T15:04", scheduleTimeStr)
		if err == nil {
			scheduleTime = &t
		}
	}

	// Обработка кнопок
	var buttons []models.Button
	buttonTexts := c.PostFormArray("button_text")
	buttonURLs := c.PostFormArray("button_url")

	for i := range buttonTexts {
		if buttonTexts[i] != "" && buttonURLs[i] != "" {
			buttons = append(buttons, models.Button{
				Text: buttonTexts[i],
				URL:  buttonURLs[i],
			})
		}
	}

	buttonsJSON, _ := json.Marshal(buttons)

	post := models.Post{
		Content:      content,
		MediaType:    mediaType,
		Buttons:      buttonsJSON,
		ScheduleTime: scheduleTime,
		Status:       "draft",
		CreatedBy:    c.MustGet("username").(string),
		CreatedAt:    time.Now(),
	}

	// Обработка загрузки медиа
	file, header, err := c.Request.FormFile("media")
	if err == nil {
		defer file.Close()

		uploadDir := "web/assets/uploads"
		os.MkdirAll(uploadDir, 0755)

		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
		filepath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(filepath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		post.MediaPath = filepath
	}

	if sendNow {
		post.Status = "sending"
	} else if scheduleTime != nil {
		post.Status = "scheduled"
	} else {
		post.Status = "draft"
	}

	if err := h.storage.CreatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Немедленная отправка
	if sendNow {
		go h.sendPostToChannels(post)
	}

	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func (h *Handler) sendPostToChannels(post models.Post) {
	channels, err := h.storage.GetActiveChannels()
	if err != nil {
		return
	}

	for _, channel := range channels {
		messageID, err := h.telegram.SendMessage(channel.TelegramID, post)

		status := "sent"
		errorMsg := ""
		if err != nil {
			status = "error"
			errorMsg = err.Error()
		}

		postChannel := models.PostChannel{
			PostID:    post.ID,
			ChannelID: channel.ID,
			MessageID: messageID,
			Status:    status,
			Error:     errorMsg,
			SentAt:    time.Now(),
		}
		h.storage.CreatePostChannel(&postChannel)
	}

	// Обновление статуса поста
	post.Status = "sent"
	now := time.Now()
	post.SentAt = &now
	h.storage.UpdatePost(&post)
}
