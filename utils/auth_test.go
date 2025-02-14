package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Проверяем, что хеш не пустой
	if hashedPassword == "" {
		t.Fatal("Expected hashed password to be non-empty")
	}

	// Проверяем, что хеш соответствует исходному паролю
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		t.Fatalf("Expected password to match hashed password, got %v", err)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Проверяем, что хеш соответствует исходному паролю
	if !CheckPasswordHash(password, string(hashedPassword)) {
		t.Fatal("Expected password to match hashed password")
	}

	// Проверяем, что неверный пароль не соответствует хешу
	if CheckPasswordHash("wrongpassword", string(hashedPassword)) {
		t.Fatal("Expected password to not match hashed password")
	}
}

func TestGenerateJWT(t *testing.T) {
	username := "testuser"
	tokenString, err := GenerateJWT(username)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Проверяем, что токен не пустой
	if tokenString == "" {
		t.Fatal("Expected token to be non-empty")
	}

	// Проверяем, что токен можно распарсить и что он содержит правильные claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("Expected token to be valid, got %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Expected claims to be of type jwt.MapClaims")
	}

	// Проверяем, что username в claims совпадает с ожидаемым
	if claims["username"] != username {
		t.Fatalf("Expected username to be %s, got %v", username, claims["username"])
	}

	// Проверяем, что токен имеет правильное время истечения
	exp := int64(claims["exp"].(float64))
	if time.Unix(exp, 0).Before(time.Now().Add(23*time.Hour)) || time.Unix(exp, 0).After(time.Now().Add(25*time.Hour)) {
		t.Fatalf("Expected expiration time to be within 24 hours, got %v", time.Unix(exp, 0))
	}
}
