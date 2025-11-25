package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/your-username/channelBot/config"
	"github.com/your-username/mgok-alert-bot/internal/admin"
	"github.com/your-username/mgok-alert-bot/internal/auth"
	"github.com/your-username/mgok-alert-bot/internal/scheduler"
	"github.com/your-username/mgok-alert-bot/internal/storage"
	"github.com/your-username/mgok-alert-bot/internal/telegram"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация хранилища
	db, err := storage.NewPostgresStorage(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Инициализация Telegram клиента
	tgClient := telegram.NewClient(cfg.TelegramBotToken)

	// Инициализация аутентификации
	auth := auth.NewAuth(cfg.JWTSecret)

	// Инициализация планировщика
	sched := scheduler.NewScheduler(db, tgClient)
	sched.Start()

	// Инициализация обработчиков админ-панели
	adminHandler := admin.NewHandler(db, tgClient, auth, cfg)

	// Настройка Gin
	if cfg.LogLevel == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Загрузка шаблонов
	router.LoadHTMLGlob("web/templates/*")

	// Статические файлы
	router.Static("/static", "./web/static")
	router.Static("/assets", "./web/assets")

	// Группа админ-панели с аутентификацией
	adminGroup := router.Group("/admin")
	adminGroup.Use(admin.AuthMiddleware(auth))
	{
		adminGroup.GET("/dashboard", adminHandler.Dashboard)
		adminGroup.GET("/channels", adminHandler.Channels)
		adminGroup.POST("/channels", adminHandler.CreateChannel)
		adminGroup.GET("/posts/create", adminHandler.CreatePostPage)
		adminGroup.POST("/posts/create", adminHandler.CreatePost)
		adminGroup.GET("/statistics", adminHandler.Statistics)
		adminGroup.POST("/logout", adminHandler.Logout)
	}

	// Аутентификация
	router.GET("/admin/login", adminHandler.LoginPage)
	router.POST("/admin/login", adminHandler.Login)

	// Корневой маршрут
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/admin/dashboard")
	})

	log.Printf("Starting server on :%s", cfg.WebPort)
	if err := router.Run(":" + cfg.WebPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
