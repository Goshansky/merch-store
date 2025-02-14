package services

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"merch-store/repositories"
	"regexp"
	"testing"
)

func setupMockDB() (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestSendCoinService(t *testing.T) {
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

	err := SendCoin("user1", "user2", 100)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuyItemService(t *testing.T) {
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

	err := BuyItem("user1", "t-shirt", 2)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserInfoService(t *testing.T) {
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

	userInfo, err := GetUserInfo("user1")
	assert.NoError(t, err)
	assert.Equal(t, 1000, userInfo.Coins)
	assert.Equal(t, 1, len(userInfo.Inventory))
	assert.Equal(t, 1, len(userInfo.CoinHistory["received"]))
	assert.NoError(t, mock.ExpectationsWereMet())
}
