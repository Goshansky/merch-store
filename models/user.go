package models

type User struct {
	ID       uint   `db:"id"`
	Username string `db:"name"`
	Password string `db:"password"`
	Coins    int    `db:"coins"`
}
