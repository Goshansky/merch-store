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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	// Проверяем баланс отправителя
	var sender models.User
	err := repositories.DB.Get(&sender, "SELECT coins FROM users WHERE name=$1", username)
	if err != nil || sender.Coins < request.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недостаточно монет"})
		return
	}

	// Проверяем существование получателя
	var receiver models.User
	err = repositories.DB.Get(&receiver, "SELECT id FROM users WHERE name=$1", request.ToUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Получатель не найден"})
		return
	}

	// Обновляем баланс
	tx := repositories.DB.MustBegin()
	tx.MustExec("UPDATE users SET coins = coins - $1 WHERE name = $2", request.Amount, username)
	tx.MustExec("UPDATE users SET coins = coins + $1 WHERE name = $2", request.Amount, request.ToUser)
	tx.MustExec("INSERT INTO transactions (from_user, to_user, amount) VALUES ($1, $2, $3)", username, request.ToUser, request.Amount)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Перевод выполнен"})
}

// BuyItem - покупка товара за монеты
func BuyItem(c *gin.Context) {
	username, _ := c.Get("username")
	itemName := c.Param("item")

	// Проверяем наличие товара
	itemPrices := map[string]int{
		"t-shirt": 80, "cup": 20, "book": 50, "pen": 10,
		"powerbank": 200, "hoody": 300, "umbrella": 200,
		"socks": 10, "wallet": 50, "pink-hoody": 500,
	}
	price, exists := itemPrices[itemName]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Товар не найден"})
		return
	}

	// Проверяем баланс
	var user models.User
	err := repositories.DB.Get(&user, "SELECT id, coins FROM users WHERE name=$1", username)
	if err != nil || user.Coins < price {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недостаточно монет"})
		return
	}

	// Обновляем баланс и добавляем товар в инвентарь
	tx := repositories.DB.MustBegin()
	tx.MustExec("UPDATE users SET coins = coins - $1 WHERE name = $2", price, username)
	tx.MustExec(`
		INSERT INTO inventory (user_id, item_name, amount)
		VALUES ($1, $2, 1) ON CONFLICT (user_id, item_name)
		DO UPDATE SET amount = inventory.amount + 1`,
		user.ID, itemName)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Товар приобретен"})
}
