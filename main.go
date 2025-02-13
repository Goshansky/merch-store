package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"merch-store/handlers"
	"merch-store/middlewares"
	"merch-store/repositories"
)

func main() {

	// Инициализация базы данных
	repositories.InitDB()

	// Проверка подключения
	if err := repositories.DB.Ping(); err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}
	log.Println("Подключение к базе данных успешно!")

	r := gin.Default()

	// Роуты для регистрации и авторизации
	r.POST("/api/register", handlers.Register)
	r.POST("/api/auth", handlers.Auth)

	// Роуты для работы с монетами и товарами
	auth := r.Group("/api")
	auth.Use(middlewares.AuthMiddleware())
	{
		auth.GET("/info", handlers.GetUserInfo)
		auth.POST("/sendCoin", handlers.SendCoin)
		auth.POST("/buy/:item", handlers.BuyItem)
	}

	r.Run(":8080")
}
