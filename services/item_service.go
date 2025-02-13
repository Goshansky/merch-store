package services

import (
	"errors"
	"merch-store/models"
	"merch-store/repositories"
)

// BuyItem - бизнес-логика для покупки товара
func BuyItem(username, itemName string, amount int) error {
	// Проверяем наличие товара
	itemPrices := map[string]int{
		"t-shirt": 80, "cup": 20, "book": 50, "pen": 10,
		"powerbank": 200, "hoody": 300, "umbrella": 200,
		"socks": 10, "wallet": 50, "pink-hoody": 500,
	}
	price, exists := itemPrices[itemName]
	if !exists {
		return errors.New("товар не найден")
	}

	totalCost := price * amount

	// Проверяем баланс пользователя
	var user models.User
	err := repositories.DB.Get(&user, "SELECT id, coins FROM users WHERE name=$1", username)
	if err != nil {
		return errors.New("ошибка получения данных пользователя")
	}
	if user.Coins < totalCost {
		return errors.New("недостаточно монет")
	}

	// Обновляем баланс и добавляем товар в инвентарь
	tx := repositories.DB.MustBegin()

	_, err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE name = $2", totalCost, username)
	if err != nil {
		tx.Rollback()
		return errors.New("ошибка обновления баланса")
	}

	_, err = tx.Exec(
		"INSERT INTO inventory (user_id, item_name, amount) VALUES ($1, $2, $3) ON CONFLICT (user_id, item_name) DO UPDATE SET amount = inventory.amount + EXCLUDED.amount",
		user.ID, itemName, amount)
	if err != nil {
		tx.Rollback()
		return errors.New("ошибка обновления инвентаря")
	}

	err = tx.Commit()
	if err != nil {
		return errors.New("ошибка сохранения транзакции")
	}

	return nil
}
