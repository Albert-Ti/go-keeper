package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MyCustomClaims расширяет стандартные claims jwt.RegisteredClaims полем UserID,
// чтобы связать выданный токен с конкретным пользователем сервиса.
// generate:reset
type MyCustomClaims struct {
	jwt.RegisteredClaims
	UserID string
}

func CreateAccessToken(userID string, secretKey string, expiresAt time.Duration) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &MyCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresAt)),
		},
		userID,
	}

	return t.SignedString([]byte(secretKey))
}

func CreateRefreshToken(userID string, secretKey string, expiresAt time.Duration) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &MyCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresAt)),
		},
		userID,
	}

	return t.SignedString([]byte(secretKey))
}
