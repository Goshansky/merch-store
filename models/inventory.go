package models

type Inventory struct {
	ID       uint   `db:"id"`
	UserID   uint   `db:"user_id"`
	ItemName string `db:"item_name"`
	Amount   int    `db:"amount"`
}
