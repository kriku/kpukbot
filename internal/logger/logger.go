package logger

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	// Create a new logger with JSON output
	// You can customize the output format as needed
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return logger
}
