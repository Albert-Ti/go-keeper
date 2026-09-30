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

func AuthUnary(secretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "token not found")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "token not found")
		}

		tokenStr := values[0]
		claims := &mytoken.MyCustomClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid || claims.UserID == "" {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return nil, status.Error(codes.Unauthenticated, "access token is expired")
			}
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)
		return handler(newCtx, req)
	}
}

func AuthStream(secretKey string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if publicMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "token not found")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return status.Error(codes.Unauthenticated, "token not found")
		}

		tokenStr := values[0]
		claims := &mytoken.MyCustomClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid || claims.UserID == "" {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return status.Error(codes.Unauthenticated, "access token is expired")
			}
			return status.Error(codes.Unauthenticated, err.Error())
		}

		newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)
		return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: newCtx})
	}
}

// wrappedServerStream переопределяет Context(), чтобы прокинуть новый ctx с userID внутрь handler.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// GetAuthUserID извлекает идентификатор пользователя.
func GetAuthUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user id not found")
	}
	return userID, nil
}
