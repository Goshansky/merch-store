package models

type Transaction struct {
	ID       uint   `db:"id"`
	FromUser string `db:"from_user"`
	ToUser   string `db:"to_user"`
	Amount   int    `db:"amount"`
}
