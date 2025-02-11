package handlers

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"merch-store/models"
	"merch-store/repositories"
	"merch-store/utils"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// Хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при хешировании пароля"})
		return
	}

	// Создаем пользователя в базе данных
	_, err = repositories.DB.Exec("INSERT INTO users (name, password, coins) VALUES ($1, $2, $3)", req.Username, hash, 1000)
	if err != nil {
		log.Println("Ошибка создания пользователя:", err)
		c.JSON(http.StatusConflict, gin.H{"error": "Пользователь уже существует"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Регистрация успешна"})
}

// Auth - аутентификация пользователя и выдача JWT-токена
func Auth(c *gin.Context) {
	var creds models.User
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	var user models.User
	err := repositories.DB.Get(&user, "SELECT id, name, password, coins FROM users WHERE name=$1", creds.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	// Проверяем пароль
	if !utils.CheckPasswordHash(creds.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный пароль"})
		return
	}

	// Генерируем JWT-токен
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetUserInfo - получение информации о пользователе
func GetUserInfo(c *gin.Context) {
	username, _ := c.Get("username")

	var user models.User
	err := repositories.DB.Get(&user, "SELECT id, name, coins FROM users WHERE name=$1", username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных"})
		return
	}

	var inventory []models.Inventory
	err = repositories.DB.Select(&inventory, "SELECT item_name, amount FROM inventory WHERE user_id=$1", user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения инвентаря"})
		return
	}

	var transactions []models.Transaction
	err = repositories.DB.Select(&transactions, "SELECT from_user, to_user, amount FROM transactions WHERE from_user=$1 OR to_user=$1", username)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения транзакций"})
		return
	}

	// Формируем историю монет
	coinHistory := gin.H{
		"received": []gin.H{},
		"sent":     []gin.H{},
	}
	for _, t := range transactions {
		if t.ToUser == username {
			coinHistory["received"] = append(coinHistory["received"].([]gin.H), gin.H{"fromUser": t.FromUser, "amount": t.Amount})
		} else {
			coinHistory["sent"] = append(coinHistory["sent"].([]gin.H), gin.H{"toUser": t.ToUser, "amount": t.Amount})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"coins":       user.Coins,
		"inventory":   inventory,
		"coinHistory": coinHistory,
	})
}
