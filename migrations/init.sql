-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    coins INT DEFAULT 1000
);

-- Создание таблицы транзакций
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    from_user TEXT REFERENCES users(name) ON DELETE CASCADE,
    to_user TEXT REFERENCES users(name) ON DELETE CASCADE,
    amount INT NOT NULL CHECK (amount > 0)
);

-- Создание таблицы инвентаря
CREATE TABLE IF NOT EXISTS inventory (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    item_name TEXT NOT NULL,
    amount INT NOT NULL DEFAULT 1,
    CONSTRAINT unique_user_item UNIQUE (user_id, item_name)
);