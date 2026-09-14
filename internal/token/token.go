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

func CreateAccessToken(userID string, secretKey string) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &MyCustomClaims{
		jwt.RegisteredClaims{
			// TODO 2 минуты для тестирования клиента для вызова RefreshToken
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
		},
		userID,
	}

	return t.SignedString([]byte(secretKey))
}

func CreateRefreshToken(userID string, secretKey string) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &MyCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
		userID,
	}

	return t.SignedString([]byte(secretKey))
}

func CreateConfirmEmailToken(userID string, secretKey string) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &MyCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
		userID,
	}

	return t.SignedString([]byte(secretKey))
}
