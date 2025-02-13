package tests

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"merch-store/handlers"
	"merch-store/middlewares"
	"merch-store/models"
	"merch-store/repositories"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тестирую регистрацию
func TestRegister(t *testing.T) {
	// Инициализируем базу данных
	repositories.InitDB()
	// Очищаем тестовую таблицу пользователей
	repositories.DB.Exec("DELETE FROM users")

	// Регистрируем пользователя
	registerUser(t, "testuser", "password123")

	// Проверяем, что пользователь был добавлен в базу
	var user models.User
	err := repositories.DB.Get(&user, "SELECT name FROM users WHERE name=$1", "testuser")
	assert.Nil(t, err)
	assert.Equal(t, "testuser", user.Username)

	repositories.DB.Exec("DELETE FROM users")
}

// Тестирую авторизацию
func TestAuth(t *testing.T) {
	// Инициализируем базу данных
	repositories.InitDB()
	// Очищаем тестовую таблицу пользователей
	repositories.DB.Exec("DELETE FROM users")

	// Регистрируем пользователя
	registerUser(t, "testuser", "password123")

	// Авторизуем пользователя
	token := authUser(t, "testuser", "password123")

	// Проверяем, что токен был возвращен
	assert.NotEmpty(t, token)

	repositories.DB.Exec("DELETE FROM users")
}

// Тестирую покупку мерча
func TestBuyItem(t *testing.T) {
	// Инициализируем базу данных
	repositories.InitDB()
	// Очищаем тестовую таблицу пользователей и инвентаря
	repositories.DB.Exec("DELETE FROM users")
	repositories.DB.Exec("DELETE FROM inventory")

	// Регистрируем пользователя
	registerUser(t, "testuser", "password123")

	// Авторизуем пользователя
	token := authUser(t, "testuser", "password123")

	// Подготовка запроса на покупку товара
	buyRequest := map[string]int{
		"amount": 1,
	}
	buyRequestBody, _ := json.Marshal(buyRequest)
	buyReq, _ := http.NewRequest("POST", "/api/buy/t-shirt", bytes.NewBuffer(buyRequestBody))
	buyReq.Header.Set("Authorization", "Bearer "+token)
	buyW := httptest.NewRecorder()

	// Инициализация маршрутов с middleware
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	api := router.Group("/api")
	api.Use(middlewares.AuthMiddleware())
	{
		api.POST("/buy/:item", handlers.BuyItem)
	}

	// Выполняем запрос на покупку товара
	router.ServeHTTP(buyW, buyReq)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, buyW.Code)

	// Проверяем, что товар был добавлен в инвентарь
	var inventory []models.Inventory
	err := repositories.DB.Select(&inventory, "SELECT item_name, amount FROM inventory WHERE user_id=(SELECT id FROM users WHERE name=$1)", "testuser")
	assert.Nil(t, err)
	assert.Equal(t, "t-shirt", inventory[0].ItemName)
	assert.Equal(t, 1, inventory[0].Amount)

	repositories.DB.Exec("DELETE FROM users")
	repositories.DB.Exec("DELETE FROM inventory")
}

// Тестирую перевод монет
func TestSendCoin(t *testing.T) {
	// Инициализируем базу данных
	repositories.InitDB()
	// Очищаем тестовую таблицу пользователей и транзакций
	repositories.DB.Exec("DELETE FROM users")
	repositories.DB.Exec("DELETE FROM transactions")

	// Регистрируем двух пользователей
	registerUser(t, "sender", "password123")
	registerUser(t, "receiver", "password123")

	// Авторизуем отправителя
	token := authUser(t, "sender", "password123")

	// Подготовка запроса на перевод монет
	sendCoinRequest := map[string]interface{}{
		"toUser": "receiver",
		"amount": 50,
	}
	sendCoinRequestBody, _ := json.Marshal(sendCoinRequest)
	sendCoinReq, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(sendCoinRequestBody))
	sendCoinReq.Header.Set("Authorization", "Bearer "+token)
	sendCoinW := httptest.NewRecorder()

	// Инициализация маршрутов с middleware
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	api := router.Group("/api")
	api.Use(middlewares.AuthMiddleware())
	{
		api.POST("/sendCoin", handlers.SendCoin)
	}

	// Выполняем запрос на перевод монет
	router.ServeHTTP(sendCoinW, sendCoinReq)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, sendCoinW.Code)

	// Проверяем, что монеты были переведены
	var sender, receiver models.User
	err := repositories.DB.Get(&sender, "SELECT coins FROM users WHERE name=$1", "sender")
	assert.Nil(t, err)
	assert.Equal(t, 950, sender.Coins)

	err = repositories.DB.Get(&receiver, "SELECT coins FROM users WHERE name=$1", "receiver")
	assert.Nil(t, err)
	assert.Equal(t, 1050, receiver.Coins)

	// Проверяем, что транзакция была записана
	var transaction models.Transaction
	err = repositories.DB.Get(&transaction, "SELECT from_user, to_user, amount FROM transactions WHERE from_user=$1 AND to_user=$2", "sender", "receiver")
	assert.Nil(t, err)
	assert.Equal(t, "sender", transaction.FromUser)
	assert.Equal(t, "receiver", transaction.ToUser)
	assert.Equal(t, 50, transaction.Amount)

	repositories.DB.Exec("DELETE FROM users")
	repositories.DB.Exec("DELETE FROM transactions")
}

// Тестирую информацию о пользователе
func TestGetUserInfo(t *testing.T) {
	// Инициализируем базу данных
	repositories.InitDB()
	// Очищаем тестовую таблицу пользователей и инвентаря
	repositories.DB.Exec("DELETE FROM users")
	repositories.DB.Exec("DELETE FROM inventory")
	repositories.DB.Exec("DELETE FROM transactions")

	// Регистрируем пользователя
	registerUser(t, "testuser", "password123")

	// Авторизуем пользователя
	token := authUser(t, "testuser", "password123")

	// Подготовка запроса на получение информации о пользователе
	getUserInfoReq, _ := http.NewRequest("GET", "/api/info", nil)
	getUserInfoReq.Header.Set("Authorization", "Bearer "+token)
	getUserInfoW := httptest.NewRecorder()

	// Инициализация маршрутов с middleware
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	api := router.Group("/api")
	api.Use(middlewares.AuthMiddleware())
	{
		api.GET("/info", handlers.GetUserInfo)
	}

	// Выполняем запрос на получение информации о пользователе
	router.ServeHTTP(getUserInfoW, getUserInfoReq)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, getUserInfoW.Code)

	// Проверяем, что информация о пользователе корректна
	var response map[string]interface{}
	json.Unmarshal(getUserInfoW.Body.Bytes(), &response)

	assert.Equal(t, "Успешный ответ.", response["description"])
	assert.Equal(t, 1000, int(response["schema"].(map[string]interface{})["coins"].(float64)))
	assert.Empty(t, response["schema"].(map[string]interface{})["inventory"])
	assert.Empty(t, response["schema"].(map[string]interface{})["coinHistory"].(map[string]interface{})["received"])
	assert.Empty(t, response["schema"].(map[string]interface{})["coinHistory"].(map[string]interface{})["sent"])
}

// Вспомогательная функция для регистрации
func registerUser(t *testing.T, username, password string) {
	registerRequest := map[string]string{
		"username": username,
		"password": password,
	}
	registerRequestBody, _ := json.Marshal(registerRequest)
	registerReq, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(registerRequestBody))
	registerW := httptest.NewRecorder()

	router := gin.Default()
	router.POST("/api/register", handlers.Register)

	// Выполняем запрос на регистрацию
	router.ServeHTTP(registerW, registerReq)
	assert.Equal(t, http.StatusOK, registerW.Code)
}

// Вспомогательная функция для авторизации
func authUser(t *testing.T, username, password string) string {
	authRequest := map[string]string{
		"username": username,
		"password": password,
	}
	authRequestBody, _ := json.Marshal(authRequest)
	authReq, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authRequestBody))
	authW := httptest.NewRecorder()

	router := gin.Default()
	router.POST("/api/auth", handlers.Auth)

	// Выполняем запрос на авторизацию
	router.ServeHTTP(authW, authReq)
	assert.Equal(t, http.StatusOK, authW.Code)

	// Получаем токен из ответа
	var authResponse map[string]string
	json.Unmarshal(authW.Body.Bytes(), &authResponse)
	return authResponse["token"]
}
