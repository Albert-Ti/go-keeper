package interceptor

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Logging - интерцептор для логирования unary gRPC запросов.
func Logging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		logCall(ctx, info.FullMethod, start, err)

		return resp, err
	}
}

// LoggingStream - интерцептор для логирования streaming gRPC запросов.
func LoggingStream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		wrapped := &loggingServerStream{ServerStream: ss}

		err := handler(srv, wrapped)

		logCall(ss.Context(), info.FullMethod, start, err,
			slog.Bool("client_stream", info.IsClientStream),
			slog.Bool("server_stream", info.IsServerStream),
			slog.Int64("msgs_received", wrapped.received.Load()),
			slog.Int64("msgs_sent", wrapped.sent.Load()),
		)

		return err
	}
}

// loggingServerStream - обёртка над grpc.ServerStream для подсчёта сообщений.
type loggingServerStream struct {
	grpc.ServerStream
	received atomic.Int64
	sent     atomic.Int64
}

func (s *loggingServerStream) RecvMsg(m any) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.received.Add(1)
	}
	return err // io.EOF не считаем и не подменяем
}

func (s *loggingServerStream) SendMsg(m any) error {
	err := s.ServerStream.SendMsg(m)
	if err == nil {
		s.sent.Add(1)
	}
	return err
}

// logCall - общая логика логирования для unary и stream.
func logCall(ctx context.Context, method string, start time.Time, err error, extra ...slog.Attr) {
	code := status.Code(err) // для nil вернёт OK, для не-status ошибок - Unknown

	attrs := []slog.Attr{
		slog.String("method", method),
		slog.Duration("duration", time.Since(start)),
		slog.String("code", code.String()),
	}
	attrs = append(attrs, extra...)

	level := levelByCode(code)
	msg := "gRPC request success"
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		msg = "gRPC request failed"
	}

	slog.LogAttrs(ctx, level, msg, attrs...)
}

// levelByCode - серверные ошибки в Error, клиентские в Warn, остальное в Info.
func levelByCode(code codes.Code) slog.Level {
	switch code {
	case codes.OK:
		return slog.LevelInfo
	case codes.Unknown, codes.Internal, codes.DataLoss, codes.Unavailable, codes.DeadlineExceeded:
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}
