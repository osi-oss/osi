package logger

import (
	"log/slog"
	"os"
)

var (
	// Log глобальный логгер
	Log *slog.Logger
)

// Init инициализирует глобальный логгер
func Init(isDev bool) {
	var opts *slog.HandlerOptions

	if isDev {
		opts = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
	} else {
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// Error логирует ошибку с контекстом
func Error(msg string, err error, attrs ...any) {
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	Log.Error(msg, attrs...)
}

// Warn логирует предупреждение
func Warn(msg string, attrs ...any) {
	Log.Warn(msg, attrs...)
}

// Info логирует информацию
func Info(msg string, attrs ...any) {
	Log.Info(msg, attrs...)
}

// Debug логирует отладочную информацию
func Debug(msg string, attrs ...any) {
	Log.Debug(msg, attrs...)
}
