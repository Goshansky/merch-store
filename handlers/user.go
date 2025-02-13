package handlers

import (
	"merch-store/models"
	"merch-store/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register - регистрация пользователя
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"description": "Неверный запрос."})
		return
	}

	err := services.RegisterUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"description": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"description": "Пользователь зарегистрирован."})
}

// Auth - аутентификация пользователя и выдача JWT-токена
func Auth(c *gin.Context) {
	var creds models.User
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"description": "Неверный запрос."})
		return
	}

	token, err := services.AuthenticateUser(creds.Username, creds.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"description": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func GetUserInfo(c *gin.Context) {
	username, _ := c.Get("username")

	userInfo, err := services.GetUserInfo(username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"description": "Внутренняя ошибка сервера."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"description": "Успешный ответ.",
		"schema": gin.H{
			"coins":       userInfo.Coins,
			"inventory":   userInfo.Inventory,
			"coinHistory": userInfo.CoinHistory,
		},
	})
}
