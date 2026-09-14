package interceptor

import (
	"context"
	"errors"

	mytoken "github.com/Albert-Ti/go-keeper/internal/token"
	"github.com/golang-jwt/jwt/v5"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type UserIDType string

// UserIDKey — ключ контекста.
const UserIDKey UserIDType = "userID"

var publicMethods = map[string]bool{
	"/gokeeper.GoKeeperService/Register":     true,
	"/gokeeper.GoKeeperService/Login":        true,
	"/gokeeper.GoKeeperService/ConfirmEmail": true,
	"/gokeeper.GoKeeperService/RefreshToken": true,
}

func Auth(secretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		var tokenStr string
		var authorizedUserID string

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get("authorization")
			if len(values) > 0 {
				tokenStr = values[0]
				claims := &mytoken.MyCustomClaims{}

				token, err := jwt.ParseWithClaims(
					tokenStr,
					claims,
					func(t *jwt.Token) (any, error) {
						return []byte(secretKey), nil
					},
				)

				if err != nil || !token.Valid || claims.UserID == "" {
					return nil, status.Error(codes.Unauthenticated, err.Error())
				}
				authorizedUserID = claims.UserID

			} else {
				return nil, status.Error(codes.Unauthenticated, "token not found")
			}
		} else {
			return nil, status.Error(codes.Unauthenticated, "token not found")
		}

		ctx = context.WithValue(ctx, UserIDKey, authorizedUserID)
		return handler(ctx, req)
	}
}

// GetAuthUserID извлекает идентификатор пользователя.
func GetAuthUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user id not found")
	}
	return userID, nil
}
