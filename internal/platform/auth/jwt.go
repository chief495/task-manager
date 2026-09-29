package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — структура данных в JWT‑токене
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"` // "guest", "user", "admin"
	jwt.RegisteredClaims
}

// Секретный ключ для подписи (в продакшене храните в env)
var jwtKey = []byte("your-secret-key-change-in-production")

// GenerateJWT — создаёт JWT‑токен для пользователя
func GenerateJWT(userID uint, email, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // токен действует 24 ч

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "task-manager",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken — вспомогательная функция для парсинга токена
func ParseToken(tokenString string, claims *Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	return token, err
}
