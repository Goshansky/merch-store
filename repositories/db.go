package repositories

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func InitDB() {
	var err error
	DB, err = sqlx.Connect("postgres", "host=localhost port=5431 user=postgres password=password dbname=shop sslmode=disable")
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	log.Println("Успешное подключение к базе данных!")

	DB.MustExec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			coins INT DEFAULT 1000
		);
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			from_user TEXT,
			to_user TEXT,
			amount INT
		);
		CREATE TABLE IF NOT EXISTS inventory (
			id SERIAL PRIMARY KEY,
    		user_id INT REFERENCES users(id) ON DELETE CASCADE,
    		item_name TEXT NOT NULL,
    		amount INT NOT NULL DEFAULT 1,
    		CONSTRAINT unique_user_item UNIQUE (user_id, item_name)
		);
	`)
}
