package main

import (
	"log/slog"
	"os"
)

// TODO Использовать для разработки и тестирования
// будет печатать логи в файл, чтобы не внутри процесса с TUI
func setupLogger(path string) (*slog.Logger, func() error, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}

	logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return logger, f.Close, nil
}
