package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"merch-store/repositories"
)

func setupMockDB() (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestSendCoinHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sqlxDB, mock := setupMockDB()
	repositories.DB = sqlxDB

	mock.ExpectQuery(regexp.QuoteMeta("SELECT coins FROM users WHERE name=$1")).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(1000))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE name=$1")).
		WithArgs("user2").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET coins = coins - $1 WHERE name = $2")).
		WithArgs(100, "user1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET coins = coins + $1 WHERE name = $2")).
		WithArgs(100, "user2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO transactions")).
		WithArgs("user1", "user2", 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "user1")

	requestBody := bytes.NewBufferString(`{"toUser": "user2", "amount": 100}`)
	c.Request, _ = http.NewRequest("POST", "/sendCoin", requestBody)
	c.Request.Header.Set("Content-Type", "application/json")

	SendCoin(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuyItemHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sqlxDB, mock := setupMockDB()
	repositories.DB = sqlxDB

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, coins FROM users WHERE name=$1")).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).AddRow(1, 1000))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET coins = coins - $1 WHERE name = $2")).
		WithArgs(160, "user1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO inventory")).
		WithArgs(1, "t-shirt", 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "user1")
	c.Params = append(c.Params, gin.Param{Key: "item", Value: "t-shirt"})

	requestBody := bytes.NewBufferString(`{"amount": 2}`)
	c.Request, _ = http.NewRequest("POST", "/buy/t-shirt", requestBody)
	c.Request.Header.Set("Content-Type", "application/json")

	BuyItem(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserInfoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sqlxDB, mock := setupMockDB()
	repositories.DB = sqlxDB

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, coins FROM users WHERE name=$1")).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "coins"}).AddRow(1, "user1", 1000))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT item_name, amount FROM inventory WHERE user_id=$1")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"item_name", "amount"}).AddRow("t-shirt", 2))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT from_user, to_user, amount FROM transactions WHERE from_user=$1 OR to_user=$1")).
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"from_user", "to_user", "amount"}).AddRow("user2", "user1", 100))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "user1")

	GetUserInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sqlxDB, mock := setupMockDB()
	repositories.DB = sqlxDB

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs("user1", sqlmock.AnyArg(), 1000).
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := bytes.NewBufferString(`{"username": "user1", "password": "password123"}`)
	c.Request, _ = http.NewRequest("POST", "/register", requestBody)
	c.Request.Header.Set("Content-Type", "application/json")

	Register(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
