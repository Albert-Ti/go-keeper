package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"time"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicMethods = map[string]bool{
	"/gokeeper.GoKeeperService/Register":     true,
	"/gokeeper.GoKeeperService/Login":        true,
	"/gokeeper.GoKeeperService/ConfirmEmail": true,
	"/gokeeper.GoKeeperService/RefreshToken": true,
}

const (
	expiryLeeway   = 30 * time.Second // обновляем чуть раньше реального истечения
	refreshTimeout = 5 * time.Second
)

type AuthInterceptor struct {
	mu           sync.Mutex // сериализует refresh между горутинами
	localStorage *FileStorage
}

func NewAuthInterceptor(s *FileStorage) *AuthInterceptor {
	return &AuthInterceptor{localStorage: s}
}

func (a *AuthInterceptor) UnaryInterceptor(
	ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	if publicMethods[method] {
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	// 1. Берём валидный токен (при необходимости обновляем заранее).
	token, err := a.ensureToken(ctx, cc, "")
	if err != nil {
		return err
	}

	err = invoker(withToken(ctx, token), method, req, reply, cc, opts...)
	if status.Code(err) != codes.Unauthenticated {
		return err
	}

	// 2. Сервер всё равно отверг токен (например, отозван или часы разошлись):
	//    обновляем и повторяем запрос один раз.
	token, rerr := a.ensureToken(ctx, cc, token)
	if rerr != nil {
		return err // исходная Unauthenticated: UI отправит на логин
	}

	return invoker(withToken(ctx, token), method, req, reply, cc, opts...)
}

func (a *AuthInterceptor) StreamInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	if publicMethods[method] {
		return streamer(ctx, desc, cc, method, opts...)
	}

	// Для стримов обновляем только заранее: отказ приходит уже при Recv,
	// и прозрачно повторить стрим после отправленных сообщений нельзя.
	token, err := a.ensureToken(ctx, cc, "")
	if err != nil {
		return nil, err
	}

	return streamer(withToken(ctx, token), desc, cc, method, opts...)
}

// ensureToken возвращает действующий access-токен.
// rejected — токен, который сервер только что отверг ("" если не было такого).
func (a *AuthInterceptor) ensureToken(
	ctx context.Context,
	cc grpc.ClientConnInterface,
	rejected string,
) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Под локом перечитываем: возможно, соседняя горутина уже обновила токен.
	cur := a.localStorage.Get(accessTokenKey)
	if cur != "" && cur != rejected && !tokenExpired(cur, expiryLeeway) {
		return cur, nil
	}

	refresh := a.localStorage.Get(refreshTokenKey)
	if refresh == "" {
		return "", status.Error(codes.Unauthenticated, "not logged in")
	}

	rctx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()

	resp, err := pb.NewGoKeeperServiceClient(cc).RefreshToken(rctx, pb.TokenRequest_builder{
		RefreshToken: refresh,
	}.Build())
	if err != nil {
		return "", err
	}

	if err := a.localStorage.Set(accessTokenKey, resp.GetAccessToken()); err != nil {
		return "", err
	}
	if err := a.localStorage.Set(refreshTokenKey, resp.GetRefreshToken()); err != nil {
		return "", err
	}

	return resp.GetAccessToken(), nil
}

func withToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", token)
}

// tokenExpired читает exp из JWT без проверки подписи (проверяет сервер).
// Если токен не JWT или exp нет, считаем его валидным: тогда сработает
// реактивный путь по Unauthenticated.
func tokenExpired(token string, leeway time.Duration) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return false
	}
	return time.Now().Add(leeway).After(time.Unix(claims.Exp, 0))
}
