package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type PrettyHandler struct {
	slog.Handler
}

var LogLevels = map[string]int{
	"DEBUG": Debug,
	"INFO":  Info,
	"WARN":  Warn,
	"ERROR": Error,
}

const (
	Debug int = -4
	Info  int = 0
	Warn  int = 4
	Error int = 8
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	timeStr := fmt.Sprintf("[%02d-%02d-%02d %02d:%02d:%02d]", r.Time.Year(), r.Time.Month(), r.Time.Day(), r.Time.Hour(), r.Time.Minute(), r.Time.Second())

	switch r.Level {
	case slog.LevelDebug:
		level = fmt.Sprintf(Green + level + Reset)
		fmt.Fprintf(os.Stdout, "%s %s %s\n", timeStr, level, r.Message)
	case slog.LevelInfo:
		level = fmt.Sprintf(Blue + level + Reset)
		fmt.Fprintf(os.Stdout, "%s %s %s\n", timeStr, level, r.Message)
	case slog.LevelWarn:
		level = fmt.Sprintf(Yellow + level + Reset)
		fmt.Fprintf(os.Stderr, "%s %s %s\n", timeStr, level, r.Message)
	case slog.LevelError:
		level = fmt.Sprintf(Red + level + Reset)
		fmt.Fprintf(os.Stderr, "%s %s %s\n", timeStr, level, r.Message)
	}

	return nil
}

func NewPrettyHandler(
	out io.Writer,
	opts slog.HandlerOptions,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewTextHandler(out, &opts),
	}

	return h
}

func SetLogLevel(logLevel string) string {
	upperLogLevel := strings.ToUpper(logLevel)
	level, found := LogLevels[upperLogLevel]
	if !found {
		upperLogLevel = "INFO"
		level = LogLevels[upperLogLevel]
	}

	opts := slog.HandlerOptions{
		Level: slog.Level(level),
	}

	handler := NewPrettyHandler(os.Stdout, opts)
	Log = slog.New(handler)
	return upperLogLevel
}

var Log *slog.Logger
