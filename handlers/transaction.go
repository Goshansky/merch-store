package handlers

import (
	"merch-store/services"
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

	err := services.SendCoin(username.(string), request.ToUser, request.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"description": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"description": "Успешная передача монет."})
}

// BuyItem - покупка товара за монеты
func BuyItem(c *gin.Context) {
	username, _ := c.Get("username")
	itemName := c.Param("item")

	var request struct {
		Amount int `json:"amount"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"description": "Некорректное количество товара"})
		return
	}

	err := services.BuyItem(username.(string), itemName, request.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"description": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"description": "Товар приобретен", "amount": request.Amount})
}
