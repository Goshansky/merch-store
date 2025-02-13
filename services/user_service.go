package services

import (
	"errors"
	"fmt"
	"merch-store/models"
	"merch-store/repositories"
	"merch-store/utils"
)

// RegisterUser - регистрация нового пользователя
func RegisterUser(username, password string) error {
	// Хешируем пароль
	hash, err := utils.HashPassword(password)
	if err != nil {
		return errors.New("внутренняя ошибка сервера")
	}

	// Создаем пользователя в базе данных
	_, err = repositories.DB.Exec("INSERT INTO users (name, password, coins) VALUES ($1, $2, $3)", username, hash, 1000)
	if err != nil {
		return errors.New("пользователь уже существует")
	}

	return nil
}

// AuthenticateUser - аутентификация пользователя
func AuthenticateUser(username, password string) (string, error) {
	var user models.User
	err := repositories.DB.Get(&user, "SELECT id, name, password, coins FROM users WHERE name=$1", username)
	if err != nil {
		return "", errors.New("неавторизован")
	}

	// Проверяем пароль
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("неавторизован")
	}

	// Генерируем JWT-токен
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		return "", errors.New("внутренняя ошибка сервера")
	}

	return token, nil
}

type UserInfo struct {
	Coins       int                                 `json:"coins"`
	Inventory   []UserItem                          `json:"inventory"`
	CoinHistory map[string][]map[string]interface{} `json:"coinHistory"`
}

type UserItem struct {
	ItemName string `db:"item_name"`
	Amount   int    `db:"amount"`
}

func GetUserInfo(username string) (UserInfo, error) {
	var user models.User
	err := repositories.DB.Get(&user, "SELECT id, name, coins FROM users WHERE name=$1", username)
	if err != nil {
		return UserInfo{}, fmt.Errorf("error fetching user: %w", err)
	}

	var inventory []UserItem
	err = repositories.DB.Select(&inventory, "SELECT item_name, amount FROM inventory WHERE user_id=$1", user.ID)
	if err != nil {
		return UserInfo{}, fmt.Errorf("error fetching inventory: %w", err)
	}

	var transactions []models.Transaction
	err = repositories.DB.Select(&transactions, "SELECT from_user, to_user, amount FROM transactions WHERE from_user=$1 OR to_user=$1", username)
	if err != nil {
		return UserInfo{}, fmt.Errorf("error fetching transactions: %w", err)
	}

	coinHistory := map[string][]map[string]interface{}{
		"received": {},
		"sent":     {},
	}

	for _, t := range transactions {
		if t.ToUser == username {
			coinHistory["received"] = append(coinHistory["received"], map[string]interface{}{"fromUser": t.FromUser, "amount": t.Amount})
		} else {
			coinHistory["sent"] = append(coinHistory["sent"], map[string]interface{}{"toUser": t.ToUser, "amount": t.Amount})
		}
	}

	return UserInfo{
		Coins:       user.Coins,
		Inventory:   inventory,
		CoinHistory: coinHistory,
	}, nil
}
