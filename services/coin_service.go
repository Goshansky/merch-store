package services

import (
	"errors"
	"merch-store/models"
	"merch-store/repositories"
)

// SendCoin - бизнес-логика для передачи монет
func SendCoin(fromUser, toUser string, amount int) error {
	// Проверяем баланс отправителя
	var sender models.User
	err := repositories.DB.Get(&sender, "SELECT coins FROM users WHERE name=$1", fromUser)
	if err != nil || sender.Coins < amount {
		return errors.New("недостаточно монет")
	}

	// Проверяем существование получателя
	var receiver models.User
	err = repositories.DB.Get(&receiver, "SELECT id FROM users WHERE name=$1", toUser)
	if err != nil {
		return errors.New("получатель не найден")
	}

	// Обновляем баланс
	tx := repositories.DB.MustBegin()
	tx.MustExec("UPDATE users SET coins = coins - $1 WHERE name = $2", amount, fromUser)
	tx.MustExec("UPDATE users SET coins = coins + $1 WHERE name = $2", amount, toUser)
	tx.MustExec("INSERT INTO transactions (from_user, to_user, amount) VALUES ($1, $2, $3)", fromUser, toUser, amount)
	tx.Commit()

	return nil
}
