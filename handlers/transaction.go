package handlers

import (
	"merch-store/models"
	"merch-store/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SendCoin - передача монет другому пользователю
func SendCoin(c *gin.Context) {
	username, _ := c.Get("username")

	var request struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"description": "Неверный запрос."})
		return
	}

	// Проверяем баланс отправителя
	var sender models.User
	err := repositories.DB.Get(&sender, "SELECT coins FROM users WHERE name=$1", username)
	if err != nil || sender.Coins < request.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"description": "Недостаточно монет."})
		return
	}

	// Проверяем существование получателя
	var receiver models.User
	err = repositories.DB.Get(&receiver, "SELECT id FROM users WHERE name=$1", request.ToUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"description": "Получатель не найден"})
		return
	}

	// Обновляем баланс
	tx := repositories.DB.MustBegin()
	tx.MustExec("UPDATE users SET coins = coins - $1 WHERE name = $2", request.Amount, username)
	tx.MustExec("UPDATE users SET coins = coins + $1 WHERE name = $2", request.Amount, request.ToUser)
	tx.MustExec("INSERT INTO transactions (from_user, to_user, amount) VALUES ($1, $2, $3)", username, request.ToUser, request.Amount)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"description": "Успешный ответ."})
}

// BuyItem - покупка товара за монеты
func BuyItem(c *gin.Context) {
	username, _ := c.Get("username")
	itemName := c.Param("item")

	// Структура для парсинга тела запроса
	var request struct {
		Amount int `json:"amount"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"description": "Некорректное количество товара"})
		return
	}

	// Проверяем наличие товара
	itemPrices := map[string]int{
		"t-shirt": 80, "cup": 20, "book": 50, "pen": 10,
		"powerbank": 200, "hoody": 300, "umbrella": 200,
		"socks": 10, "wallet": 50, "pink-hoody": 500,
	}
	price, exists := itemPrices[itemName]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"description": "Товар не найден"})
		return
	}

	totalCost := price * request.Amount

	// Проверяем баланс пользователя
	var user models.User
	err := repositories.DB.Get(&user, "SELECT id, coins FROM users WHERE name=$1", username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"description": "Ошибка получения данных пользователя"})
		return
	}
	if user.Coins < totalCost {
		c.JSON(http.StatusBadRequest, gin.H{"description": "Недостаточно монет"})
		return
	}

	// Обновляем баланс и добавляем товар в инвентарь
	tx := repositories.DB.MustBegin()

	_, err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE name = $2", totalCost, username)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"description": "Ошибка обновления баланса"})
		return
	}

	_, err = tx.Exec(`
		INSERT INTO inventory (user_id, item_name, amount)
		VALUES ($1, $2, $3) 
		ON CONFLICT (user_id, item_name)
		DO UPDATE SET amount = inventory.amount + EXCLUDED.amount`,
		user.ID, itemName, request.Amount)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"description": "Ошибка обновления инвентаря"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"description": "Ошибка сохранения транзакции"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"description": "Товар приобретен", "amount": request.Amount})
}
