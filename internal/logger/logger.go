package logger

import (
	"io"
	"log/slog"
	"os"
)

func Init(filePath string, level slog.Level) (func(), error) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return func() {}, err
	}
	h := slog.NewJSONHandler(io.Writer(f), &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	slog.SetDefault(slog.New(h))
	return func() { _ = f.Close() }, nil
}
