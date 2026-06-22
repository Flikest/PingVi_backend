package logger

import (
	"log/slog"
	"os"
)

const (
	envDebug = "local"
	envProd  = "prod"
)

func NewLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envDebug:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
